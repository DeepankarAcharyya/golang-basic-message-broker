package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Topic   string
	Payload interface{}
}

type Subscriber struct {
	Channel     chan interface{} //channel that can transmit data of any type
	Unsubscribe chan bool
}

type Broker struct {
	subsribers map[string][]*Subscriber
	mutex      sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		subsribers: make(map[string][]*Subscriber),
	}
}

func (b *Broker) Subscribe(topic string) *Subscriber {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscriber := &Subscriber{
		Channel:     make(chan interface{}, 1),
		Unsubscribe: make(chan bool),
	}

	b.subsribers[topic] = append(b.subsribers[topic], subscriber)

	return subscriber
}

func (b *Broker) Unsubscribe(topic string, subscriber *Subscriber) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subsribers[topic]; found {
		for i, sub := range subscribers {
			if sub == subscriber {
				close(sub.Channel)
				b.subsribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				return
			}
		}
	}
}

func (b *Broker) Publish(topic string, payload interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers, found := b.subsribers[topic]; found {
		for _, sub := range subscribers {
			select {
			case sub.Channel <- payload:
			case <-time.After(time.Second):
				fmt.Printf("Subscriber slow. Unsubscribing from topic: %s\n", topic)
				b.Unsubscribe(topic, sub)
			}
		}
	}
}

func main() {
	broker := NewBroker()

	subscriber := broker.Subscribe("Topic_1")

	go func() {
		for {
			select {
			case msg, ok := <-subscriber.Channel:
				if !ok {
					fmt.Println("Subscriber channel closed.")
					return
				}
				fmt.Printf("Received: %v\n", msg)
			case <-subscriber.Unsubscribe:
				fmt.Println("Unsubscribed.")
				return
			}
		}
	}()

	broker.Publish("Topic_1", "Hello World!")
	broker.Publish("topic_1", "hello again!")
	broker.Publish("Topic_1", "This is a test message.")

	time.Sleep(2 * time.Second)
	broker.Unsubscribe("Topic_1", subscriber)

	broker.Publish("Topic_1", "This message wont be received by anyone.")
	time.Sleep(time.Second)
}
