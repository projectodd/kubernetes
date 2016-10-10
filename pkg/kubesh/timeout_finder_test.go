// Copyright 2016 Red Hat, Inc, and individual contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubesh

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TimedFinder struct {
	sleep time.Duration
}

func (tf TimedFinder) Lookup(args []string) ([]Resource, error) {
	time.Sleep(tf.sleep)
	return []Resource{
		Resource{"pod", "pod-1"},
		Resource{"pod", "pod-2"},
	}, nil
}

func TestLongLookupTimesOut(t *testing.T) {
	assert := assert.New(t)
	tf := TimeoutFinder{TimedFinder{time.Second}, time.Millisecond * 100}
	t0 := time.Now()
	_, err := tf.Lookup([]string{"pod"})
	t1 := time.Now()
	assert.NotNil(err)
	assert.WithinDuration(t1, t0, 500*time.Millisecond)
}

func TestShortLookupReturnsResult(t *testing.T) {
	assert := assert.New(t)
	tf := TimeoutFinder{TimedFinder{time.Millisecond}, time.Second}
	resources, err := tf.Lookup([]string{"pod"})
	assert.Nil(err)
	assert.Equal(2, len(resources))
}
