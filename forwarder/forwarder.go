package forwarder

import (
	"encoding/base64"
	"github.com/streadway/amqp"
)

const (
	// EmptyMessageError empty error message
	EmptyMessageError = "message is empty"
)

// Client interface to forwarding messages
type Client interface {
	Name() string
	Push(message amqp.Delivery) error
}

func Base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}
