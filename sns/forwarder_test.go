package sns

import (
	"errors"
	"github.com/streadway/amqp"
	"testing"

	"github.com/snap40/rmq-aws-forwarder/config"
	"github.com/snap40/rmq-aws-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

var badRequest = "bad request"

func TestCreateForwarder(t *testing.T) {
	entry := config.AmazonEntry{Type: "SNS",
		Name:   "sns-test",
		Target: "arn",
	}
	forwarder := CreateForwarder(entry)
	if forwarder.Name() != entry.Name {
		t.Errorf("wrong forwarder name, expected:%s, found: %s", entry.Name, forwarder.Name())
	}
}

func TestPush(t *testing.T) {
	topicName := "topic1"
	entry := config.AmazonEntry{Type: "SNS",
		Name:   "sns-test",
		Target: topicName,
	}
	scenarios := []struct {
		name    string
		mock    snsiface.SNSAPI
		message amqp.Delivery
		topic   string
		err     error
	}{
		{
			name:    "empty message",
			mock:    mockAmazonSNS{resp: sns.PublishOutput{MessageId: aws.String("messageId")}, topic: topicName, message: amqp.Delivery{}},
			message: amqp.Delivery{},
			topic:   topicName,
			err:     errors.New(forwarder.EmptyMessageError),
		},
		{
			name:    "bad request",
			mock:    mockAmazonSNS{resp: sns.PublishOutput{MessageId: aws.String("messageId")}, topic: topicName, message: amqp.Delivery{Body: []byte(badRequest)}},
			message: amqp.Delivery{Body: []byte(badRequest)},
			topic:   topicName,
			err:     errors.New(badRequest),
		},
		{
			name:    "success",
			mock:    mockAmazonSNS{resp: sns.PublishOutput{MessageId: aws.String("messageId")}, topic: topicName, message: amqp.Delivery{Body: []byte("ABC")}},
			message: amqp.Delivery{Body: []byte("ABC")},
			topic:   topicName,
			err:     nil,
		},
	}
	for _, scenario := range scenarios {
		t.Log("Scenario name: ", scenario.name)
		forwarder := CreateForwarder(entry, scenario.mock)
		err := forwarder.Push(scenario.message)
		if scenario.err == nil && err != nil {
			t.Errorf("error should not occur: %s", err.Error())
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

type mockAmazonSNS struct {
	snsiface.SNSAPI
	resp    sns.PublishOutput
	topic   string
	message amqp.Delivery
}

func (m mockAmazonSNS) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	if *input.TargetArn != m.topic {
		return nil, errors.New("wrong topic name")
	}
	if *input.Message != string(m.message.Body) {
		return nil, errors.New("wrong message body")
	}
	if *input.Message == badRequest {
		return nil, errors.New(badRequest)
	}
	return &m.resp, nil
}
