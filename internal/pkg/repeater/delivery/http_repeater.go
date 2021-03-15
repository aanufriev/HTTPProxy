package delivery

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aanufriev/httpproxy/internal/pkg/models"
	"github.com/aanufriev/httpproxy/internal/pkg/proxy/interfaces"
	"github.com/gorilla/mux"
)

const responseTemplate = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>requests</title>
	</head>
	<body>
	%s
	</body>
	</html>
`

type RepeatHandler struct {
	proxyUsecase interfaces.Usecase
}

func NewRepeaterHandler(proxyUsecase interfaces.Usecase) RepeatHandler {
	return RepeatHandler{
		proxyUsecase: proxyUsecase,
	}
}

func (h RepeatHandler) ShowAllRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.proxyUsecase.GetRequests()
	if err != nil {
		log.Printf("couldn't get requests: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	var response string
	for _, request := range requests {
		response += strconv.Itoa(int(request.ID))
		response += ": "
		response += request.Method
		response += " "
		response += request.Scheme
		response += "://"
		response += request.Host
		response += request.Path
		response += models.StrFromEncodedParams(request.Params)
		response += "<br>"
	}

	_, err = w.Write([]byte(
		fmt.Sprintf(responseTemplate, response),
	))
	if err != nil {
		log.Printf("couldn't write requests to client: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func (h RepeatHandler) ShowRequest(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Printf("couldn't get id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("couldn't convert id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	request, err := h.proxyUsecase.GetRequest(id)
	if err != nil {
		log.Printf("couldn't get request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	response := request.Method + " " + request.Scheme + "://" + request.Host + request.Path + "<br>"
	response += fmt.Sprintf("Headers: %s <br>", request.Headers)
	response += fmt.Sprintf("Body: %s <br>", request.Body)

	_, err = w.Write([]byte(
		fmt.Sprintf(responseTemplate, response),
	))
	if err != nil {
		log.Printf("couldn't write to client: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func (h RepeatHandler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Printf("couldn't get id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("couldn't convert id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	request, err := h.proxyUsecase.GetRequest(id)
	if err != nil {
		log.Printf("couldn't get request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	httpReq, err := models.ConvertToHttpRequest(request)
	if err != nil {
		log.Printf("couldn't convert to http request: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	client := http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(&httpReq)
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

func (h RepeatHandler) ScanRequest(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Printf("couldn't get id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("couldn't convert id")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	request, err := h.proxyUsecase.GetRequest(id)
	if err != nil {
		log.Printf("couldn't get request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	httpRequest, err := models.ConvertToHttpRequest(request)
	if err != nil {
		log.Printf("couldn't convert request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	client := http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var resp *http.Response

	query := httpRequest.URL.Query()
	for key, values := range query {
		for i, value := range values {
			query[key][i] = value + ";cat /etc/passwd;"
			httpRequest.URL.RawQuery = query.Encode()
			fmt.Println(httpRequest.URL.Query())

			resp, err = client.Do(&httpRequest)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}

			if strings.Contains(string(body), "root:") {
				fmt.Println("YES")
			}

			query[key][i] = value
			httpRequest.URL.RawQuery = query.Encode()
			fmt.Println(httpRequest.URL.Query())
		}
	}
}
