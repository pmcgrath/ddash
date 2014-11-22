package main

import (
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

var (
	host       string
	port       int
	dockerHost string
	queryer    dockerQueryer
)

func init() {
	port = 8080
	dockerHost = DOCKER_DEFAULT_HOST
	queryer = newDockerQueryer(dockerHost)
}

func containerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("containerHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	id := ""

	found, data, err := getContainerRaw(queryer, id)
	if !found {
		log.Printf("containerHandler: Container not found for id: %s", id, err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("containerHandler: Get container raw error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(data))
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("containersHandler: Unsupported method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	containers, err := getContainers(queryer)
	if err != nil {
		log.Printf("containersHandler: Get containers errorr: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageInfo := struct {
		SocketPath string
		Containers []*container
	}{
		"/events",
		containers,
	}

	err = containersTemplate.Execute(w, pageInfo)
	if err != nil {
		log.Printf("containersHandler: Execute template error : %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func eventsHandler(ws *websocket.Conn) {
	log.Printf("eventsHandler: Registering connection for %s\n", ws.Request().RemoteAddr)
	disconnectedChannel := eventDistr.Register(ws)
	<-disconnectedChannel
	log.Printf("eventsHandler: Closing for %s\n", ws.Request().RemoteAddr)
}

func main() {
	go eventDistr.Run(queryer)

	http.HandleFunc("/containers", containersHandler)
	http.HandleFunc("/containers/", containerHandler)
	http.Handle("/events", websocket.Handler(eventsHandler))

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
