package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/flynn/flynn-discovery/Godeps/_workspace/src/github.com/flynn/flynn/pkg/httphelper"
	"github.com/flynn/flynn-discovery/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
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
		CreatorIP:        sourceIP(req),
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
	inst.CreatorIP = sourceIP(req)
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

func sourceIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[len(ips)-1])
	}
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}
