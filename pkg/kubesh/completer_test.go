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

package kubesh_test

import (
	"os"
	"testing"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubesh"
)

type TestFinder map[string][]string

var resources = TestFinder{
	"pod":     {"p1", "p2"},
	"service": {"s1", "s2", "s3"},
}
var context []string
var completer = kubesh.NewCompleter(
	cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr),
	resources,
	&context,
)

func TestCompleteCommands(t *testing.T) {
	affirm(t, in("g", 1), out(1, "et "))
	affirm(t, in("ro", 2), out(2, "lling-update ", "llout "))
}

func TestCompleteSubCommands(t *testing.T) {
	affirm(t, in("create dep", 10), out(3, "loyment "))
	affirm(t, in("create dep", 9), out(2, "ployment "))
}

func TestCompleteResourceTypes(t *testing.T) {
	affirm(t, in("get dep", 7), out(3, "loyment "))
	affirm(t, in("get pod,dep", 11), out(3, "loyment "))
}

func TestCompleteResources(t *testing.T) {
	affirm(t, in("get pod p", 9), out(1, "1 ", "2 "))
	affirm(t, in("get service ", 12), out(0, "s1 ", "s2 ", "s3 "))
	affirm(t, in("get service s1", 14), out(2, " "))
}

func TestCompleteFlags(t *testing.T) {
	affirm(t, in("get --output", 12), out(8, "=", "-version="))
	affirm(t, in("get --output= pod", 13), out(0))
}

func TestCompleteWithContext(t *testing.T) {
	context = []string{"pod"}
	affirm(t, in("get ", 4), out(0, "p1 ", "p2 "))
	context = []string{"pod", "p1"}
	affirm(t, in("get serv", 8), out(4, "ice ", "iceaccount "))
	context = []string{}
}

func (f TestFinder) Lookup(args []string) ([]kubesh.Resource, error) {
	t := args[0]
	result := make([]kubesh.Resource, 0, len(f[t]))
	for _, name := range f[t] {
		result = append(result, kubesh.Resource{Type: t, Name: name})
	}
	return result, nil
}

type input struct {
	line string
	pos  int
}

type output struct {
	candidates []string
	length     int
}

func in(line string, pos int) input {
	return input{line, pos}
}
func out(length int, candidates ...string) output {
	return output{candidates, length}
}
func got(candidates [][]rune, length int) output {
	strs := make([]string, len(candidates))
	for i := range candidates {
		strs[i] = string(candidates[i])
	}
	return output{strs, length}
}

func affirm(t *testing.T, in input, out output) {
	newLine, length := completer.Do([]rune(in.line), in.pos)
	pass := length == out.length
	if pass {
		pass = len(newLine) == len(out.candidates)
		if pass {
			for i, v := range newLine {
				if string(v) != out.candidates[i] {
					pass = false
					break
				}
			}
		}
	}
	if !pass {
		t.Error("Expected", out, "From", in, "Got", got(newLine, length))
	}
}
