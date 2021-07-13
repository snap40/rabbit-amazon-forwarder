package consumer

import "github.com/snap40/rmq-aws-forwarder/forwarder"

// Client intarface for consuming messages
type Client interface {
	Name() string
	Start(forwarder.Client, chan bool, chan bool) error
}
