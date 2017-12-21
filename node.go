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
)

const keyPath = "key"
const idParam = "id"
const addressParam = "address"
const entitiesPath = "/entities"
const nodesPath = "/nodes"
const registerPath = "/register"

type Node struct {
	Address    string
	RootPath   string
	ID         int

	server  *http.Server
	service *restful.WebService
	store   Store
	nodes   map[int]string
}

func NewNode(store Store) Node {
	return Node{
		RootPath: "/v1",
		store: store,
		nodes: make(map[int]string),
	}
}

func (n *Node) Start(port int) {
	n.Address = buildLocalUri(port)
	n.ID = hash(n.Address)

	n.service = new(restful.WebService)
	n.service.Path(n.RootPath).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	n.service.Route(n.service.GET(entitiesPath + "/{" + keyPath + "}").To(n.getEntity))
	n.service.Route(n.service.POST(entitiesPath + "/{" + keyPath + "}").To(n.putEntity))
	n.service.Route(n.service.GET(registerPath).To(n.registerNode))

	go func() {
		restful.Add(n.service)
		n.server = &http.Server{Addr:":" + strconv.Itoa(port)}
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
	key := request.PathParameter(keyPath)
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
	}
	value := string(bytes)
	n.store.Put(key, value)
	response.WriteEntity(entity)
}

func (n *Node) registerNode(request *restful.Request, response *restful.Response) {
	rawId, err := url.QueryUnescape(request.QueryParameter(idParam))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	id, err := strconv.Atoi(rawId)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	address, err := url.QueryUnescape(request.QueryParameter(addressParam))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	n.nodes[id] = address
}
