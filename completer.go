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
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type CommandCompleter struct {
	Root    *cobra.Command
	Factory *cmdutil.Factory
}

func (cc *CommandCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	cmd, args := cc.Root, []string{}
	word := string(line[:pos])
	lastSpace := strings.LastIndex(string(line[:pos]), " ") + 1
	lastComma := strings.LastIndex(string(line[:pos]), ",") + 1
	if lastSpace > 0 {
		if lastComma > lastSpace {
			word = word[lastComma:pos]
		} else {
			word = word[lastSpace:pos]
		}
		var err error
		cmd, args, err = cc.Root.Find(strings.Split(string(line), " "))
		if err != nil {
			return
		}
	}
	for _, completion := range completions(cc.Factory, word, cmd, args) {
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

func completions(factory *cmdutil.Factory, prefix string, cmd *cobra.Command, args []string) []string {
	candidates := []string{}
	if strings.HasPrefix(prefix, "-") {
		candidates = flags(cmd)
	} else {
		candidates = subCommands(cmd)
		if len(candidates) == 0 {
			switch cmd.Name() {
			case "logs", "attach", "exec", "port-forward":
				candidates = resources(factory, "pods")
			case "rolling-update":
				candidates = resources(factory, "rc")
			case "cordon", "uncordon", "drain":
				candidates = resources(factory, "node")
			default:
				if t := resourceType(args); len(t) > 0 {
					candidates = resources(factory, t)
				} else {
					candidates = resourceTypes(cmd)
				}
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
		prefixes[i] = c.Name() + " "
	}
	return prefixes
}

func resourceTypes(cmd *cobra.Command) []string {
	args := cmd.ValidArgs
	sort.Strings(args)
	return args
}

// resourceType returns the resource type identified in the args,
// which could be a comma-delimited list of multiple types
func resourceType(args []string) string {
	// TODO: something more sophisticated?
	x := []string{}
	for _, s := range args {
		if !strings.HasPrefix(s, "-") {
			x = append(x, s)
		}
	}
	if len(x) > 1 {
		return x[0]
	} else {
		return ""
	}
}

func resources(factory *cmdutil.Factory, resourceType string) []string {

	resources, err := lookupResource(factory, []string{resourceType})
	ret := []string{}

	if err == nil {
		for _, r := range resources {
			ret = append(ret, r.name)
		}
	}

	return ret
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
		} else {
			flag += " "
		}
		flags = append(flags, flag)
	}
	cmd.NonInheritedFlags().VisitAll(fn)
	cmd.InheritedFlags().VisitAll(fn)
	return flags
}
