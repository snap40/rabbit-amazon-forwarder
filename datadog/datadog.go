package datadog

import (
	"fmt"
	"github.com/AirHelp/rabbit-amazon-forwarder/utils"
	"github.com/DataDog/datadog-go/statsd"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

var (
	lock            = &sync.Mutex{}
	statsdSingleton *statsd.Client

	agentHost = utils.GetEnvOrDefault("DD_AGENT_HOST", "127.0.0.1")
	service   = utils.GetEnvOrDefault("DD_SERVICE", "rmq-forwarder")
	version   = utils.GetEnvOrDefault("DD_VERSION", "0.0.1")
	env       = utils.GetEnvOrDefault("DD_ENV", "dev")

	defaultTags = []string{
		fmt.Sprintf("env:%s", env),
		fmt.Sprintf("service:%s", service),
		fmt.Sprintf("version:%s", version),
	}
	suppliedTags = utils.GetEnvOrDefault("DD_TAGS", "")
)

// DefaultTags returns the default tags to be used for DataDog metrics, traces, and logs
func DefaultTags() []string {
	mergedTags := make([]string, len(defaultTags))
	copy(mergedTags, defaultTags)

	if suppliedTags != "" {
		tagsList := strings.Split(suppliedTags, " ")
		mergedTags = append(defaultTags, tagsList...)
	}

	return mergedTags
}

// TagsIncludingDefaults returns the additional tags list appended to the default tags
func TagsIncludingDefaults(additionalTags []string) []string {
	mergedTags := make([]string, len(defaultTags))
	copy(mergedTags, defaultTags)

	return append(mergedTags, additionalTags...)
}

// TagsToMap takes a []string whose elements are "tag:value" and returns a
// map[string]string version
func TagsToMap(tags []string) map[string]interface{} {
	tagsMap := make(map[string]interface{})

	for _, tag := range tags {
		splitTag := strings.SplitN(tag, ":", 2)
		tagsMap[splitTag[0]] = splitTag[1]
	}

	return tagsMap
}

// DefaultTagsAsMap returns all default tags as a map[string]string
func DefaultTagsAsMap() map[string]interface{} {
	return TagsToMap(DefaultTags())
}

// NewStatsd returns a *statsd.Client singleton
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
