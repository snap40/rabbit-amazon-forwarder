package sqs

import (
	"errors"
	"github.com/streadway/amqp"
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

var badRequest = "Bad request"

func TestCreateForwarder(t *testing.T) {
	entry := config.AmazonEntry{Type: "SQS",
		Name:   "sqs-test",
		Target: "arn",
	}
	forwarder := CreateForwarder(entry)
	if forwarder.Name() != entry.Name {
		t.Errorf("wrong forwarder name, expected:%s, found: %s", entry.Name, forwarder.Name())
	}
}

func TestPush(t *testing.T) {
	queueName := "queue1"
	entry := config.AmazonEntry{Type: "SQS",
		Name:   "sqs-test",
		Target: queueName,
	}
	scenarios := []struct {
		name    string
		mock    sqsiface.SQSAPI
		message amqp.Delivery
		queue   string
		err     error
	}{
		{
			name:    "empty message",
			mock:    mockAmazonSQS{resp: sqs.SendMessageOutput{MessageId: aws.String("messageId")}, queue: queueName, message: amqp.Delivery{}},
			message: amqp.Delivery{},
			queue:   queueName,
			err:     errors.New(forwarder.EmptyMessageError),
		},
		{
			name:    "bad request",
			mock:    mockAmazonSQS{resp: sqs.SendMessageOutput{MessageId: aws.String("messageId")}, queue: queueName, message: amqp.Delivery{Body: []byte(badRequest)}},
			message: amqp.Delivery{Body: []byte(badRequest)},
			queue:   queueName,
			err:     errors.New(badRequest),
		},
		{
			name:    "success",
			mock:    mockAmazonSQS{resp: sqs.SendMessageOutput{MessageId: aws.String("messageId")}, queue: queueName, message: amqp.Delivery{Body: []byte("ABC")}},
			message: amqp.Delivery{Body: []byte("ABC")},
			queue:   queueName,
			err:     nil,
		},
	}
	for _, scenario := range scenarios {
		t.Log("Scenario name: ", scenario.name)
		forwarder := CreateForwarder(entry, scenario.mock)
		err := forwarder.Push(scenario.message)
		if scenario.err == nil && err != nil {
			t.Errorf("Error should not occur")
			return
		}
		if scenario.err == err {
			return
		}
		if err != nil && err.Error() != scenario.err.Error() {
			t.Errorf("wrong error, expecting:%v, got:%v", scenario.err, err)
		}
	}
}

type mockAmazonSQS struct {
	sqsiface.SQSAPI
	resp    sqs.SendMessageOutput
	queue   string
	message amqp.Delivery
}

func (m mockAmazonSQS) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	if *input.QueueUrl != m.queue {
		return nil, errors.New("wrong queue name")
	}
	if *input.MessageBody != string(m.message.Body) {
		return nil, errors.New("wrong message body")
	}
	if *input.MessageBody == badRequest {
		return nil, errors.New(badRequest)
	}
	return &m.resp, nil
}
