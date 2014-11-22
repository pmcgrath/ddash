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
	"time"
)

type container struct {
	CreatedAt                      time.Time
	DomainName                     string
	ExitCode                       int
	FinishedAt                     time.Time
	HostName                       string
	Id                             string
	Image                          string
	IPAddress                      string
	Name                           string
	NetworkMode                    string
	Paused                         bool
	Pid                            int
	PublishAllPorts                bool
	Restarting                     bool
	RestartPolicyMaximumRetryCount int
	RestartPolicyName              string
	Running                        bool
	StartedAt                      time.Time
	VolumesFrom                    string
}

func getContainers(queryer dockerQueryer) ([]*container, error) {
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

	var containerSummaries []clqContainer
	if err := json.Unmarshal(data, &containerSummaries); err != nil {
		log.Printf("getContainers: Unmarshal containers list error: %s\n", err)
		return nil, err
	}

	containers := make([]*container, len(containerSummaries))
	for index, containerSummary := range containerSummaries {
		log.Printf("getContainers: About to get container with Id: %s\n", containerSummary.Id)
		found, container, err := getContainer(queryer, containerSummary.Id)
		if !found {
			message := fmt.Sprintf("getContainers: Failed to find container with id: %s", containerSummary.Id)
			log.Println(message)
			return nil, fmt.Errorf(message)
		}
		if err != nil {
			log.Printf("getContainers: Get container error for id: error: %s\n", containerSummary.Id, err)
			return nil, err
		}

		containers[index] = container
	}
	log.Printf("getContainers: Completed with %d containers\n", len(containers))

	return containers, nil
}

func getContainer(queryer dockerQueryer, id string) (bool, *container, error) {
	log.Printf("getContainer: About to get for Id: %s\n", id)
	found, data, err := getContainerRaw(queryer, id)
	if !found || err != nil {
		// Not found or error
		return found, nil, err
	}

	var containerDetail cqContainer
	if err := json.Unmarshal(data, &containerDetail); err != nil {
		log.Printf("getContainer: Unmarshal container error: %s\n", err)
		return false, nil, err
	}

	container := &container{
		DomainName:                     containerDetail.Config.Domainname,
		ExitCode:                       containerDetail.State.ExitCode,
		FinishedAt:                     parseTimestamp(containerDetail.State.FinishedAt),
		HostName:                       containerDetail.Config.Hostname,
		Id:                             containerDetail.Id,
		Image:                          containerDetail.Image,
		IPAddress:                      containerDetail.NetworkSettings.IPAddress,
		Name:                           containerDetail.Name,
		NetworkMode:                    containerDetail.HostConfig.NetworkMode,
		Paused:                         containerDetail.State.Paused,
		Pid:                            containerDetail.State.Pid,
		PublishAllPorts:                containerDetail.HostConfig.PublishAllPorts,
		Restarting:                     containerDetail.State.Restarting,
		RestartPolicyMaximumRetryCount: containerDetail.HostConfig.RestartPolicy.MaximumRetryCount,
		RestartPolicyName:              containerDetail.HostConfig.RestartPolicy.Name,
		Running:                        containerDetail.State.Running,
		StartedAt:                      parseTimestamp(containerDetail.State.StartedAt),
		VolumesFrom:                    containerDetail.HostConfig.VolumesFrom,
	}
	log.Printf("getContainer: Completed getting for Id: %s\n", id)

	return true, container, nil
}

func getContainerRaw(queryer dockerQueryer, id string) (bool, []byte, error) {
	log.Printf("getContainerRaw: About to get for Id: %s\n", id)
	container_url := fmt.Sprintf("containers/%s/json", id)

	resp, err := queryer(container_url)
	if err != nil {
		log.Printf("getContainerRaw: queryer error for id: %s error: %s\n", id, err)
		return false, nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// Good
	case 404:
		log.Printf("getContainerRaw: Not found for id: %s\n", id)
		return false, nil, nil
	default:
		message := fmt.Sprintf("getContainerRaw: Unexpected response code for id: %s code: %d", id, resp.StatusCode)
		log.Println(message)
		return false, nil, fmt.Errorf(message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := fmt.Sprintf("getContainerRaw: Read response error for id: %s error: %s", id, err)
		log.Println(message)
		return false, nil, fmt.Errorf(message)
	}

	return true, data, nil
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

func parseTimestamp(data string) time.Time {
	res, _ := time.Parse(time.RFC3339, data)
	return res
}
