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
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/bbrowning/readline"
	"github.com/spf13/cobra"
)

var flagRegex *regexp.Regexp = regexp.MustCompile(`.*--([a-z\-]+)=$`)

type CommandCompleter struct {
	root    Command
	finder  ResourceFinder
	context *[]string
}

func NewCompleter(cmd *cobra.Command, finder ResourceFinder, ctx *[]string) *CommandCompleter {
	return &CommandCompleter{KubectlCommand{cmd}, finder, ctx}
}

func (cc *CommandCompleter) Do(lune []rune, pos int) (newLine [][]rune, offset int) {
	var err error
	cmd, args, line := cc.root, []string{}, string(lune[:pos])
	word := line
	lastSpace := strings.LastIndex(line, " ") + 1
	lastComma := strings.LastIndex(line, ",") + 1
	if lastSpace > 0 {
		if lastComma > lastSpace {
			word = word[lastComma:pos]
		} else {
			word = word[lastSpace:pos]
		}
		cmd, args, err = cc.findCommand(line)
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

func (cc *CommandCompleter) findCommand(cli string) (result Command, args []string, err error) {
	result, args, err = cc.root.Find(cli)
	if strings.HasSuffix(cli, " ") {
		// space implies completed completion. without it,
		// we're not sure if we have a valid type or not. see
		// the len check in resourceType(args)
		args = append(args, "")
	}
	return
}

func (cc *CommandCompleter) completions(word string, cmd Command, args []string) []string {
	candidates := []string{}
	if strings.HasPrefix(word, "-") && !strings.HasSuffix(word, "=") {
		candidates = flags(cmd)
	} else {
		candidates = subCommands(cmd)
		if len(candidates) == 0 {
			switch cmd.Name() {
			case "logs", "attach", "exec", "port-forward":
				candidates = cc.resources("pods")
			case "rolling-update":
				candidates = cc.resources("rc")
			case "cordon", "uncordon", "drain":
				candidates = cc.resources("node")
			case "explain":
				getCmd, _, _ := cc.root.Find("get")
				candidates = resourceTypes(getCmd)
			default:
				if len(*cc.context) == 1 {
					candidates = cc.resources((*cc.context)[0])
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
	return complete(word, candidates)
}

func (cc *CommandCompleter) resources(resourceType string) []string {
	results := []string{}
	resources, err := cc.finder.Lookup([]string{resourceType})
	if err == nil {
		for _, r := range resources {
			results = append(results, r.name+" ")
		}
	}
	return results
}

func complete(prefix string, candidates []string) (results []string) {
	for _, s := range candidates {
		if strings.HasPrefix(s, prefix) {
			results = append(results, s)
		}
	}
	return
}

func subCommands(cmd Command) []string {
	cmds := cmd.SubCommands()
	results := make([]string, len(cmds))
	for i, c := range cmds {
		results[i] = c + " "
	}
	return results
}

func flags(cmd Command) []string {
	flags := cmd.Flags()
	results := make([]string, len(flags))
	for i, f := range flags {
		flag := "--" + f.Name
		if f.Assignable {
			flag += "="
		} else {
			flag += " "
		}
		results[i] = flag
	}
	return results
}

func resourceTypes(cmd Command) []string {
	types := cmd.ResourceTypes()
	dupe := map[string]bool{}
	args := make([]string, 0, len(types))
	for _, v := range types {
		if !dupe[v] {
			dupe[v] = true
			args = append(args, v+" ")
		}
	}
	sort.Strings(args)
	return args
}

// resourceType returns the resource type identified in the args,
// which could be a comma-delimited list of multiple types
func resourceType(cmd Command, args []string) string {
	x := cmd.NonFlags(args)
	if len(x) > 1 {
		return x[0]
	} else {
		return ""
	}
}

func (cc *CommandCompleter) lastFlag(cli string) (result Flag, err error) {
	match := flagRegex.FindStringSubmatch(cli)
	if len(match) > 1 {
		cmd, _, err := cc.findCommand(cli)
		if err == nil {
			for _, f := range cmd.Flags() {
				if match[1] == f.Name {
					return f, nil
				}
			}
		}
	}
	return result, errors.New("No flag found at end of command line")
}

func (cc *CommandCompleter) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	newLine, newPos = line, pos
	if key == readline.CharTab {
		flag, err := cc.lastFlag(string(line[:pos]))
		if err == nil {
			fmt.Printf("\n\n%s\n\n", flag.Usage())
			ok = true
		}
	}
	return
}
