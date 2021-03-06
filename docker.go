package main

// See
//	docker/docker.go
//	api/client/cli.go
//	api/client/commands.go
//	api/client/utils.go
import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	dockerAPIVersion  = "1.18" // Based on "docker version" command at this time
	dockerDefaultHost = "unix:///var/run/docker.sock"
)

type dockerQueryer func(string) (*http.Response, error)

func newDockerQueryer(host string) dockerQueryer {
	return func(url string) (*http.Response, error) {
		url = fmt.Sprintf("/v%s/%s", dockerAPIVersion, url)
		return execGet(host, url)
	}
}

func execGet(dockerHost, url string) (*http.Response, error) {
	protoAddrParts := strings.SplitN(dockerHost, "://", 2)
	proto := protoAddrParts[0]
	addr := protoAddrParts[1]

	scheme := "http"

	transport := &http.Transport{
		Dial: func(dial_network, dial_addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, 32*time.Second)
		},
	}
	if proto == "unix" {
		// No need in compressing for local communications
		transport.DisableCompression = true
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("execGet: New request error: %s", err)
		return nil, err
	}

	req.Header.Set("User-Agent", "Docker-Client")
	req.URL.Host = addr
	req.URL.Scheme = scheme

	client := http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("execGet: Make request error: %s", err)
		return nil, err
	}

	return resp, nil
}
