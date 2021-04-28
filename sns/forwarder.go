package sns

import (
	"errors"
	"fmt"
	"github.com/AirHelp/rabbit-amazon-forwarder/datadog"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

const (
	// Type forwarder type
	Type = "SNS"
)

// Forwarder forwarding client
type Forwarder struct {
	name      string
	snsClient snsiface.SNSAPI
	topic     string
}

var statsd = datadog.NewStatsd()

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, snsClient ...snsiface.SNSAPI) forwarder.Client {
	var client snsiface.SNSAPI
	if len(snsClient) > 0 {
		client = snsClient[0]
	} else {
		client = sns.New(session.Must(session.NewSession()))
	}
	forwarder := Forwarder{entry.Name, client, entry.Target}
	log.WithField("forwarderName", forwarder.Name()).Info("Created forwarder")
	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(d amqp.Delivery) error {
	var messageBody string
	if d.ContentType == "application/octet-stream" {
		messageBody = string(forwarder.Base64Encode(d.Body))
	} else {
		messageBody = string(d.Body)
	}

	messageAttributes := make(map[string]*sns.MessageAttributeValue)
	if len(d.Headers) > 0 {
		for k, v := range d.Headers {
			stringValue := fmt.Sprintf("%s", v)
			messageAttributes[k] = &sns.MessageAttributeValue{}
			messageAttributes[k].SetStringValue(stringValue)
			messageAttributes[k].SetDataType("String")
		}
	}

	if messageBody == "" {
		return errors.New(forwarder.EmptyMessageError)
	}
	params := &sns.PublishInput{
		Message:   aws.String(messageBody),
		TargetArn: aws.String(f.topic),
	}
	if len(messageAttributes) > 0 {
		params.MessageAttributes = messageAttributes
	}

	resp, err := f.snsClient.Publish(params)
	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	statsd.Count("messages.sent", 1, []string{"type:sns", fmt.Sprintf("destination:%s", f.topic)}, 1)
	log.WithFields(log.Fields{
		"forwarderName": f.Name(),
		"responseID":    resp.MessageId}).Debug("Forward succeeded")
	return nil
}
