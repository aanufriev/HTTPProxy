package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func RunProxyServer() {
	router := http.NewServeMux()

	router.HandleFunc("/", proxyHandler)

	port := ":8080"

	server := http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting server at %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read bytes err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respBody)
	if err != nil {
		fmt.Println("write response err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
