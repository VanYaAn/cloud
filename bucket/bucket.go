package bucket

import (
	"log"
	"sync"
	"time"
)

const DEFAULT_CAPACITY = 100
const DEFAULT_RATE = 10

type Bucket struct {
	Capacity int64
	Tokens   int64 //Tokens at the moment
	Rate     int64 // Add a token to the bucket every 1/r units of time
	mux      sync.Mutex
}
type Client struct {
	ID       int64 `json:"id"`
	Capacity int64 `json:"capacity"`
	Rate     int64 `json:"rate"`
}

type RateLimiter struct {
	Buckets     map[int64]*Bucket
	mux         sync.RWMutex
	Clients     RepositoryInterface
	TokenAmount int64 //сколько токенов добавляется каждый TimeAmount
	TimeAmount  int64 // как часто мы добавляем в каждый бакет токены
}
type RepositoryInterface interface {
	GetID(string) (*Client, error)
	GetALL() ([]*Client, error)
	AddCLient(*Client) error
}

func (rl *RateLimiter) SetClients(something RepositoryInterface) {
	rl.mux.Lock()
	rl.Clients = something
	rl.mux.Unlock()
	return
}
func (rl *RateLimiter) AddToAllBuckets() {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-t.C:
			for _, bucket := range rl.Buckets {
				bucket.mux.Lock()
				if bucket.Tokens+bucket.Rate >= bucket.Capacity {
					bucket.Tokens = bucket.Capacity
				} else {
					bucket.Tokens = bucket.Tokens + bucket.Rate
				}
				bucket.mux.Unlock()
			}
		}
	}
}

func (rl *RateLimiter) GetID(id string) (*Client, error) {
	rl.mux.Lock()

	client, err := rl.Clients.GetID(id)
	if err != nil {
		log.Fatalf("RateLimiter GetID: %v", err)
		return &Client{}, err
	} //если этот клиент уже есть в базе, то просто вернем его бакет
	//если этого клиента нету , то этот клиент создается в базе с дефолт значениями
	return client, nil
}
func (rl *RateLimiter) GetAllID() {
	rl.mux.Lock()
	allclients, err := rl.Clients.GetALL()
	if err != nil {
		log.Fatalf("RateLimiter GetALL: %v", err)
	}
	for _, client := range allclients {
		rl.Buckets[client.ID] = &Bucket{Capacity: client.Capacity, Rate: client.Rate}
	}
	rl.mux.Unlock()
	return
}

func (b *Bucket) GetToken() bool {
	b.mux.Lock()
	if (b.Tokens) <= 0 {
		return false
	} else {
		b.Tokens--
	}
	b.mux.Unlock()
	return true
}

// func handlerClient(w http.ResponseWriter, r *http.Request) {
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Error reading request body", http.StatusBadRequest)
// 		log.Printf("Error reading request body: %v", err)
// 		return
// 	}
// 	defer r.Body.Close()

// 	var data Client
// 	err = json.Unmarshal(body, &data)
// 	if err != nil {
// 		http.Error(w, "Error unmarshaling JSON", http.StatusBadRequest)
// 		log.Printf("Error unmarshaling JSON: %v", err)
// 		return
// 	}
// 	fmt.Fprintf(w, "id =  %s capacity = %x rate = %s", data.ID, data.Capacity, data.Rate)
// }
