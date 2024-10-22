package alb

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// TODO reimplement this using Linked List
type LoadBalancer struct {
	serverPool *ServerPool
}

type Status struct {
	isAlive bool
}

type ServerConfiguration struct {
	Timeout uint8
}

type AppServer struct {
	serverConfiguration ServerConfiguration
	status              Status
	ReverseProxy        *httputil.ReverseProxy
	mux                 *sync.RWMutex
}

type ServerPool struct {
	servers []*AppServer
	current uint32
}

func CreateLoadBalancer(pool *ServerPool) *LoadBalancer {
	return &LoadBalancer{pool}
}

func CreateServerPool() *ServerPool {
	return &ServerPool{}
}

func CreateAppServer(configuration ServerConfiguration, url2 *url.URL) *AppServer {
	return &AppServer{serverConfiguration: configuration,
		status:       Status{isAlive: false},
		ReverseProxy: httputil.NewSingleHostReverseProxy(url2),
		mux:          &sync.RWMutex{},
	}

}

func AddServer(s *ServerPool, a *AppServer) {
	a.mux.Lock()

	defer a.mux.Unlock()
	s.servers = append(s.servers, a)
}

func (a *AppServer) SetAlive(alive bool) {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.status.isAlive = alive
}

func (a *AppServer) GetAlive() bool {
	a.mux.RLock()
	defer a.mux.RUnlock()

	alive := a.status.isAlive

	return alive
}

func (sp *ServerPool) NextIndex() int {
	return int(atomic.AddUint32(&sp.current, 1) % uint32(len(sp.servers)))
}

func (sp *ServerPool) NextPeer() *AppServer {
	if len(sp.servers) == 0 {
		return nil
	}

	next := sp.NextIndex()

	for i := 0; i < len(sp.servers); i++ {
		idx := (next + i) % len(sp.servers)
		if sp.servers[idx].GetAlive() {
			if i != next {
				atomic.StoreUint32(&sp.current, uint32(idx))
			}
			return sp.servers[idx]
		}
	}
	return nil
}

func (sp *ServerPool) HandleHTTPRequests(gin *gin.Context) {
	log.Printf("HEY HEY HEY I AM HANDLING THE HTTP")
	peer := sp.NextPeer()

	if peer == nil {
		log.Printf("no peer found in this context \n")
		return
	}
	if peer.GetAlive() {
		urlToShorten := gin.Param("urlToShorten") // This extracts `google.com` from the path
		gin.Request.URL.Path = "/url/" + urlToShorten
		log.Printf("current server: %d", sp.current)
		peer.ReverseProxy.ServeHTTP(gin.Writer, gin.Request)
	}
	return
}
