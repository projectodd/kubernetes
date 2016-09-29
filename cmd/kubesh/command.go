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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type KubectlCommand interface {
	SubCommands() []string
	ResourceTypes() []string
	Flags() []Flag
}

type Command struct {
	ref *cobra.Command
}

type Flag struct {
	Name       string
	Assignable bool
}

func (cmd *Command) SubCommands() []string {
	cmds := cmd.ref.Commands()
	results := make([]string, 0, len(cmds))
	for _, c := range cmds {
		if c.IsAvailableCommand() {
			results = append(results, c.Name())
		}
	}
	return results
}

func (cmd *Command) ResourceTypes() []string {
	args := cmd.ref.ValidArgs
	sort.Strings(args)
	return args
}

func (cmd *Command) Flags() []Flag {
	flags := []Flag{}
	fn := func(f *pflag.Flag) {
		if len(f.Deprecated) > 0 || f.Hidden {
			return
		}
		flag := Flag{f.Name, (len(f.NoOptDefVal) == 0)}
		flags = append(flags, flag)
	}
	cmd.ref.NonInheritedFlags().VisitAll(fn)
	cmd.ref.InheritedFlags().VisitAll(fn)
	return flags
}
