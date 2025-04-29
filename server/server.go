package server

import (
	"cloud/backend"
	"cloud/connect"
	"log"
	"sync/atomic"
)

type Servers struct {
	Current  uint64
	Backends []*backend.Backend
}

func (s *Servers) AddBackend(b *backend.Backend) {
	s.Backends = append(s.Backends, b)
}

func (s *Servers) NextIndex() int {
	log.Printf("NextIndex: Backend[%s] -> Backend[%d] ", s.Current, uint64(1)%uint64(len(s.Backends)))
	return int(atomic.AddUint64(&s.Current, uint64(1)) % uint64(len(s.Backends)))
}

func (s *Servers) GetNextBackend() *backend.Backend {
	next := s.NextIndex()
	length := len(s.Backends) + next //чтобы наврняка знать , чтобы не сдлелаем полный цикла, а не замкнемся на крайях
	for i := next; i < length; i++ {
		idx := i % len(s.Backends)
		log.Printf("GetNextBackend: idx = %d ", idx)
		if ok := s.Backends[idx].IsAlive(); ok {
			if idx != next {
				atomic.StoreUint64(&s.Current, uint64(idx))
			}
			return s.Backends[idx]
		}
	}
	return nil
}
func (s *Servers) StateCheck() {
	for _, b := range s.Backends {
		err := connect.Ping(b)
		if err != nil {
			b.SetAlive(false)
			log.Println(b.Addr.String()+"is not working because", err)
		} else {
			b.SetAlive(true)
			log.Println("StateCheck" + b.Addr.String() + "is working correct")
		}
	}
}
