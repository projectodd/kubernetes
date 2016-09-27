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

package main // TODO: _test suffix

import (
	"os"
	"testing"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var completer *CommandCompleter

type TestFinder map[string][]string

func (f TestFinder) Lookup(args []string) ([]ResourceInfo, error) {
	t := args[0]
	result := make([]ResourceInfo, 0, len(f[t]))
	for _, name := range f[t] {
		result = append(result, ResourceInfo{t, name})
	}
	return result, nil
}

func TestFoo(t *testing.T) {
	if false {
		t.Error("Nope")
	}
}

func init() {
	kubectl := cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	completer = &CommandCompleter{kubectl, TestFinder{}}
}
