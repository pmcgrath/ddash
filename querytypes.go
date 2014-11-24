package main

/*
Docker event - Subset of data, see https://github.com/docker/docker/blob/master/utils/jsonmessage.go and http://crosbymichael.com/docker-events.html
*/
type dEvent struct {
	Status string
	Id     string
	From   string
	Time   int64
}
