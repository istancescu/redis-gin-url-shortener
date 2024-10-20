package alb

import (
	"net/http/httputil"
	"sync"
	"sync/atomic"
)

type LoadBalancer struct {
	algorithm any
	current   any
}

type Status struct {
	isAlive bool
}

type ServerConfiguration struct {
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

func CreateLoadBalancer(algorithm any) *LoadBalancer {
	return &LoadBalancer{algorithm: algorithm}
}

func CreateServerPool() *ServerPool {
	return &ServerPool{}
}

func CreateAppServer() *AppServer {
	return &AppServer{}

}

func ProvideConfigurationForServer() *ServerConfiguration {
	return &ServerConfiguration{}
}

func Create(s *ServerPool) *LoadBalancer {

	//ServerPool{servers: make([]*AppServer, 0), current: nil}
	return nil
}

func AddServer(s *ServerPool, server *AppServer) {
	s.servers = append(s.servers, server)

}

func (a *AppServer) SetAlive(alive bool) {
	a.mux.Lock()
	a.status.isAlive = alive
	a.mux.Unlock()
}

func (a *AppServer) GetAlive() bool {
	a.mux.RLock()
	alive := a.status.isAlive
	a.mux.RUnlock()
	return alive
}

func (sp *ServerPool) NextIndex() int {
	return int(atomic.AddUint32(&sp.current, 1) % uint32(len(sp.servers)))
}

func (sp *ServerPool) Next() *AppServer {
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
