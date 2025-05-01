package launch

import (
	"cloud/config"
	"cloud/connect"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

// func Launch() {

// 	cfg := config.NewConfig()
// 	log.Println("Launch: launching servers ")
// 	StringAddrs := cfg.BackendAddresses
// 	urls, err := connect.FormatStringToURL(StringAddrs)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, url := range urls {
// 		go func() {
// 			http.ListenAndServe(url.Host, nil)
// 		}()
// 	}
// 	time.Sleep(10 * time.Minute)
// }

// func Launch() {
// 	cfg := config.NewConfig()
// 	fmt.Println(cfg)
// 	StringAddrs := cfg.BackendAddresses
// 	urls, err := connect.FormatStringToURL(StringAddrs)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var wg sync.WaitGroup
// 	for _, url1 := range urls {
// 		wg.Add(1)
// 		go func(url *url.URL) {
// 			defer wg.Done()
// 			log.Printf("Starting server on %s\n", url.Host)
// 			err := http.ListenAndServe(url.Host, nil) // Исправлено: используем url.Host
// 			if err != nil {
// 				log.Printf("Error starting server on %s: %v\n", url.Host, err)
// 			}
// 		}(url1)
// 	}
// 	wg.Wait()
// 	log.Println("All servers started.")

// }

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
	log.Println("defaultHandler:" + r.Host)
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}
func Launch() {
	cfg := config.NewConfig()
	fmt.Println(cfg)
	StringAddrs := cfg.BackendAddresses
	urls, err := connect.FormatStringToURL(StringAddrs)
	if err != nil {
		log.Fatalf("Launch: error formatting URLs: %v", err)
		return
	}

	var wg sync.WaitGroup

	for _, urlValue := range urls {
		wg.Add(1)
		go func(urlValue *url.URL) {
			defer wg.Done()
			addr := urlValue.Host // Получаем адрес и порт
			mux := http.NewServeMux()

			log.Printf("Starting server on %s\n", addr)
			mux.HandleFunc("/", defaultHandler)
			mux.HandleFunc("/health", healthHandler)

			err := http.ListenAndServe(addr, mux) // Используем addr
			if err != nil {
				log.Printf("Error starting server on %s: %v\n", addr, err)
			}
		}(urlValue)
	}
	wg.Wait()
	log.Println("Launch: all servers started (or failed)")
}
