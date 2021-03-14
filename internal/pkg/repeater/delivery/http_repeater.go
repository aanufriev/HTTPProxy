package delivery

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
		response += strconv.Itoa(int(request.ID)) + ": " + request.Method + " " + request.Scheme + "://" + request.Host + request.Path + "<br>"
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
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return
	}

	request, err := h.proxyUsecase.GetRequest(id)
	if err != nil {
		return
	}

	response := request.Method + " " + request.Scheme + "://" + request.Host + request.Path + "<br>"
	response += fmt.Sprintf("Headers: %s <br>", request.Headers)
	response += fmt.Sprintf("Body: %s <br>", request.Body)

	_, err = w.Write([]byte(
		fmt.Sprintf(responseTemplate, response),
	))

	if err != nil {
		return
	}
}
