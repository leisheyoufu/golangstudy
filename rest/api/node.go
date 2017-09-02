package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/leisheyoufu/golangstudy/rest/utils"
)

const (
	configFile = "/tmp/config.json"
)

var (
	config *Config
)

type NodeApi struct {
	Router *mux.Router
	wgMu   sync.RWMutex
	routes Routes
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type Node struct {
	Name   string            `json:"name"`
	Driver string            `json:"driver"` // node type cmd, ssh, ipmitool
	Params map[string]string `json:params`
	Status string
}

type Config struct {
	nodes   []Node
	nodeMap map[string]int
}

func refreshNodeMap() {
	if config == nil {
		return
	}
	config.nodeMap = make(map[string]int)
	for i, v := range config.nodes {
		config.nodeMap[v.Name] = i
	}
}

func NewConfig() *Config {
	var err error
	config = new(Config)
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(bytes, &config.nodes); err != nil {
		panic(err)
	}
	refreshNodeMap()
	return config
}

func NewNodeApi() *NodeApi {
	router := mux.NewRouter().StrictSlash(true)
	api := NodeApi{Router: router}
	routes := Routes{
		Route{"Node", "GET", "/nodes", api.list},
		Route{"Node", "POST", "/nodes", api.post},
		Route{"Node", "GET", "/nodes/{node}", api.show},
		Route{"Node", "DELETE", "/nodes/{node}", api.delete},
	}
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	NewConfig()
	return &api
}

//func (api *NodeApi) hendler(w http.ResponseWriter, req *http.Request) {
//	info(req)
//	switch req.Method {
//		case "GET":
//			api.list(w, req)
//		case "POST":
//			api.post(w, req)
//		case "PUT":
//		case "DELETE":
//		default:
//	}
//}

func (api *NodeApi) list(w http.ResponseWriter, req *http.Request) {
	var resp []byte
	var err error
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	nodes := make(map[string][]string)
	for _, node := range config.nodes {
		nodes["nodes"] = append(nodes["nodes"], node.Name)
	}
	if resp, err = json.Marshal(nodes); err != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func (api *NodeApi) show(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var resp []byte
	var err error
	var index int
	if index, err = api.index(vars["node"]); err != nil {
		handle(w, req, http.StatusBadRequest, err)
		return
	}
	node := config.nodes[index]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if resp, err = json.Marshal(node); err != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", resp)
}

func (api *NodeApi) post(w http.ResponseWriter, req *http.Request) {
	var node Node
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return
	}
	if err := req.Body.Close(); err != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return
	}
	if err := json.Unmarshal(body, &node); err != nil {
		handle(w, req, http.StatusUnprocessableEntity, err)
		return
	}

	if api.exists(node) {
		err := errors.New("Already exist")
		handle(w, req, http.StatusConflict, err)
		return
	}
	node.Status = "Enroll"
	config.nodes = append(config.nodes, node)
	refreshNodeMap()
	api.save(w, req)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
}

func (api *NodeApi) delete(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var err error
	var index int
	if index, err = api.index(vars["node"]); err != nil {
		handle(w, req, http.StatusBadRequest, err)
		return
	}
	// delete(config.nodeMap, vars["node"])
	config.nodes = append(config.nodes[:index], config.nodes[index+1:]...)
	refreshNodeMap()
	if api.save(w, req) != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (api *NodeApi) exists(node Node) bool {
	if _, ok := config.nodeMap[node.Name]; ok {
		return true
	}
	return false
}

func (api *NodeApi) index(name string) (int, error) {
	var index int
	var ok bool
	if index, ok = config.nodeMap[name]; !ok {
		return -1, errors.New("Could not be found")
	}
	return index, nil
}

func (api *NodeApi) save(w http.ResponseWriter, req *http.Request) error {
	var data []byte
	var err error
	if data, err = json.Marshal(config.nodes); err != nil {
		handle(w, req, http.StatusInternalServerError, err)
		return err
	}
	api.wgMu.Lock()
	utils.WriteJsonFile(configFile, data)
	api.wgMu.Unlock()
	return nil
}
