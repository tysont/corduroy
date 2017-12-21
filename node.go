package corduroy

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"io/ioutil"
	"strconv"
	"log"
	"time"
	"context"
)

const keyPath string = "key"

type Node struct {
	Port       int
	Address    string
	RootPath   string

	server  *http.Server
	service *restful.WebService
	store   Store
}

func NewNode(store Store) Node {
	return Node{
		RootPath: "/v1/entities",
		store: store,
	}
}

func (n *Node) Start(port int) {
	n.Port = port
	n.Address = buildLocalUri(port)
	n.service = new(restful.WebService)

	n.service.Path(n.RootPath).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	n.service.Route(n.service.GET("/{" + keyPath + "}").To(n.Get))
	n.service.Route(n.service.POST("/{" + keyPath + "}").To(n.Put))

	go func() {
		restful.Add(n.service)
		n.server = &http.Server{Addr:":" + strconv.Itoa(n.Port)}
		log.Fatal(n.server.ListenAndServe())
	}()
}

func (n *Node) Stop() {
	if n.server != nil {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond * 100)
		log.Fatal(n.server.Shutdown(ctx))
	}
}

func (n *Node) Get(request *restful.Request, response *restful.Response) {
	key := request.PathParameter(keyPath)
	value := n.store.Get(key)
	bytes := []byte(value)
	response.Write(bytes)
}

func (n *Node) Put(request *restful.Request, response *restful.Response) {
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