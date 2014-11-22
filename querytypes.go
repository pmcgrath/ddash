package main

/*
Container list query types - Complete content of response body to wget on "api/containers"
*/
type clqPort struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

type clqContainer struct {
	Command string
	Created int64
	Id      string
	Image   string
	Names   []string
	Ports   []clqPort
	Status  string
}

/*
Container query types - Subset of docker inspect output
*/
type cqConfig struct {
	Domainname string
	Hostname   string
}

type cqRestartPolicy struct {
	MaximumRetryCount int
	Name              string
}

type cqHostConfig struct {
	NetworkMode     string
	PublishAllPorts bool
	RestartPolicy   cqRestartPolicy
	VolumesFrom     string
}

type cqPort struct {
	HostIp   string
	HostPort string
}

type cqNetworkSettings struct {
	IPAddress string
	Ports     map[string][]cqPort
}

type cqState struct {
	ExitCode   int
	FinishedAt string
	Paused     bool
	Pid        int
	Restarting bool
	Running    bool
	StartedAt  string
}

type cqContainer struct {
	Config          cqConfig
	HostConfig      cqHostConfig
	Id              string
	Image           string
	Name            string
	NetworkSettings cqNetworkSettings
	State           cqState
	Volumes         map[string]string
	VolumesRW       map[string]bool
}

/*
Docker event - Subset of data, see https://github.com/docker/docker/blob/master/utils/jsonmessage.go and http://crosbymichael.com/docker-events.html
*/
type dEvent struct {
	Status string
	Id     string
	From   string
	Time   int64
}
