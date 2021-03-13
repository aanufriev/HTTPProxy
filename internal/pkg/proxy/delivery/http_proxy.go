package delivery

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/aanufriev/httpproxy/internal/pkg/models"
	"github.com/aanufriev/httpproxy/internal/pkg/proxy/interfaces"
	"github.com/aanufriev/httpproxy/pkg/cert"
)

type ProxyHandler struct {
	usecase interfaces.Usecase
}

func NewProxyHandler(usecase interfaces.Usecase) ProxyHandler {
	return ProxyHandler{
		usecase: usecase,
	}
}

func (h ProxyHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
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

	req, err := models.ConvertFromHttpRequest(*r)
	if err != nil {
		log.Printf("couldn't convert request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	err = h.usecase.SaveRequest(req)
	if err != nil {
		log.Printf("couldn't save request to db: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func (h ProxyHandler) HandleHTTPS(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		log.Printf("couldn't get host: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	certificate, err := cert.GetCertificate(host)
	if err != nil {
		log.Printf("couldn't get certifate: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	var serverConn *tls.Conn

	serverConfig := new(tls.Config)

	serverConfig.Certificates = []tls.Certificate{certificate}
	serverConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		clientConfig := new(tls.Config)
		clientConfig.ServerName = hello.ServerName
		serverConn, err = tls.Dial("tcp", r.Host, clientConfig)
		if err != nil {
			log.Printf("dial tcp error: %v", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return nil, err
		}

		helloCert, err := cert.GetCertificate(hello.ServerName)
		if err != nil {
			log.Printf("couldn't get cert: %v", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return nil, err
		}

		return &helloCert, nil
	}

	clientConn, err := h.handshake(w, serverConfig)
	if err != nil {
		return
	}
	defer clientConn.Close()
	defer serverConn.Close()

	rp := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = r.Host
			r.URL.Scheme = "https"
		},
		Transport: &http.Transport{
			DialTLSContext: func(
				ctx context.Context, network string, address string,
			) (net.Conn, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return serverConn, nil
				}
			}},
	}

	wc := newOnCloseConn(clientConn)
	http.Serve(&oneShotListener{wc}, h.wrap(rp))
	wc.Wait()
}

func (h ProxyHandler) wrap(upstream http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := models.ConvertFromHttpRequest(*r)
		if err != nil {
			log.Printf("couldn't convert request: %v", err)
			return
		}

		err = h.usecase.SaveRequest(req)
		if err != nil {
			log.Printf("couldn't save request to db: %v", err)
			return
		}

		upstream.ServeHTTP(w, r)
	})
}

func (h ProxyHandler) handshake(w http.ResponseWriter, config *tls.Config) (net.Conn, error) {
	raw, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return nil, err
	}

	if _, err = raw.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		raw.Close()
		return nil, err
	}

	conn := tls.Server(raw, config)
	err = conn.Handshake()
	if err != nil {
		conn.Close()
		raw.Close()
		return nil, err
	}
	return conn, nil
}

type onCloseConn struct {
	net.Conn
	stopChan chan struct{}
}

func newOnCloseConn(conn net.Conn) *onCloseConn {
	c := &onCloseConn{conn, make(chan struct{})}
	return c
}

func (c *onCloseConn) Close() error {
	c.stopChan <- struct{}{}
	return c.Conn.Close()
}

func (c *onCloseConn) Wait() {
	<-c.stopChan
}

type oneShotListener struct {
	conn net.Conn
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.conn == nil {
		return nil, errors.New("conn is nil")
	}
	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}
