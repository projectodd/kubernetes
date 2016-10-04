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
	"bytes"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Command interface {
	Name() string
	SubCommands() []string
	ResourceTypes() []string
	Flags() []Flag
	NonFlags(args []string) []string
	Find(cli string) (Command, []string, error)
}

type KubectlCommand struct {
	ref *cobra.Command
}

type Flag struct {
	Name       string
	Assignable bool
	usage      string
	Shorthand  string
}

func (cmd KubectlCommand) Name() string {
	return cmd.ref.Name()
}

func (cmd KubectlCommand) SubCommands() []string {
	cmds := cmd.ref.Commands()
	results := make([]string, 0, len(cmds))
	for _, c := range cmds {
		if c.IsAvailableCommand() {
			results = append(results, c.Name())
		}
	}
	return results
}

func (cmd KubectlCommand) ResourceTypes() []string {
	args := cmd.ref.ValidArgs
	sort.Strings(args)
	return args
}

func (cmd KubectlCommand) Flags() []Flag {
	flags := []Flag{}
	fn := func(f *pflag.Flag) {
		if len(f.Deprecated) > 0 || f.Hidden {
			return
		}
		flag := Flag{
			Name:       f.Name,
			Assignable: (len(f.NoOptDefVal) == 0),
			usage:      f.Usage,
			Shorthand:  f.Shorthand,
		}
		flags = append(flags, flag)
	}
	cmd.ref.NonInheritedFlags().VisitAll(fn)
	cmd.ref.InheritedFlags().VisitAll(fn)
	return flags
}

func (cmd KubectlCommand) NonFlags(args []string) []string {
	cmd.ref.ParseFlags(args)
	return cmd.ref.Flags().Args()
}

func (cmd KubectlCommand) Find(cli string) (result Command, args []string, err error) {
	args, err = tokenize(cli)
	if err != nil {
		return
	}
	cc, args, err := cmd.ref.Find(args)
	if err != nil {
		return
	}
	return KubectlCommand{cc}, args, err
}

func (flag Flag) Usage() string {
	x := new(bytes.Buffer)
	format := "--%s: %s"
	if flag.Assignable {
		format = "--%s=: %s"
	}
	if len(flag.Shorthand) > 0 {
		format = "  -%s, " + format
	} else {
		format = "   %s   " + format
	}
	fmt.Fprintf(x, format, flag.Shorthand, flag.Name, flag.usage)
	return x.String()
}
