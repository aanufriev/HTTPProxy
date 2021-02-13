package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func RunProxyServer() {
	port := ":8080"

	server := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handlerHTTPS(w, r)
				return
			}

			handlerHTTP(w, r)
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting server at %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handlerHTTP(w http.ResponseWriter, r *http.Request) {
	delete(r.Header, "Proxy-Connection")
	r.RequestURI = ""

	client := http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		fmt.Println("request err: ", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Println("transfer answer err: ", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func handlerHTTPS(w http.ResponseWriter, r *http.Request) {

}
