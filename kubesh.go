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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	readline "gopkg.in/readline.v1"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func main() {
	cmdutil.BehaviorOnFatal(func(msg string, code int) {
		fmt.Println(msg)
	})

	cmd := cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	l, err := readline.NewEx(&readline.Config{
		Prompt:       ">>> ",
		AutoComplete: Completer(cmd),
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		cmd.ResetFlags()
		cmd.SetArgs(strings.Split(line, " "))
		cmd.Execute()
	}
}

func Completer(cmd *cobra.Command) readline.AutoCompleter {
	return readline.NewPrefixCompleter(mapPrefixes(cmd.Commands())...)
}

func mapPrefixes(cmds []*cobra.Command) []readline.PrefixCompleterInterface {
	prefixes := make([]readline.PrefixCompleterInterface, len(cmds))
	for i, cmd := range cmds {
		prefixes[i] = prefixForCommand(cmd)
	}
	return prefixes
}

func prefixForCommand(cmd *cobra.Command) readline.PrefixCompleterInterface {
	prefixes := prefixesForSubcommands(cmd)
	prefixes = append(prefixes, prefixesForResourceTypes(cmd)...)
	prefixes = append(prefixes, prefixesForFlags(cmd)...)
	return readline.PcItem(cmd.Name(), prefixes...)
}

func prefixesForSubcommands(cmd *cobra.Command) []readline.PrefixCompleterInterface {
	return mapPrefixes(cmd.Commands())
}

func prefixesForResourceTypes(cmd *cobra.Command) []readline.PrefixCompleterInterface {
	prefixes := make([]readline.PrefixCompleterInterface, len(cmd.ValidArgs))
	for i, arg := range cmd.ValidArgs {
		prefixes[i] = readline.PcItem(arg)
	}
	return prefixes
}

func prefixesForFlags(cmd *cobra.Command) []readline.PrefixCompleterInterface {
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

	prefixes := make([]readline.PrefixCompleterInterface, len(flags))
	for i := range flags {
		// we create enough completers to support including
		// one of each flag, which is a little weird, i admit
		prefixes[i] = readline.PcItemDynamic(func(line string) []string {
			// TODO: omit flags already present in line
			return flags
		})
	}
	return prefixes
}
