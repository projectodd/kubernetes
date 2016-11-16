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
	"errors"
	"time"
)

type TimeoutFinder struct {
	Delegate ResourceFinder
	Timeout  time.Duration
}

type resourcesError struct {
	resources []Resource
	err       error
}

func (tf TimeoutFinder) Lookup(args []string) ([]Resource, error) {
	cr := make(chan resourcesError, 1)
	go func() {
		defer func() {
			// Ignore any panics from the delegate
			recover()
		}()
		resources, err := tf.Delegate.Lookup(args)
		cr <- resourcesError{resources, err}
	}()
	select {
	case ie := <-cr:
		return ie.resources, ie.err
	case <-time.After(tf.Timeout):
		return nil, errors.New("Timed out connecting to the Kubernetes API")
	}
}
