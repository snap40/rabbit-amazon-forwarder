package sqs

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
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

const (
	// Type forwarder type
	Type = "SQS"
)

// Forwarder forwarding client
type Forwarder struct {
	name      string
	sqsClient sqsiface.SQSAPI
	queue     string
}

var statsd = datadog.NewStatsd()

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, sqsClient ...sqsiface.SQSAPI) forwarder.Client {
	var client sqsiface.SQSAPI
	if len(sqsClient) > 0 {
		client = sqsClient[0]
	} else {
		client = sqs.New(session.Must(session.NewSession()))
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
func (f Forwarder) Push(message amqp.Delivery) error {
	messageBody := string(message.Body)

	if messageBody == "" {
		return errors.New(forwarder.EmptyMessageError)
	}
	params := &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody), // Required
		QueueUrl:    aws.String(f.queue),     // Required
	}

	resp, err := f.sqsClient.SendMessage(params)

	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	statsd.Count("message.sent", 1, []string{"type:sqs", fmt.Sprintf("target:%s", f.queue)}, 1)
	log.WithFields(log.Fields{
		"forwarderName": f.Name(),
		"responseID":    resp.MessageId}).Debug("Forward succeeded")
	return nil
}
