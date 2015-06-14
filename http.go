package main

import (
	"fmt"
	"net/http"

	"github.com/flynn/flynn/pkg/httphelper"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	URL     string
	Backend StorageBackend
	router  *httprouter.Router
}

func NewServer(url string, backend StorageBackend) *Server {
	s := &Server{
		URL:     url,
		Backend: backend,
		router:  httprouter.New(),
	}
	s.router.POST("/clusters", s.CreateCluster)
	s.router.POST("/clusters/:cluster_id/instances", s.CreateInstance)
	s.router.GET("/clusters/:cluster_id/instances", s.GetInstances)

	return s
}

func (s *Server) CreateCluster(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cluster := &Cluster{
		CreatorIP:        req.RemoteAddr, // TODO: parse X-Forwarded-For
		CreatorUserAgent: req.Header.Get("User-Agent"),
	}

	if len(cluster.CreatorUserAgent) > 1000 {
		cluster.CreatorUserAgent = cluster.CreatorUserAgent[:1000]
	}

	if err := s.Backend.CreateCluster(cluster); err != nil {
		httphelper.Error(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s/clusters/%s", s.URL, cluster.ID))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) CreateInstance(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var data struct {
		Data *Instance `json:"data"`
	}
	if err := httphelper.DecodeJSON(req, &data); err != nil {
		httphelper.Error(w, err)
		return
	}
	inst := data.Data
	inst.ClusterID = params.ByName("cluster_id")
	inst.CreatorIP = req.RemoteAddr // TODO: parse X-Forwarded-For
	// TODO: validate with JSON schema

	status := http.StatusCreated
	if err := s.Backend.CreateInstance(inst); err == ErrExists {
		status = http.StatusConflict
	} else if err != nil {
		httphelper.Error(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s/clusters/%s/instances/%s", s.URL, inst.ClusterID, inst.ID))
	httphelper.JSON(w, status, data)
}

func (s *Server) GetInstances(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	instances, err := s.Backend.GetClusterInstances(params.ByName("cluster_id"))
	if err != nil {
		httphelper.Error(w, err)
		return
	}
	if instances == nil {
		instances = []*Instance{}
	}
	httphelper.JSON(w, 200, struct {
		Data []*Instance `json:"data"`
	}{instances})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
