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

import "io"

type NewlineEnsuringWriter struct {
	delegate io.Writer
	lastByte byte
	written  bool
}

func (w *NewlineEnsuringWriter) Write(data []byte) (int, error) {
	if len(data) > 0 {
		w.lastByte = data[len(data)-1]
		w.written = true
	}

	return w.delegate.Write(data)
}

func (w *NewlineEnsuringWriter) EnsureNewline() error {
	if w.written && w.lastByte != '\n' {
		_, err := w.Write([]byte{'\n'})
		w.written = false

		return err
	}

	return nil
}
