package transports

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type MessageTransport struct {
	Subscriber message.Subscriber
	Publisher  message.Publisher
}
