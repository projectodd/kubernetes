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
	"path"
	"strings"

	"github.com/bbrowning/readline"
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/homedir"
)

type InternalCommand func(*kubesh, []string) error

type kubesh struct {
	finder           ResourceFinder
	context          []string
	rl               *readline.Instance
	internalCommands map[string]InternalCommand
}

func main() {
	cmdutil.BehaviorOnFatal(func(msg string, code int) {
		fmt.Println(msg)
		panic("kubectl")
	})

	factory := cmdutil.NewFactory(nil)
	finder := Resourceful{factory}
	kubectl := cmd.NewKubectlCommand(factory, os.Stdin, os.Stdout, os.Stderr)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt([]string{}),
		AutoComplete: &CommandCompleter{kubectl, finder},
		HistoryFile:  path.Join(homedir.HomeDir(), ".kubesh_history"),
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	sh := kubesh{
		finder: finder,
		rl:     rl,
		internalCommands: map[string]InternalCommand{
			"exit": func(_ *kubesh, _ []string) error {
				fmt.Println("Bye!")
				os.Exit(0)

				return nil
			},

			"pin": setContextCommand,
		},
	}

	fmt.Println("Welcome to kubesh, the kubectl shell!")
	fmt.Println("Type 'help' or <TAB> to see available commands")
	fmt.Println("For options/flags, tab complete a dash, '--<TAB>'")
	fmt.Println("Use 'pin' when multiple commands apply to same resource")
	fmt.Println("Use GNU readline key bindings for command editing and history")
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
			kubectl := cmd.NewKubectlCommand(factory, os.Stdin, os.Stdout, os.Stderr)
			// TODO: what do we do with an error here? do we care?
			args, _ = applyContext(sh.context, args, kubectl)
			kubectl.SetArgs(args)
			runKubeCommand(kubectl)
		}
	}
}

func runKubeCommand(kubectl *cobra.Command) {
	defer func() {
		// Ignore any panics from kubectl
		recover()
	}()
	kubectl.Execute()
}

func (sh *kubesh) runInternalCommand(args []string) (bool, error) {
	if len(args) > 0 {
		if f := sh.internalCommands[args[0]]; f != nil {

			return true, f(sh, args)
		}
	}

	return false, nil
}

func prompt(context []string) string {
	path := ""
	if len(context) > 0 {
		path = fmt.Sprintf("[%v]", strings.Join(context, "/"))
	}

	return fmt.Sprintf("kubesh%v> ", path)
}
