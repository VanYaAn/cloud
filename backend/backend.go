package backend

import (
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	Addr         *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) IsAlive() bool {
	b.mux.Lock()
	a := b.Alive
	b.mux.Unlock()
	return a
}

func (b *Backend) SetAlive(state bool) {
	b.mux.Lock()
	b.Alive = state
	b.mux.Unlock()
}

func Ping(b *Backend) error {
	conn, err := net.DialTimeout("tcp", b.Addr.String(), 2*time.Millisecond)
	if err != nil {
		log.Println("Ping: Backend"+b.Addr.Host+"is not working for reason", err)
		return err
	}
	defer conn.Close()
	return nil
}
