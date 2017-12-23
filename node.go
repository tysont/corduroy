package corduroy

import (
	"bytes"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const redundantCopies = 3
const syncFrequencySeconds = 20

const keyPath = "key"
const idParam = "id"
const addressParam = "address"
const pingPath = "/ping"
const entitiesPath = "/entities"
const nodesPath = "/nodes"
const registerPath = "/register"
const visitedHeader = "X-Corduroy-Visited"
const hopsHeader = "X-Corduroy-Hops"

type Node struct {
	Address  string
	ID       int
	client   *http.Client
	server   *http.Server
	service  *restful.WebService
	store    Store
	nodes    map[int]string
	nodesMux sync.Mutex
	tickers  []*time.Ticker
}

func NewNode(port int, path string, store Store) *Node {
	address := buildLocalUri(port)
	node := &Node{
		Address: "http://" + buildLocalUri(port) + path,
		ID:      hash(address),
		client:  &http.Client{},
		server:  &http.Server{Addr: ":" + strconv.Itoa(port)},
		store:   store,
		nodes:   make(map[int]string),
	}

	node.service = new(restful.WebService)
	node.service.Path(path).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	node.service.Route(node.service.GET(pingPath).To(node.ping))
	node.service.Route(node.service.GET(entitiesPath + "/{" + keyPath + "}").To(node.getValue))
	node.service.Route(node.service.PUT(entitiesPath + "/{" + keyPath + "}").To(node.putValue))
	node.service.Route(node.service.PUT(registerPath).To(node.registerNode))
	node.service.Route(node.service.GET(nodesPath).To(node.getNodes))
	restful.Add(node.service)
	return node
}

func (n *Node) Start(port int) {
	go func() {
		log.Printf("starting server at node '%d' with address '%s'", n.ID, n.Address)
		err := n.server.ListenAndServe()
		if err != nil {
			log.Printf("server error at node '%d': '%s'", n.ID, err)
		}
	}()

	n.nodesMux.Lock()
	n.nodes[n.ID] = n.Address
	n.nodesMux.Unlock()

	syncTicker := time.NewTicker(time.Second * syncFrequencySeconds)
	go func() {
		for {
			<-syncTicker.C
			n.syncRandomNode()
		}
	}()
	n.tickers = append(n.tickers, syncTicker)

	time.Sleep(time.Millisecond * 10)
	n.waitStart()
}

func (n *Node) waitStart() {
	statusCode, _, err := n.pingRemote(n.Address)
	for down := true; down; down = statusCode != http.StatusOK || err != nil {
		time.Sleep(time.Millisecond * 20)
		statusCode, _, err = n.pingRemote(n.Address)
	}
}

func (n *Node) Stop() {
	if n.server != nil {
		log.Printf("stopping server at node '%d'", n.ID)
		go func() {
			err := n.server.Shutdown(nil)
			if err != nil {
				log.Printf("unable to stop server at node '%d': '%s'", n.ID, err)
			}
		}()

		for _, ticker := range n.tickers {
			ticker.Stop()
		}

		n.nodesMux.Lock()
		delete(n.nodes, n.ID)
		n.nodesMux.Unlock()

		time.Sleep(time.Millisecond * 10)
		n.waitStop()
	}
}

func (n *Node) waitStop() {
	statusCode, _, err := n.pingRemote(n.Address)
	for up := true; up; up = statusCode == http.StatusOK && err == nil {
		time.Sleep(time.Millisecond * 20)
		statusCode, _, err = n.pingRemote(n.Address)
	}
}

func (n *Node) ping(request *restful.Request, response *restful.Response) {
	response.WriteHeader(http.StatusOK)
	log.Printf("node '%d' responded to ping", n.ID)
}

func (n *Node) pingRemote(address string) (int, string, error) {
	uri := address + pingPath
	return n.send("GET", uri, "", []int{n.ID}, 1)
}

func (n *Node) getValue(request *restful.Request, response *restful.Response) {
	key, err := url.QueryUnescape(request.PathParameter(keyPath))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	if n.store.Contains(key) {
		value := n.store.Get(key)
		b := []byte(value)
		log.Printf("retrieved value for key '%s' from node '%d'", key, n.ID)
		response.WriteHeader(http.StatusOK)
		response.Write(b)
		return
	}

	visited, hops, err := n.parse(request.Request)
	if hops <= 0 {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	hops--
	visited = append(visited, n.ID)
	next := n.bestMatch(key, visited)
	if next < 0 {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	n.nodesMux.Lock()
	address := n.nodes[next]
	n.nodesMux.Unlock()

	statusCode, body, err := n.getValueRemote(address, key, visited, hops)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	if statusCode != http.StatusOK {
		response.WriteHeader(statusCode)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte(body))
}

func (n *Node) getValueRemote(address string, key string, visited []int, hops int) (int, string, error) {
	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	return n.send("GET", uri, "", visited, hops)
}

func (n *Node) putValue(request *restful.Request, response *restful.Response) {
	key := request.PathParameter(keyPath)
	bytes, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	value := string(bytes)
	n.store.Put(key, value)
	log.Printf("wrote key '%s' and associated value to node '%d'", key, n.ID)

	visited, hops, err := n.parse(request.Request)
	if hops <= 0 {
		response.WriteHeader(http.StatusOK)
		return
	}
	hops--
	visited = append(visited, n.ID)
	next := n.bestMatch(key, visited)
	if next < 0 {
		response.WriteHeader(http.StatusOK)
		return
	}

	n.nodesMux.Lock()
	address := n.nodes[next]
	n.nodesMux.Unlock()

	statusCode, body, err := n.putValueRemote(address, key, value, visited, hops)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	if statusCode != http.StatusOK {
		response.WriteHeader(statusCode)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte(body))
}

func (n *Node) putValueRemote(address string, key string, value string, visited []int, hops int) (int, string, error) {
	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	return n.send("PUT", uri, value, visited, hops)
}

func (n *Node) registerNode(request *restful.Request, response *restful.Response) {
	i, err := url.QueryUnescape(request.QueryParameter(idParam))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.Atoi(i)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	address, err := url.QueryUnescape(request.QueryParameter(addressParam))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	n.nodesMux.Lock()
	n.nodes[id] = address
	log.Printf("registered node '%d' with node '%d'", id, n.ID)
	n.nodesMux.Unlock()
}

func (n *Node) registerNodeRemote(address string) error {
	uri := address + registerPath + "?" + idParam + "=" + strconv.Itoa(n.ID) + "&" + addressParam + "=" + n.Address
	_, _, err := n.send("PUT", uri, "", []int{n.ID}, 0)
	return err
}

func (n *Node) getNodes(request *restful.Request, response *restful.Response) {
	n.nodesMux.Lock()
	response.WriteEntity(n.nodes)
	log.Printf("provided '%d' nodes registered to node '%d'", len(n.nodes), n.ID)
	n.nodesMux.Unlock()
}

func (n *Node) syncRandomNode() {
	n.nodesMux.Lock()
	r := rand.Int() % len(n.nodes)
	var id int
	i := 0
	for id = range n.nodes {
		if i == r {
			break
		}
		i++
	}
	n.nodesMux.Unlock()
	n.syncNode(id)
}

func (n *Node) syncNode(id int) {
	n.nodesMux.Lock()
	address := n.nodes[id]
	n.nodesMux.Unlock()

	uri := address + pingPath
	statusCode, _, err := n.send("GET", uri, "", []int{n.ID}, 1)
	if err != nil || statusCode != http.StatusOK {
		n.nodesMux.Lock()
		delete(n.nodes, id)
		log.Printf("removed node '%d' from node '%d' registry", id, n.ID)
		n.nodesMux.Unlock()
	} else {
		err = n.syncNodeRemote(address)
	}
}

func (n *Node) syncNodeRemote(address string) error {
	uri := address + nodesPath
	_, body, err := n.send("GET", uri, "", []int{n.ID}, 0)
	if err != nil {
		return err
	}

	nodes := &map[int]string{}
	err = json.Unmarshal([]byte(body), nodes)
	if err != nil {
		return err
	}

	n.nodesMux.Lock()
	for id, address := range *nodes {
		n.nodes[id] = address
	}
	n.nodesMux.Unlock()
	return nil
}

func (n *Node) bestMatches(s string, count int, excludes []int) []int {
	matches := make([]int, 0)
	for i := 0; i < count; i++ {
		match := n.bestMatch(s, excludes)
		if match < 0 {
			return matches
		}
		matches = append(matches, match)
		excludes = append(excludes, match)
	}
	return matches
}

func (n *Node) bestMatch(s string, excludes []int) int {
	keys := make([]int, 0)
	x := make(map[int]bool, len(excludes))
	for _, exclude := range excludes {
		x[exclude] = true
	}

	n.nodesMux.Lock()
	if len(n.nodes) == 0 {
		return -1
	}

	for id, _ := range n.nodes {
		if _, ok := x[id]; !ok {
			keys = append(keys, id)
		}
	}
	n.nodesMux.Unlock()

	if len(keys) == 0 {
		return -1
	}

	sort.Ints(keys)
	h := hash(s)
	var last int
	for i, key := range keys {
		if i == 0 {
			last = key
		} else {
			if key >= h {
				return last
			}
			last = key
		}
	}

	return last
}

func (n *Node) send(verb string, uri string, body string, visited []int, hops int) (int, string, error) {
	b1 := []byte(body)
	buff := bytes.NewBuffer(b1[:])
	request, err := http.NewRequest(verb, uri, buff)
	if err != nil {
		return 0, "", err
	}

	v := ""
	for _, id := range visited {
		if v != "" {
			v = v + ","
		}
		v = v + strconv.Itoa(id)
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set(visitedHeader, v)
	request.Header.Set(hopsHeader, strconv.Itoa(hops))
	log.Printf("node '%d' sending '%s' request to address '%s'", n.ID, verb, uri)
	response, err := n.client.Do(request)
	if err != nil {
		return 0, "", err
	}

	defer response.Body.Close()
	b2, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, "", err
	}

	return response.StatusCode, string(b2), nil
}

func (n *Node) parse(request *http.Request) ([]int, int, error) {
	v := request.Header.Get(visitedHeader)
	s := strings.Split(v, ",")
	visited := make([]int, len(s))
	for _, id := range s {
		n, err := strconv.Atoi(id)
		if err != nil {
			visited = append(visited, n)
		}
	}

	hops, err := strconv.Atoi(request.Header.Get(hopsHeader))
	if err != nil {
		return make([]int, 0), 0, err
	}
	return visited, hops, nil
}
