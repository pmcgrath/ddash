package main

import (
	"log"
	"net"
	"sync"

	"code.google.com/p/go.net/websocket"
)

var (
	eventChannel chan dEvent
	eventDistr   *eventDistributor
)

func init() {
	eventChannel = make(chan dEvent)

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
	Incomming   <-chan dEvent
	Subscribers []*subscriber
}

func (self *eventDistributor) Register(connection *websocket.Conn) <-chan struct{} {
	disconnectedChannel := make(chan struct{})
	subscriber := &subscriber{
		Connection:          connection,
		DisconnectedChannel: disconnectedChannel,
	}

	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	self.Subscribers = append(self.Subscribers, subscriber)

	return subscriber.DisconnectedChannel
}

func (self *eventDistributor) Run(queryer dockerQueryer) {
	go watchForEvents(queryer, eventChannel)

	for {
		event := <-self.Incomming
		log.Printf("Run: Got event %#v will attempt to publish to %d subscribers\n", event, len(self.Subscribers))
		if len(self.Subscribers) == 0 {
			continue
		}

		disconnectedSubscribers := make([]*subscriber, 0)
		for _, subscriber := range self.Subscribers {
			log.Printf("Run: Sending event to %s\n", self.Connection.Request().RemoteAddr)
			if err := websocket.JSON.Send(subscriber.Connection, event); err != nil {
				log.Printf("Run: Send error: %s\n", err)
				switch err.(type) {
				case *net.OpError:
					close(subscriber.DisconnectedChannel)
					disconnectedSubscribers = append(disconnectedSubscribers, subscriber)
				}
			}
		}

		self.Mutex.Lock()
		self.Subscribers = removeDisconnectedSubscribers(self.Subscribers, disconnectedSubscribers)
		self.Mutex.Unlock()
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
