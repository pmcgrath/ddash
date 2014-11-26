package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

func containerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("containerHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	startIndex := len("/containers/")
	id := r.URL.Path[startIndex:]

	found, container, err := getContainer(queryer, id)
	if !found {
		log.Printf("containerHandler: Container not found for id: %s", id, err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("containerHandler: Get container error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prettyJsonData, err := json.MarshalIndent(container, "", "    ")
	if err != nil {
		log.Printf("containerHandler: Convert to pretty json data error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(prettyJsonData))
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("containersHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	containers, err := getContainers(queryer)
	if err != nil {
		log.Printf("containersHandler: Get containers error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prettyJsonData, err := json.MarshalIndent(containers, "", "    ")
	if err != nil {
		log.Printf("containersHandler: Convert to pretty json data error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(prettyJsonData))
}

func eventsHandler(ws *websocket.Conn) {
	log.Printf("eventsHandler: Registering connection for %s\n", ws.Request().RemoteAddr)
	disconnectedChannel := eventDistr.Register(ws)
	<-disconnectedChannel
	log.Printf("eventsHandler: Closing for %s\n", ws.Request().RemoteAddr)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("rootHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	pageInfo := struct {
		ContainerPathPrefix string
		ContainersPath      string
		SocketPath          string
	}{
		"/container/",
		"/containers",
		"/events",
	}

	err := rootTemplate.Execute(w, pageInfo)
	if err != nil {
		log.Printf("rootHandler: Execute template error : %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
