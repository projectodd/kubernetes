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

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type InternalCommand func(*kubesh, []string) error

type kubesh struct {
	factory          *cmdutil.Factory
	cmd              *cobra.Command
	context          []string
	rl               *readline.Instance
	internalCommands map[string]InternalCommand
}

func main() {
	cmdutil.BehaviorOnFatal(func(msg string, code int) {
		fmt.Println(msg)
	})

	factory := cmdutil.NewFactory(nil)
	cmd := cmd.NewKubectlCommand(factory, os.Stdin, os.Stdout, os.Stderr)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: &CommandCompleter{cmd},
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	sh := kubesh{
		factory: factory,
		cmd:     cmd,
		rl:      rl,
		internalCommands: map[string]InternalCommand{
			"exit": func(_ *kubesh, _ []string) error {
				fmt.Println("Bye!")
				os.Exit(0)

				return nil
			},

			"sc": setContextCommand,
		},
	}

	for {
		line, err := sh.rl.Readline()
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
		args := strings.Split(line, " ")
		internal, err := sh.runInternalCommand(args)
		if err != nil {
			panic(err)
		}

		if !internal {
			sh.cmd.ResetFlags()
			sh.cmd.SetArgs(args)
			sh.cmd.Execute()
		}
	}
}

func (sh *kubesh) runInternalCommand(args []string) (bool, error) {
	if len(args) > 0 {
		if f := sh.internalCommands[args[0]]; f != nil {

			return true, f(sh, args)
		}
	}

	return false, nil
}

func setContextCommand(sh *kubesh, args []string) error {
	if len(args) == 1 || len(args) > 3 {
		fmt.Println("Usage: " + args[0] + " TYPE [NAME]")

		//TODO: return an error?
		return nil
	}

	resources, err := sh.lookupResource(args[1:])
	if err != nil {

		return err
	}

	typeOnly := len(args) == 2

	if len(resources) > 0 {
		res := resources[0]
		typeName := res["type"]
		if typeOnly {
			sh.context = []string{typeName}
		} else {
			sh.context = []string{typeName, res["name"]}
		}
	}

	sh.rl.SetPrompt(prompt(sh.context))

	return nil
}

func prompt(context []string) string {
	return strings.Join(context, ":") + "> "
}

// takes a type, or type and resource name, returning a slice of maps of name and type for each record returned by the api
func (sh *kubesh) lookupResource(args []string) ([]map[string]string, error) {
	cmdNamespace, _, err := sh.factory.DefaultNamespace()
	if err != nil {

		return nil, err
	}

	mapper, typer := sh.factory.Object()
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(sh.factory.ClientForMapping), sh.factory.Decoder(true)).
		NamespaceParam(cmdNamespace).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		fmt.Println(err)

		return nil, nil
	}

	infos, err := r.Infos()
	if err != nil {
		fmt.Println(err)

		return nil, nil
	}

	ret := make([]map[string]string, 0, len(infos))
	for _, i := range infos {
		ret = append(ret, map[string]string{
			"type": i.Mapping.Resource,
			"name": i.Name,
		})
	}

	return ret, nil
}
