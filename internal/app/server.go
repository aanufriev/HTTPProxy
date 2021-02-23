package app

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func RunProxyServer() {
	port := ":8080"

	server := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			delete(r.Header, "Proxy-Connection")
			r.RequestURI = ""

			if r.Method == http.MethodConnect {
				handlerHTTPS(w, r)
				return
			}

			handlerHTTP(w, r)
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

func handlerHTTP(w http.ResponseWriter, r *http.Request) {
	client := http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		log.Printf("request err: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("transfer answer err: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func handlerHTTPS(w http.ResponseWriter, r *http.Request) {
	serverConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("couldn't connect to server: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		serverConn.Close()
		log.Printf("couldn't convert RepsonceWriter to Hijacker: %v", err)
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		serverConn.Close()
		log.Printf("couldn't Hijack: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	if _, err = clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		serverConn.Close()
		clientConn.Close()
		log.Printf("couldn't write 200: %v", err)
		return
	}

	go transfer(serverConn, clientConn)
	go transfer(clientConn, serverConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	_, err := io.Copy(destination, source)
	if err != nil {
		return
	}
}
