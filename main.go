package main

import (
	"cloud/backend"
	"cloud/config"
	"cloud/connect"
	"cloud/launch"
	"cloud/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

const MAXATTEMPTS = 5

var server models.Servers

func StateCheck() {
	t := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			server.StateCheck()
			log.Println("Health check completed")
		}
	}

}
func lb(w http.ResponseWriter, r *http.Request) {

	attempts := connect.GetAttemptsFromContext(r)
	if attempts > MAXATTEMPTS {
		log.Println("max attempts are reached", r.RemoteAddr)
		http.Error(w, "Backend is dead", http.StatusServiceUnavailable)
		return
	}
	backend := server.GetNextBackend()
	if backend != nil {

		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {

	cfg := config.NewConfig()
	fmt.Println(cfg)
	go launch.Launch()
	StringAddrs := cfg.BackendAddresses
	urls, err := connect.FormatStringToURL(StringAddrs)
	if err != nil {
		log.Fatal(err)
	}
	for _, url := range urls {
		proxy := httputil.NewSingleHostReverseProxy(url)
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
			backend.SetAlive(false)
		}
		server.AddBackend(&backend)
		log.Printf("Configured server: %s\n", url)

	}
	httpserver := http.Server{
		Addr:    ":" + cfg.Main,
		Handler: http.HandlerFunc(lb),
	}
	go StateCheck()
	err = httpserver.ListenAndServe()
	if err != nil {
		log.Println(err)
	}

}
