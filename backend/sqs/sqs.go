package sqs

import (
	"time"

	sqs_client "github.com/AdRoll/goamz/sqs"
	"github.com/eliothedeman/statsdaemon/config"
	"github.com/eliothedeman/statsdaemon/metric"
)

// a backend which pushed data to any of a list of queues
type SQS struct {
	client *sqs_client.SQS
	queues []*sqs_client.Queue
	conf   *SQSConfig
}

// SQSConfig provides config information for the SQS provider
type SQSConfig struct {
	AccessKey string   `json:"access_key"`
	SecretKey string   `json:"secret_key"`
	Region    string   `json:"region"`
	Queues    []string `json:"queues"`
}

func (s *SQS) Submit(all []metric.Metric, deadline time.Time) error {
	return nil
}

func (s *SQS) Init(i interface{}) error {
	conf, ok := i.(*SQSConfig)
	if !ok {
		return config.WRONG_CONFIG_TYPE
	}

	client, err := sqs_client.NewFrom(conf.AccessKey, conf.SecretKey, conf.Region)
	if err != nil {
		return err
	}
	s.client = client

	s.queues = make([]*sqs_client.Queue, len(conf.Queues))
	for i, queue_name := range conf.Queues {
		queue, err := s.client.CreateQueue(queue_name)
		if err != nil {
			return err
		}

		s.queues[i] = queue
	}
	return nil
}

func (s *SQS) ConfigStruct() interface{} {
	return &SQSConfig{}
}
