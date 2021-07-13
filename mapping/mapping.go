package mapping

import (
	"github.com/snap40/rmq-aws-forwarder/datadog"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/snap40/rmq-aws-forwarder/connector"

	"github.com/snap40/rmq-aws-forwarder/config"
	"github.com/snap40/rmq-aws-forwarder/consumer"
	"github.com/snap40/rmq-aws-forwarder/forwarder"
	"github.com/snap40/rmq-aws-forwarder/lambda"
	"github.com/snap40/rmq-aws-forwarder/rabbitmq"
	"github.com/snap40/rmq-aws-forwarder/sns"
	"github.com/snap40/rmq-aws-forwarder/sqs"
)

var log = logrus.WithFields(logrus.Fields(datadog.DefaultTagsAsMap()))

// Client mapping client
type Client struct {
	helper Helper
}

// Helper interface for creating consumers and forwaders
type Helper interface {
	createConsumer(entry config.RabbitEntry) consumer.Client
	createForwarder(entry config.AmazonEntry) forwarder.Client
}

// ConsumerForwarderMapping mapping for consumers and forwarders
type ConsumerForwarderMapping struct {
	Consumer  consumer.Client
	Forwarder forwarder.Client
}

type helperImpl struct{}

// New creates new mapping client
func New(helpers ...Helper) Client {
	var helper Helper
	helper = helperImpl{}
	if len(helpers) > 0 {
		helper = helpers[0]
	}
	return Client{helper}
}

// Load loads mappings
func (c Client) Load() ([]ConsumerForwarderMapping, error) {
	var consumerForwarderMapping []ConsumerForwarderMapping

	log.Info("Loading consumer - forwarder pairs")
	pairsList, err := config.Load()
	if err != nil {
		return nil, err
	}

	for _, pair := range pairsList {
		consumer := c.helper.createConsumer(pair.Source)
		forwarder := c.helper.createForwarder(pair.Destination)
		consumerForwarderMapping = append(consumerForwarderMapping, ConsumerForwarderMapping{consumer, forwarder})
	}
	return consumerForwarderMapping, nil
}

func (c Client) loadFile() ([]byte, error) {
	filePath := os.Getenv(config.MappingFile)
	log.WithField("mappingFile", filePath).Info("Loading mapping file")
	return ioutil.ReadFile(filePath)
}

func (h helperImpl) createConsumer(entry config.RabbitEntry) consumer.Client {
	log.WithFields(logrus.Fields{
		"consumerType": entry.Type,
		"consumerName": entry.Name}).Info("Creating consumer")
	switch entry.Type {
	case rabbitmq.Type:
		rabbitConnector := connector.CreateConnector(entry.ConnectionURL)
		return rabbitmq.CreateConsumer(entry, rabbitConnector)
	}
	return nil
}

func (h helperImpl) createForwarder(entry config.AmazonEntry) forwarder.Client {
	log.WithFields(logrus.Fields{
		"forwarderType": entry.Type,
		"forwarderName": entry.Name}).Info("Creating forwarder")
	switch entry.Type {
	case sns.Type:
		return sns.CreateForwarder(entry)
	case sqs.Type:
		return sqs.CreateForwarder(entry)
	case lambda.Type:
		return lambda.CreateForwarder(entry)
	}
	return nil
}
