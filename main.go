package main

import (
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

func main() {
	go eventDistr.Run(queryer)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/containers", containersHandler)
	http.HandleFunc("/containers/", containerHandler)
	http.Handle("/events", websocket.Handler(eventsHandler))

	addr := fmt.Sprintf(":%d", *applicationPort)
	log.Printf("Using runtime %s\n", runtime.Version())
	log.Printf("Commit = %s build @ %s Full commit = %s\n", shortCommitHash, buildDate, commitHash)
	log.Printf("About to listen at %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Listen and server error : %s", err)
		os.Exit(1)
	}

	os.Exit(0)
}
