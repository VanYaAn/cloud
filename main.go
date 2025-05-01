package main

import (
	"cloud/backend"
	"cloud/bucket"
	"cloud/config"
	"cloud/connect"
	"cloud/launch"
	"cloud/models"
	"cloud/repo"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

const MAXATTEMPTS = 5

var server models.Servers
var ratelimiter bucket.RateLimiter

func StateCheck() {
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			server.StateCheck()
			log.Println("Health check completed")
		}
	}
}

func firstStateCheck() {

	log.Println("Starting health check...")
	server.StateCheck()
	log.Println("Health check completed")
}

func lb(w http.ResponseWriter, r *http.Request) {
	log.Println("lb: request received") // Добавлено

	attempts := connect.GetAttemptsFromContext(r)

	if attempts > MAXATTEMPTS {
		log.Println("max attempts are reached", r.RemoteAddr)
		http.Error(w, "Backend is dead", http.StatusServiceUnavailable)
		return
	}
	backend := server.GetNextBackend()
	if backend != nil {
		log.Printf("lb: forwarding request to %s\n", backend.Addr.String())
		backend.ReverseProxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}
func GetClientID(r *http.Request) int64 {
	body, err := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		r.Body.Close()
		return 0
	}
	defer r.Body.Close()

	var data bucket.Client
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err, data)

		r.Body.Close()
		return 0
	}
	return data.ID
}

func lbrl(w http.ResponseWriter, r *http.Request) {

	clientID := GetClientID(r)

	if clientID == 0 {
		log.Println("not enought parametrs in request: dont have id ", r.RemoteAddr)
		http.Error(w, "No id", http.StatusNoContent)
		return
	}
	if _, ok := ratelimiter.Buckets[clientID]; !ok {
		ratelimiter.Buckets[clientID] = &bucket.Bucket{Capacity: bucket.DEFAULT_CAPACITY, Rate: bucket.DEFAULT_RATE, Tokens: bucket.DEFAULT_CAPACITY}
		//добавлять в бд можно, но лучше не здесб
		//ratelimiter.Clients.AddCLient(bucket.Client{ID: clientID,Rate: bucket.DEFAULT_RATE,Capacity: bucket.DEFAULT_CAPACITY})
		log.Println("create new  id ", clientID)
	}
	if ratelimiter.Buckets[clientID].GetToken() {
		lb(w, r)
		return
	}
	http.Error(w, "too many requests", http.StatusTooManyRequests)
}

const (
	host     = "localhost" // Или IP-адрес, если вы не на localhost
	port     = 5432        // Порт PostgreSQL (стандартный, если не меняли)
	user     = "your_user"
	password = "your_password"
	dbname   = "your_database"
)

func main() {

	cfg := config.NewConfig()
	db := repo.New(strconv.Itoa(port), user, host, dbname, password)
	ratelimiter = bucket.RateLimiter{Clients: db, Buckets: make(map[int64]*bucket.Bucket)}
	fmt.Println(cfg)
	StringAddrs := cfg.BackendAddresses
	urls, err := connect.FormatStringToURL(StringAddrs)
	if err != nil {
		log.Fatal(err)
	}

	for _, url := range urls {
		tmpurl, _ := url.Parse(url.String())
		proxy := httputil.NewSingleHostReverseProxy(tmpurl)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			retry := connect.GetRetryFromContext(r)
			if retry < 5 {
				select {
				case <-time.After(10 * time.Millisecond):

					ctx := context.WithValue(r.Context(), connect.Retry, retry+1)
					proxy.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}
			for _, b := range server.Backends {
				if b.Addr == url {
					b.SetAlive(false)
				}
			}
			attempts := connect.GetAttemptsFromContext(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), connect.Attempts, attempts+1)
			lb(w, r.WithContext(ctx))

		}

		backend := backend.Backend{
			Addr:         url,
			Alive:        true,
			ReverseProxy: proxy,
		}

		if err = connect.Ping(&backend); err != nil {
			log.Println("main: setting Alive = false ")
			backend.SetAlive(false)
		}
		server.AddBackend(&backend)
		log.Printf("Configured server: %s\n", url)

	}

	httpserver := http.Server{
		Addr:    ":" + cfg.Main,
		Handler: http.HandlerFunc(lbrl),
	}
	go firstStateCheck()
	go launch.Launch()
	time.NewTicker(5 * time.Millisecond)
	go StateCheck()
	err = httpserver.ListenAndServe()
	if err != nil {
		log.Println(err)
	}

}
