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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CommandCompleter struct {
	Root *cobra.Command
}

func (cc *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	cmd := cc.Root
	index := strings.LastIndex(string(line[:pos]), " ") + 1
	word := string(line[:pos])
	if index > 0 {
		word = word[index:pos]
		var err error
		cmd, _, err = cc.Root.Find(strings.Split(string(line), " "))
		if err != nil {
			return
		}
	}
	for _, completion := range completions(word, cmd) {
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

func completions(prefix string, cmd *cobra.Command) []string {
	candidates := []string{}
	if strings.HasPrefix(prefix, "-") {
		candidates = flags(cmd)
	} else {
		candidates = subCommands(cmd)
		if len(candidates) == 0 {
			candidates = resourceTypes(cmd)
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
	return cmd.ValidArgs
}

func flags(cmd *cobra.Command) []string {
	flags := []string{}
	fn := func(f *pflag.Flag) {
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
