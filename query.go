package main

// See
//	docker/docker.go
//	api/client/cli.go
//	api/client/commands.go
//	api/client/utils.go
import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func getContainers(queryer dockerQueryer) ([]map[string]interface{}, error) {
	log.Println("getContainers: About to get containers list")
	containers_url := "/containers/json?all=1"

	resp, err := queryer(containers_url)
	if err != nil {
		log.Printf("getContainers: exexGet error: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		message := fmt.Sprintf("getContainers: Non 200 response code: %d", resp.StatusCode)
		log.Println(message)
		return nil, fmt.Errorf(message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := fmt.Sprintf("getContainers: Read response body error: %s", err)
		log.Println(message)
		return nil, fmt.Errorf(message)
	}

	var sourceContainers []map[string]interface{}
	if err := json.Unmarshal(data, &sourceContainers); err != nil {
		log.Printf("getContainers: Unmarshal source containers error: %s\n", err)
		return nil, err
	}

	containers := make([]map[string]interface{}, len(sourceContainers))
	for index, sourceContainer := range sourceContainers {
		id := sourceContainer["Id"].(string)
		log.Printf("getContainers: About to get container with Id: %s\n", id)
		found, container, err := getContainer(queryer, id)
		if !found {
			message := fmt.Sprintf("getContainers: Failed to find container with id: %s", id)
			log.Println(message)
			return nil, fmt.Errorf(message)
		}
		if err != nil {
			log.Printf("getContainers: Get container error for id: error: %s\n", id, err)
			return nil, err
		}

		containers[index] = container
	}
	log.Printf("getContainers: Completed with %d containers\n", len(containers))

	return containers, nil
}

func getContainer(queryer dockerQueryer, id string) (bool, map[string]interface{}, error) {
	log.Printf("getContainer: About to get for Id: %s\n", id)
	container_url := fmt.Sprintf("containers/%s/json", id)

	resp, err := queryer(container_url)
	if err != nil {
		log.Printf("getContainer: queryer error for id: %s error: %s\n", id, err)
		return false, nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// Good
	case 404:
		log.Printf("getContainer: Not found for id: %s\n", id)
		return false, nil, nil
	default:
		message := fmt.Sprintf("getContainer: Unexpected response code for id: %s code: %d", id, resp.StatusCode)
		log.Println(message)
		return false, nil, fmt.Errorf(message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := fmt.Sprintf("getContainer: Read response error for id: %s error: %s", id, err)
		log.Println(message)
		return false, nil, fmt.Errorf(message)
	}

	var container map[string]interface{}
	if err := json.Unmarshal(data, &container); err != nil {
		log.Printf("getContainer: Unmarshall error for id: %s error: %s", id, err)
		return false, nil, err
	}

	return true, container, nil
}

func watchForEvents(queryer dockerQueryer, outgoing chan<- dEvent) {
	log.Println("watchForEvents: About to start watching")
	events_url := "events"

	resp, err := queryer(events_url)
	if err != nil {
		log.Fatalf("watchForEvents: execGet error: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("watchForEvents: Non 200 returned: %d", resp.StatusCode)
	}

	var event dEvent
	decoder := json.NewDecoder(resp.Body)
	for {
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			// This will not terminate loop
			log.Printf("watchForEvents: Decode error: %s", err)
		} else {
			outgoing <- event
		}
	}
}
