package lambda

import (
	"errors"
	"fmt"
	"github.com/snap40/rmq-aws-forwarder/config"
	"github.com/snap40/rmq-aws-forwarder/datadog"
	"github.com/snap40/rmq-aws-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	// Type forwarder type
	Type = "Lambda"
)

// Forwarder forwarding client
type Forwarder struct {
	name         string
	lambdaClient lambdaiface.LambdaAPI
	function     string
}

var statsd = datadog.NewStatsd()
var log = logrus.WithFields(logrus.Fields(datadog.DefaultTagsAsMap()))

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, lambdaClient ...lambdaiface.LambdaAPI) forwarder.Client {
	var client lambdaiface.LambdaAPI
	if len(lambdaClient) > 0 {
		client = lambdaClient[0]
	} else {
		client = lambda.New(session.Must(session.NewSession()))
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
	message_body := string(message.Body)

	if message_body == "" {
		return errors.New(forwarder.EmptyMessageError)
	}
	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.function),
		Payload:      []byte(message_body),
	}
	resp, err := f.lambdaClient.Invoke(params)
	if err != nil {
		log.WithFields(logrus.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	if resp.FunctionError != nil {
		log.WithFields(logrus.Fields{
			"forwarderName": f.Name(),
			"functionError": *resp.FunctionError}).Errorf("Could not forward message")
		return errors.New(*resp.FunctionError)
	}
	statsd.Count("messages.sent", 1, f.dataDogTags(), 1)
	log.WithFields(logrus.Fields{
		"forwarderName": f.Name(),
		"statusCode":    resp.StatusCode}).Debug("Forward succeeded")
	return nil
}

func (f Forwarder) dataDogTags() []string {
	return datadog.TagsIncludingDefaults([]string{
		"type:lambda",
		fmt.Sprintf("destination:%s", f.function),
	})
}
