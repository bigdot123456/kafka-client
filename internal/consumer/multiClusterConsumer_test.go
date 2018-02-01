// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package consumer

import (
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/kafka-client/kafka"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type MultiClusterConsumerTestSuite struct {
	suite.Suite
	consumer *MultiClusterConsumer
	config   *kafka.ConsumerConfig
	topics   kafka.ConsumerTopicList
	options  *Options
	msgCh    chan kafka.Message
}

func (s *MultiClusterConsumerTestSuite) SetupTest() {
	topic := kafka.ConsumerTopic{
		Topic: kafka.Topic{
			Name:       "unit-test",
			Cluster:    "production-cluster",
			BrokerList: nil,
		},
		DLQ: kafka.Topic{
			Name:       "unit-test-dlq",
			Cluster:    "dlq-cluster",
			BrokerList: nil,
		},
	}
	s.topics = []kafka.ConsumerTopic{topic}
	s.config = &kafka.ConsumerConfig{
		TopicList:   s.topics,
		GroupName:   "unit-test-cg",
		Concurrency: 4,
	}
	s.options = testConsumerOptions()
	s.msgCh = make(chan kafka.Message)
	s.consumer, _ = NewMultiClusterConsumer(
		s.config,
		s.topics,
		make(map[string]kafka.Consumer),
		make(map[string]SaramaConsumer),
		make(map[string]sarama.SyncProducer),
		s.msgCh,
		tally.NoopScope,
		zap.L(),
	)
}

func (s *MultiClusterConsumerTestSuite) TeardownTest() {
	s.consumer.Stop()
}

func TestMultiClusterConsumerSuite(t *testing.T) {
	suite.Run(t, new(MultiClusterConsumerTestSuite))
}

func (s *MultiClusterConsumerTestSuite) TestStartSucceeds() {
	cc1 := newMockConsumer("cc1", s.topics.TopicNames(), nil)
	cc2 := newMockConsumer("cc2", s.topics.TopicNames(), nil)
	s.consumer.clusterToConsumerMap["cc1"] = cc1
	s.consumer.clusterToConsumerMap["cc2"] = cc2

	s.NoError(s.consumer.Start())

	started, stopped := cc1.lifecycle.Status()
	s.True(started)
	s.False(stopped)
	started, stopped = cc2.lifecycle.Status()
	s.True(started)
	s.False(stopped)
}

func (s *MultiClusterConsumerTestSuite) TestStartConsumerCloseOnError() {
	cc1 := newMockConsumer("cc1", s.topics.TopicNames(), nil)
	cc2 := newMockConsumer("cc2", s.topics.TopicNames(), nil)
	cc2.startErr = errors.New("error")
	s.consumer.clusterToConsumerMap["cc1"] = cc1
	s.consumer.clusterToConsumerMap["cc2"] = cc2

	s.Error(s.consumer.Start())

	started, stopped := cc1.lifecycle.Status()
	s.True(stopped)
	s.True(started)
	started, stopped = cc2.lifecycle.Status()
	s.True(stopped)
	s.True(started)
}