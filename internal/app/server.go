package app

import (
	"crypto/tls"
	"database/sql"
	"log"
	"net/http"
	"time"

	proxyDelivery "github.com/aanufriev/httpproxy/internal/pkg/proxy/delivery"
	proxyRepository "github.com/aanufriev/httpproxy/internal/pkg/proxy/repository"
	ProxyUsecase "github.com/aanufriev/httpproxy/internal/pkg/proxy/usecase"
	repeaterDelivery "github.com/aanufriev/httpproxy/internal/pkg/repeater/delivery"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

func RunProxyServer() {
	port := ":8080"

	db, err := sql.Open("postgres", "host=localhost dbname=requests sslmode=disable")
	if err != nil {
		log.Printf("postgres not available: %v", err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Printf("no connection with db: %v", err)
		return
	}

	proxyRepository := proxyRepository.NewProxyRepository(db)
	proxyUsecase := ProxyUsecase.NewProxyUsecase(proxyRepository)
	proxyHandler := proxyDelivery.NewProxyHandler(proxyUsecase)

	proxyServer := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	go func() {
		log.Printf("starting proxy server at %s", port)
		log.Fatal(proxyServer.ListenAndServe())
	}()

	repeatHandler := repeaterDelivery.NewRepeaterHandler(proxyUsecase)

	mux := mux.NewRouter()

	mux.HandleFunc("/requests", repeatHandler.ShowAllRequests)
	mux.HandleFunc("/request/{id}", repeatHandler.ShowRequest)
	mux.HandleFunc("/repeat/{id}", repeatHandler.RepeatRequest)

	repeaterPort := ":8000"
	log.Printf("starting repeater at %s", repeaterPort)
	log.Fatal(http.ListenAndServe(repeaterPort, mux))
}
