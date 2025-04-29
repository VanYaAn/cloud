package launch

import (
	"cloud/config"
	"cloud/connect"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Launch() {

	cfg := config.NewConfig()
	fmt.Println(cfg)
	StringAddrs := cfg.BackendAddresses
	urls, err := connect.FormatStringToURL(StringAddrs)
	if err != nil {
		log.Fatal(err)
	}
	for _, url := range urls {
		go func() {
			http.ListenAndServe(url.Host, nil)
		}()
	}
	time.Sleep(10 * time.Minute)
}
