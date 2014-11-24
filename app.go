package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"code.google.com/p/go.net/websocket"
)

var (
	dockerHost      = flag.String("dockerhost", DOCKER_DEFAULT_HOST, "Docker host")
	applicationPort = flag.Int("port", 8090, "Port")

	queryer dockerQueryer
)

func init() {
	flag.Parse()

	queryer = newDockerQueryer(*dockerHost)
}

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

	containerAsJsonData, err := json.MarshalIndent(container, "", "    ")
	if err != nil {
		log.Printf("containerHandler: Convert to json data error for id: %s error: %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(containerAsJsonData))
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

	containersAsJsonData, err := json.Marshal(containers)
	if err != nil {
		log.Printf("containersHandler: Convert to json data error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageInfo := struct {
		SocketPath       string
		ContainersAsJson string
	}{
		"/events",
		string(containersAsJsonData),
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

	addr := fmt.Sprintf(":%d", *applicationPort)
	log.Printf("Using %s\n", runtime.Version())
	log.Printf("About to listen at %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Listen and server error : %s", err)
		os.Exit(1)
	}

	os.Exit(0)
}
