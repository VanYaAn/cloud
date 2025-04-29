package connect

import (
	"cloud/backend"

	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

const LOCALHOST = "http://localhost"
const MAXATTEMPTS = 5
const (
	Attempts = iota
	Retry
)

func Ping(b *backend.Backend) error {
	conn, err := net.DialTimeout("tcp", b.Addr.String(), 2*time.Millisecond)
	if err != nil {
		log.Println("Ping: Backend"+b.Addr.Host+"is not working for reason", err)
		return err
	}
	defer conn.Close()
	return nil
}

func FormatStringToURL(ss []string) ([]*url.URL, error) {
	res := make([]*url.URL, 0)
	for _, addr := range ss {
		tmp, err := url.Parse(LOCALHOST + addr)
		if err != nil {
			log.Fatalf("FormatStringToURL: fail parse env file %v", err)
		}
		res = append(res, tmp)
	}
	log.Println("FormatStringToURL: successful parsing env file")
	return res, nil
}

func GetAttemptsFromContext(r *http.Request) int {

	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		log.Printf("GetAttemptsFromContext: ok = true , attempts = %s", attempts)
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		log.Printf("GetRetryFromContext: ok = true , retry = %s", retry)
		return retry
	}
	return 0
}
