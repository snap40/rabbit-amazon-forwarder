package main

import (
	"fmt"
	"github.com/AirHelp/rabbit-amazon-forwarder/datadog"
	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
	"github.com/AirHelp/rabbit-amazon-forwarder/supervisor"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

const (
	LogLevel = "LOG_LEVEL"
)

var log = logrus.WithFields(logrus.Fields(datadog.DefaultTagsAsMap()))

func getListenerAddress() string {
	var port string
	addressPattern := "0.0.0.0:%s"

	port = os.Getenv("NOMAD_PORT_http")
	if port != "" {
		return fmt.Sprintf(addressPattern, port)
	}

	port = os.Getenv("SERVER_PORT")
	if port != "" {
		return fmt.Sprintf(addressPattern, port)
	}

	return fmt.Sprintf(addressPattern, "8080")
}

func main() {
	createLogger()

	consumerForwarderMapping, err := mapping.New().Load()
	if err != nil {
		log.WithField("error", err.Error()).Fatalf("Could not load consumer - forwarder pairs")
	}
	supervisor := supervisor.New(consumerForwarderMapping)
	if err := supervisor.Start(); err != nil {
		log.WithField("error", err.Error()).Fatal("Could not start supervisor")
	}
	http.HandleFunc("/restart", supervisor.Restart)
	http.HandleFunc("/health", supervisor.Check)
	log.Info("Starting http server")
	log.Fatal(http.ListenAndServe(getListenerAddress(), nil))
}

func createLogger() {
	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "message",
		},
	}

	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	if logLevel := os.Getenv(LogLevel); logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err != nil {
			log.Fatal(err)
		} else {
			logrus.SetLevel(level)
		}
	}
}
