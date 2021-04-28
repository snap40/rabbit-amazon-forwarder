package config

import (
	"encoding/json"
	"errors"
	"github.com/AirHelp/rabbit-amazon-forwarder/datadog"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"

	vault "github.com/hashicorp/vault/api"
)

const (
	MappingFile = "MAPPING_FILE"
	VaultPath   = "VAULT_PATH"
	CaCertFile  = "CA_CERT_FILE"
	CertFile    = "CERT_FILE"
	KeyFile     = "KEY_FILE"
)

var log = logrus.WithFields(logrus.Fields(datadog.DefaultTagsAsMap()))

var filePath = os.Getenv(MappingFile)
var vaultPath = os.Getenv(VaultPath)
var token = os.Getenv("VAULT_TOKEN")
var vaultAddr = os.Getenv("VAULT_ADDR")

// RabbitEntry RabbitMQ mapping entry
type RabbitEntry struct {
	Type          string   `json:"type" mapstructure:"type"`
	Name          string   `json:"name" mapstructure:"name"`
	ConnectionURL string   `json:"connection" mapstructure:"connection"`
	ExchangeName  string   `json:"topic" mapstructure:"topic"`
	QueueName     string   `json:"queue" mapstructure:"queue"`
	RoutingKey    string   `json:"routing" mapstructure:"routing"`
	RoutingKeys   []string `json:"routingKeys" mapstructure:"routingKeys"`
}

// AmazonEntry SQS/SNS mapping entry
type AmazonEntry struct {
	Type   string `json:"type" mapstructure:"type"`
	Name   string `json:"name" mapstructure:"name"`
	Target string `json:"target" mapstructure:"target"`
}

type SourceDestPairs []sourceDestPair

type sourceDestPair struct {
	Source      RabbitEntry `json:"source"`
	Destination AmazonEntry `json:"destination"`
}

func loadPairsFromFile() (SourceDestPairs, error) {
	filePath := os.Getenv(MappingFile)
	log.WithField("mappingFile", filePath).Info("Loading mapping file")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pairsList SourceDestPairs
	if err = json.Unmarshal(data, &pairsList); err != nil {
		return nil, err
	}
	log.Info("Loading consumer - forwarder pairs")
	return pairsList, nil
}

func loadPairsFromVault() (SourceDestPairs, error) {
	config := &vault.Config{
		Address: vaultAddr,
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	secret, err := client.Logical().Read(vaultPath)
	if err != nil {
		log.WithField("vaultPath", vaultPath).Error(err.Error())
		return nil, err
	}

	source, ok := secret.Data["source"].(map[string]interface{})
	if !ok {
		log.WithField("vaultPath", vaultPath).Error("source key not found in config")
	}
	var sourceStruct RabbitEntry
	err = mapstructure.Decode(source, &sourceStruct)
	if err != nil {
		panic(err)
	}

	destination, ok := secret.Data["destination"].(map[string]interface{})
	if !ok {
		panic(errors.New("destination key not found in config"))
	}
	var destinationStruct AmazonEntry
	err = mapstructure.Decode(destination, &destinationStruct)
	if err != nil {
		panic(err)
	}

	return SourceDestPairs{
		sourceDestPair{Source: sourceStruct, Destination: destinationStruct},
	}, nil
}

func Load() (SourceDestPairs, error) {
	if vaultPath != "" {
		return loadPairsFromVault()
	} else if filePath != "" {
		return loadPairsFromFile()
	} else {
		panic(errors.New("no config mode supplied"))
	}
}
