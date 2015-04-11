package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"golang.org/x/net/websocket"
)

var (
	containerPathRegexp *regexp.Regexp
)

func init() {
	var err error
	containerPathRegexp, err = regexp.Compile(`^/containers/\w{64}/?$`)
	if err != nil {
		panic(fmt.Sprintf("Container regex error : %s", err))
	}
}

func containerHandler(w http.ResponseWriter, r *http.Request) {
	if !containerPathRegexp.MatchString(r.URL.Path) {
		log.Printf("containerHandler: Unsupported url: %s", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		log.Printf("containerHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	startIndex := len("/containers/")
	id := r.URL.Path[startIndex:]

	found, container, err := getContainer(queryer, id)
	if !found {
		log.Printf("containerHandler: Container not found for id: %s", id)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("containerHandler: Get container error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prettyJSONData, err := json.MarshalIndent(container, "", "    ")
	if err != nil {
		log.Printf("containerHandler: Convert to pretty json data error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(prettyJSONData))
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/containers" {
		log.Printf("containersHandler: Unsupported url: %s", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

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

	prettyJSONData, err := json.MarshalIndent(containers, "", "    ")
	if err != nil {
		log.Printf("containersHandler: Convert to pretty json data error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(prettyJSONData))
}

func eventsHandler(ws *websocket.Conn) {
	log.Printf("eventsHandler: Registering connection for %s\n", ws.Request().RemoteAddr)
	disconnectedChannel := eventDistr.Register(ws)
	<-disconnectedChannel
	log.Printf("eventsHandler: Closing for %s\n", ws.Request().RemoteAddr)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Printf("rootHandler: Unsupported url: %s", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

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
		"/containers/",
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
