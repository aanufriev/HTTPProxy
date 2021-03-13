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

	server := http.Server{
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

	log.Printf("starting server at %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
