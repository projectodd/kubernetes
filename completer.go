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

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kubecmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type CommandCompleter struct {
	Root *cobra.Command
}

func (cc *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	cmd := cc.Root
	index := strings.LastIndex(string(line[:pos]), " ") + 1
	word := string(line[:pos])
	var nouns []string
	if index > 0 {
		word = word[index:pos]
		var args []string
		var err error
		cmd, args, err = cc.Root.Find(strings.Split(string(line), " "))
		if err != nil {
			return
		}
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				nouns = append(nouns, arg)
			}
		}
	}
	for _, completion := range completions(word, cmd, nouns) {
		if len(word) >= len(completion) {
			if len(word) == len(completion) {
				newLine = append(newLine, []rune{' '})
			} else {
				newLine = append(newLine, []rune(completion))
			}
			offset = len(completion)
		} else {
			newLine = append(newLine, []rune(completion)[len(word):])
			offset = len(word)
		}

	}
	return
}

func completions(prefix string, cmd *cobra.Command, nouns []string) []string {
	candidates := []string{}
	if strings.HasPrefix(prefix, "-") {
		candidates = flags(cmd)
	} else {
		candidates = subCommands(cmd)
		if len(candidates) == 0 {
			if len(nouns) > 1 {
				candidates = resources(nouns[0])
			} else {
				candidates = resourceTypes(cmd)
			}
		}
	}
	return complete(prefix, candidates)
}

func complete(prefix string, candidates []string) (results []string) {
	for _, s := range candidates {
		if strings.HasPrefix(s, prefix) {
			results = append(results, s)
		}
	}
	return
}

func subCommands(cmd *cobra.Command) []string {
	prefixes := make([]string, len(cmd.Commands()))
	for i, c := range cmd.Commands() {
		prefixes[i] = c.Name()
	}
	return prefixes
}

func resourceTypes(cmd *cobra.Command) []string {
	args := cmd.ValidArgs
	sort.Strings(args)
	return args
}

func resources(resourceType string) []string {
	// TODO: Replace this with something abstracted out for reuse
	// between here and setContextCommand in kubesh.go
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)
	cmd := kubecmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, writer, ioutil.Discard)
	callArgs := []string{"get", resourceType, "--output=template",
		"--template={{ range .items }}{{ .metadata.name }} {{ end }}"}
	cmd.SetArgs(callArgs)
	cmd.Execute()
	writer.Flush()
	content, err := ioutil.ReadAll(bufio.NewReader(&buff))
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return strings.Split(strings.TrimSpace(string(content)), " ")
}

func flags(cmd *cobra.Command) []string {
	flags := []string{}
	fn := func(f *pflag.Flag) {
		if len(f.Deprecated) > 0 || f.Hidden {
			return
		}
		flag := "--" + f.Name
		if len(f.NoOptDefVal) == 0 {
			flag += "="
		}
		flags = append(flags, flag)
	}
	cmd.NonInheritedFlags().VisitAll(fn)
	cmd.InheritedFlags().VisitAll(fn)
	return flags
}
