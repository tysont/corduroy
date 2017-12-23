package corduroy

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"io/ioutil"
	"strconv"
	"log"
	"time"
	"context"
	"net/url"
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

const keyPath = "key"
const idParam = "id"
const addressParam = "address"
const entitiesPath = "/entities"
const nodesPath = "/nodes"
const registerPath = "/register"
const visitedHeader = "X-Corduroy-Visited"
const hopsHeader = "X-Corduroy-Hops"
const defaultHops = 3

type Node struct {
	Address    string
	ID         int
	client  *http.Client
	server  *http.Server
	service *restful.WebService
	store   Store
	nodes   map[int]string
}

func NewNode(port int, path string, store Store) *Node {
	address := buildLocalUri(port)
	node := &Node{
		Address: "http://" + buildLocalUri(port) + path,
		ID: hash(address),
		client : &http.Client{},
		server: &http.Server{Addr:":" + strconv.Itoa(port)},
		store: store,
		nodes: make(map[int]string),
	}

	node.service = new(restful.WebService)
	node.service.Path(path).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	node.service.Route(node.service.GET(entitiesPath + "/{" + keyPath + "}").To(node.getEntity))
	node.service.Route(node.service.PUT(entitiesPath + "/{" + keyPath + "}").To(node.putEntity))
	node.service.Route(node.service.PUT(registerPath).To(node.registerNode))
	node.service.Route(node.service.GET(nodesPath).To(node.getNodes))
	restful.Add(node.service)
	return node
}

func (n *Node) Start(port int) {
	go func() {
		log.Printf("starting node '%d' at address '%s'", n.ID, n.Address)
		log.Fatal(n.server.ListenAndServe())
	}()

	n.nodes[n.ID] = n.Address
}

func (n *Node) Stop() {
	if n.server != nil {
		log.Printf("stopping node '%d'", n.ID)
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond * 100)
		log.Fatal(n.server.Shutdown(ctx))
	}
}

func (n *Node) getEntity(request *restful.Request, response *restful.Response) {
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

	address := n.nodes[next]
	statusCode, body, err := n.getEntityRemote(address, key, visited, hops)
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

func (n *Node) putEntity(request *restful.Request, response *restful.Response) {
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

	address := n.nodes[next]
	statusCode, body, err := n.putEntityRemote(address, key, value, visited, hops)
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
	n.nodes[id] = address
	log.Printf("registered node '%d' with node '%d'", id, n.ID)
}

func (n *Node) getNodes(request *restful.Request, response *restful.Response) {
	response.WriteEntity(n.nodes)
	log.Printf("provided '%d' nodes registered to node '%d'", len(n.nodes), n.ID)
}

func (n *Node) getEntityRemote(address string, key string, visited []int, hops int) (int, string, error) {
	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	return n.send("GET", uri, "", visited, hops)
}

func (n *Node) putEntityRemote(address string, key string, value string, visited []int, hops int) (int, string, error) {
	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	return n.send("PUT", uri, value, visited, hops)
}

func (n *Node) registerNodeRemote(address string) error {
	uri := address + registerPath + "?" + idParam + "=" + strconv.Itoa(n.ID) + "&" + addressParam + "=" + n.Address
	_, _, err := n.send("PUT", uri, "", []int{n.ID}, 0)
	return err
}

func (n *Node) syncNodesRemote(address string) error {
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

	for id, address := range *nodes {
		n.nodes[id] = address
	}
	return nil
}

func (n *Node) bestMatch(s string, excludes []int) int {
	if len(n.nodes) == 0 {
		return -1
	}

	keys := make([]int, 0)
	x := make(map[int]bool, len(excludes))
	for _, exclude := range excludes {
		x[exclude] = true
	}

	for id, _ := range n.nodes {
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
	defer response.Body.Close()
	if err != nil {
		return 0, "", err
	}

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

