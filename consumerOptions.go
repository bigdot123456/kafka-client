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

package kafkaclient

import (
	"github.com/uber-go/kafka-client/internal/consumer"
)

type (
	// ConsumerOption is the type for optional arguments to the NewConsumer constructor.
	ConsumerOption interface {
		apply(*consumer.Options)
	}

	consumerBuildOptions []ConsumerOption

	partialConstruction struct {
		enabled bool
		errs    *consumerErrorList
	}
)

// PartialConstructionError returns a list of topics that could not be consumed as a list of ConsumerError.
// PartialConstructionError should be called on the ConsumerOption returned by EnablePartitionConstruction.
// This is useful if you chose to EnablePartialConstruction.
func PartialConstructionError(option ConsumerOption) []ConsumerError {
	pe, ok := option.(*partialConstruction)
	if !ok {
		return nil
	}

	return pe.errs.errs
}

// EnablePartialConstruction will set the client to return a partial consumer that
// consumes from as many topics/clusters as it could and it may return an error that lists the
// topics that failed to connect to their cluster.
// You can use ConsumerBuildError(err error) to most information about the failed topics, if any.
func EnablePartialConstruction() ConsumerOption {
	return &partialConstruction{
		enabled: true,
	}
}

func (p *partialConstruction) apply(opt *consumer.Options) {
	opt.PartialConstruction = p.enabled
}

func (c consumerBuildOptions) addPartialConstructionError(errs *consumerErrorList) {
	for _, opt := range c {
		pe, ok := opt.(*partialConstruction)
		if ok {
			pe.errs = errs
		}
	}
}