package app

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	proxyDelivery "github.com/aanufriev/httpproxy/internal/pkg/proxy/delivery"
)

func RunProxyServer() {
	port := ":8080"

	proxyHandler := proxyDelivery.Handler{}

	server := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r)
			delete(r.Header, "Proxy-Connection")
			r.RequestURI = ""

			if r.Method == http.MethodConnect {
				proxyHandler.HandleHTTPS(w, r)
				return
			}

			proxyHandler.HandleHTTP(w, r)
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,

		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("starting server at %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
