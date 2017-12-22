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
)

const keyPath = "key"
const idParam = "id"
const addressParam = "address"
const entitiesPath = "/entities"
const nodesPath = "/nodes"
const registerPath = "/register"
const visitedHeader="X-Corduroy-Visited"
const hopsHeader="X-Corduroy-Hops"

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
	node.service.Route(node.service.POST(entitiesPath + "/{" + keyPath + "}").To(node.putEntity))
	node.service.Route(node.service.GET(registerPath).To(node.registerNode))
	node.service.Route(node.service.GET(nodesPath).To(node.getNodes))
	restful.Add(node.service)
	return node
}

func (n *Node) Start(port int) {
	go func() {
		log.Fatal(n.server.ListenAndServe())
	}()

	n.nodes[n.ID] = n.Address
}

func (n *Node) Stop() {
	if n.server != nil {
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
	value := n.store.Get(key)
	bytes := []byte(value)
	response.Write(bytes)
}

func (n *Node) putEntity(request *restful.Request, response *restful.Response) {
	key := request.PathParameter(keyPath)
	entity := new(interface{})
	bytes, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	value := string(bytes)
	n.store.Put(key, value)
	response.WriteEntity(entity)
}

func (n *Node) registerNode(request *restful.Request, response *restful.Response) {
	rawId, err := url.QueryUnescape(request.QueryParameter(idParam))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.Atoi(rawId)
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
}

func (n *Node) getNodes(request *restful.Request, response *restful.Response) {
	response.WriteEntity(n.nodes)
}

func (n *Node) getEntityRemote(address string, key string, entity interface{}) error {
	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	response, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(entity)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) putEntityRemote(address string, key string, entity interface{}) error {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	err := encoder.Encode(entity)
	if err != nil {
		return err
	}

	uri := address + entitiesPath + "/" + url.QueryEscape(key)
	_, err = http.Post(uri, "application/json; charset=utf-8", b)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) registerNodeRemote(address string) error {
	uri := address + registerPath + "?" + idParam + "=" + strconv.Itoa(n.ID) + "&" + addressParam + "=" + n.Address
	_, err := http.Get(uri)
	return err
}

func (n *Node) syncNodesRemote(address string) error {
	uri := address + nodesPath
	response, err := http.Get(uri)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	nodes := &map[int]string{}
	err = decoder.Decode(nodes)
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