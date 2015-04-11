package main

// See
// http://crosbymichael.com/docker-events.html
// https://github.com/docker/docker/blob/master/utils/jsonmessage.go
import (
	"log"
	"net"
	"sync"

	"golang.org/x/net/websocket"
)

type event map[string]interface{}

var (
	eventChannel chan event
	eventDistr   *eventDistributor
)

func init() {
	eventChannel = make(chan event)

	eventDistr = &eventDistributor{
		Incomming:   eventChannel,
		Subscribers: make([]*subscriber, 0),
	}
}

type subscriber struct {
	Connection          *websocket.Conn // Connection
	DisconnectedChannel chan struct{}   // Channel used to notify subscriber http handler func that the client has been disconnected, the http handler can then terminate
}

type eventDistributor struct {
	Mutex       sync.Mutex
	Incomming   <-chan event
	Subscribers []*subscriber
}

func (ev *eventDistributor) Register(connection *websocket.Conn) <-chan struct{} {
	disconnectedChannel := make(chan struct{})
	subscriber := &subscriber{
		Connection:          connection,
		DisconnectedChannel: disconnectedChannel,
	}

	ev.Mutex.Lock()
	defer ev.Mutex.Unlock()
	ev.Subscribers = append(ev.Subscribers, subscriber)

	return subscriber.DisconnectedChannel
}

func (ev *eventDistributor) Run(queryer dockerQueryer) {
	go watchForEvents(queryer, eventChannel)

	for {
		event := <-ev.Incomming
		log.Printf("Run: Got event %#v will attempt to publish to %d subscribers\n", event, len(ev.Subscribers))
		if len(ev.Subscribers) == 0 {
			continue
		}

		var disconnectedSubscribers []*subscriber
		for _, subscriber := range ev.Subscribers {
			log.Printf("Run: Sending event to %s\n", subscriber.Connection.Request().RemoteAddr)
			if err := websocket.JSON.Send(subscriber.Connection, event); err != nil {
				log.Printf("Run: Send error: %s\n", err)
				switch err.(type) {
				case *net.OpError:
					close(subscriber.DisconnectedChannel)
					disconnectedSubscribers = append(disconnectedSubscribers, subscriber)
				}
			}
		}

		ev.Mutex.Lock()
		ev.Subscribers = removeDisconnectedSubscribers(ev.Subscribers, disconnectedSubscribers)
		ev.Mutex.Unlock()
	}
}

func removeDisconnectedSubscribers(subscribers, disconnectedSubscribers []*subscriber) []*subscriber {
	// See 	http://stackoverflow.com/questions/5020958/go-what-is-the-fastest-cleanest-way-to-remove-multiple-entries-from-a-slice
	//	https://code.google.com/p/go-wiki/wiki/SliceTricks
	if len(subscribers) == 0 || len(disconnectedSubscribers) == 0 {
		return subscribers
	}

	result := make([]*subscriber, (len(subscribers) - len(disconnectedSubscribers)))
	index := 0
	for _, subscriber := range subscribers {
		isDisconnectedSubscriber := false
		for _, disconnectedSubscriber := range disconnectedSubscribers {
			if subscriber == disconnectedSubscriber {
				isDisconnectedSubscriber = true
			}
		}

		if !isDisconnectedSubscriber {
			result[index] = subscriber
			index++
		}
	}

	return result
}
