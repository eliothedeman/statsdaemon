package sqs

import (
	"encoding/json"
	"os"
	"time"

	sqs_client "github.com/AdRoll/goamz/sqs"
	"github.com/eliothedeman/statsdaemon/backend"
	"github.com/eliothedeman/statsdaemon/config"
	"github.com/eliothedeman/statsdaemon/metric"
)

func init() {
	backend.LoadBackend("sqs", &SQS{})
}

type SQSMetric struct {
	Host    string      `json:"host"`
	Plugin  string      `json:"plugin"`
	SubType string      `json:"subtype"`
	Type    string      `json:"type"`
	Kind    string      `json:"kind"`
	Value   interface{} `json:"value"`
	Time    int64       `json:"time"`
}

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

func (s *SQS) Submit(all []metric.Metric, now time.Time) error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	template := &SQSMetric{
		Host: host,
		Kind: "metric",
		Time: now.Unix(),
	}

	for _, m := range all {

		switch m := m.(type) {
		default:
		case *metric.Timer:
			val := m.MapValue()
			for k, v := range val {
				template.Plugin = m.Name() + k
				template.Value = v
				for _, q := range s.queues {
					err = sendMessage(template, q)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func sendMessage(m *SQSMetric, q *sqs_client.Queue) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	_, err = q.SendMessage(string(b))
	return err
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
