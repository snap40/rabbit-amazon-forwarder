package datadog

import (
	"github.com/AirHelp/rabbit-amazon-forwarder/utils"
	"github.com/DataDog/datadog-go/statsd"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	statsdSingleton *statsd.Client

	lock      = &sync.Mutex{}
	agentHost = utils.GetEnvOrDefault("DD_AGENT_HOST", "127.0.0.1:8125")
)

func NewStatsd() *statsd.Client {
	lock.Lock()
	defer lock.Unlock()

	var err error = nil
	if statsdSingleton == nil {
		statsdSingleton, err = statsd.New(agentHost)
		if err != nil {
			log.Fatal(err)
		}
	}

	return statsdSingleton
}
