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
)

type CommandCompleter struct {
	Root    *cobra.Command
	Finder  ResourceFinder
	Context *[]string
}

func (cc *CommandCompleter) Do(lune []rune, pos int) (newLine [][]rune, offset int) {
	cmd, args, line := cc.Root, []string{}, string(lune[:pos])
	word := line
	lastSpace := strings.LastIndex(line, " ") + 1
	lastComma := strings.LastIndex(line, ",") + 1
	if lastSpace > 0 {
		if lastComma > lastSpace {
			word = word[lastComma:pos]
		} else {
			word = word[lastSpace:pos]
		}
		var err error
		cmd, args, err = cc.Root.Find(strings.Split(line, " "))
		if err != nil {
			return
		}
	}
	for _, completion := range cc.completions(word, cmd, args) {
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

func (cc *CommandCompleter) completions(prefix string, ccmd *cobra.Command, args []string) []string {
	candidates := []string{}
	cmd := &Command{ccmd}
	if strings.HasPrefix(prefix, "-") {
		candidates = flags(cmd)
	} else {
		candidates = subCommands(cmd)
		if len(candidates) == 0 {
			switch ccmd.Name() {
			case "logs", "attach", "exec", "port-forward":
				candidates = cc.resources("pods")
			case "rolling-update":
				candidates = cc.resources("rc")
			case "cordon", "uncordon", "drain":
				candidates = cc.resources("node")
			case "explain":
				getCmd, _, _ := cc.Root.Find([]string{"get"})
				candidates = resourceTypes(&Command{getCmd})
			default:
				if len(*cc.Context) == 1 {
					candidates = cc.resources((*cc.Context)[0])
				} else {
					if t := resourceType(cmd, args); len(t) > 0 {
						candidates = cc.resources(t)
					} else {
						candidates = resourceTypes(cmd)
					}
				}
			}
		}
	}
	return complete(prefix, candidates)
}

func (cc *CommandCompleter) resources(resourceType string) []string {

	resources, err := cc.Finder.Lookup([]string{resourceType})
	ret := []string{}

	if err == nil {
		for _, r := range resources {
			ret = append(ret, r.name)
		}
	}

	return ret
}

func complete(prefix string, candidates []string) (results []string) {
	for _, s := range candidates {
		if strings.HasPrefix(s, prefix) {
			results = append(results, s)
		}
	}
	return
}

func subCommands(cmd KubectlCommand) []string {
	cmds := cmd.SubCommands()
	results := make([]string, len(cmds))
	for _, c := range cmds {
		results = append(results, c+" ")
	}
	return results
}

func flags(cmd KubectlCommand) []string {
	flags := cmd.Flags()
	results := make([]string, len(flags))
	for _, f := range flags {
		flag := "--" + f.Name
		if f.Assignable {
			flag += "="
		} else {
			flag += " "
		}
		results = append(results, flag)
	}
	return results
}

func resourceTypes(cmd KubectlCommand) []string {
	types := cmd.ResourceTypes()
	args := make([]string, len(types))
	for i, v := range types {
		args[i] = v + " "
	}
	sort.Strings(args)
	return args
}

// resourceType returns the resource type identified in the args,
// which could be a comma-delimited list of multiple types
func resourceType(cmd KubectlCommand, args []string) string {
	x := cmd.NonFlags(args)
	if len(x) > 1 {
		return x[0]
	} else {
		return ""
	}
}
