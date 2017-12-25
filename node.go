package corduroy

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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
	server   *http.Server
	service  *restful.WebService
	store    Store
	registry Registry
	tickers  []*time.Ticker
}

func NewNode(port int, path string, store Store, registry Registry) *Node {
	address := buildLocalUri(port)
	node := &Node{
		Address: "http://" + buildLocalUri(port) + path,
		ID:      hash(address),
		server:  &http.Server{Addr: ":" + strconv.Itoa(port)},
		store:   store,
		registry:   registry,
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
	n.registry.Put(n.ID, n.Address)

	syncNodeTicker := time.NewTicker(time.Second * syncFrequencySeconds)
	go func() {
		for {
			<-syncNodeTicker.C
			n.syncRandomNodeRemote()
		}
	}()
	n.tickers = append(n.tickers, syncNodeTicker)

	syncValueTicker := time.NewTicker(time.Second * syncFrequencySeconds)
	go func() {
		for {
			<- syncValueTicker.C
			n.updateRandomValue()
		}
	}()
	n.tickers = append(n.tickers, syncValueTicker)

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

		n.registry.Delete(n.ID)
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
	log.Printf("node '%d' sending ping request to address '%s'", n.ID, uri)
	return send("GET", uri, "", []int{n.ID}, 1)
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

	visited, _ := parseVisited(&request.Request.Header)
	hops, _ := parseHops(&request.Request.Header)
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

	address := n.registry.Get(next)
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
	log.Printf("node '%d' sending get value request to address '%s'", n.ID, uri)
	return send("GET", uri, "", visited, hops)
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

	visited, _ := parseVisited(&request.Request.Header)
	hops, _ := parseHops(&request.Request.Header)
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

	address := n.registry.Get(next)
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
	log.Printf("node '%d' sending put value request to address '%s'", n.ID, uri)
	return send("PUT", uri, value, visited, hops)
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

	n.registry.Put(id, address)
	log.Printf("registered node '%d' with node '%d'", id, n.ID)
}

func (n *Node) registerNodeRemote(address string) error {
	uri := address + registerPath + "?" + idParam + "=" + strconv.Itoa(n.ID) + "&" + addressParam + "=" + n.Address
	log.Printf("node '%d' sending register request to address '%s'", n.ID, uri)
	_, _, err := send("PUT", uri, "", []int{n.ID}, 0)
	return err
}

func (n *Node) getNodes(request *restful.Request, response *restful.Response) {
	nodes := n.registry.GetAll()
	response.WriteEntity(nodes)
	log.Printf("provided '%d' nodes registered to node '%d'", len(nodes), n.ID)
}

func (n *Node) updateRandomValue() {
	if n.store.Size() == 0 {
		return
	}

	key := n.store.GetRandomKey()
	matches := n.bestMatches(key, 3, []int{})
	best := false
	for _, m := range matches {
		if m == n.ID {
			best = true
		}
	}

	match := n.bestMatch(key, []int{n.ID})
	address := n.registry.Get(match)
	n.putValueRemote(address, key, n.store.Get(key), []int{n.ID}, redundantCopies)

	if !best {
		n.store.Delete(key)
	}
}

func (n *Node) syncRandomNodeRemote() {
	if n.registry.Size() == 0 {
		return
	}

	id := n.registry.GetRandomID()
	n.syncNodeRemote(id)
}

func (n *Node) syncNodeRemote(id int) {
	address := n.registry.Get(id)
	uri := address + pingPath
	log.Printf("node '%d' sending sync request to address '%s'", n.ID, uri)
	statusCode, _, err := send("GET", uri, "", []int{n.ID}, 1)
	if err != nil || statusCode != http.StatusOK {
		n.registry.Delete(id)
		log.Printf("removed node '%d' from node '%d' registry", id, n.ID)
	} else {
		err = n.syncNodeRegistryRemote(address)
	}
}

func (n *Node) syncNodeRegistryRemote(address string) error {
	uri := address + nodesPath
	log.Printf("node '%d' sending sync registry request to address '%s'", n.ID, uri)
	_, body, err := send("GET", uri, "", []int{n.ID}, 0)
	if err != nil {
		return err
	}

	nodes := &map[int]string{}
	err = json.Unmarshal([]byte(body), nodes)
	if err != nil {
		return err
	}

	for id, address := range *nodes {
		n.registry.Put(id, address)
	}
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

	nodes := n.registry.GetAll()
	if len(nodes) == 0 {
		return -1
	}

	for id, _ := range nodes {
		if _, ok := x[id]; !ok {
			keys = append(keys, id)
		}
	}

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
