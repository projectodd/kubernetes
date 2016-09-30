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
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

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
	progname         string
	root             *cobra.Command
}

func main() {
	factory := cmdutil.NewFactory(nil)
	finder := Resourceful{factory}
	kubectl := cmd.NewKubectlCommand(factory, os.Stdin, os.Stdout, os.Stderr)

	// If args are passed just run that kubectl command directly
	progArgs := os.Args[1:]
	if len(progArgs) > 0 {
		kubectl.SetArgs(progArgs)
		kubectl.Execute()
		return
	}

	cmdutil.BehaviorOnFatal(func(msg string, code int) {
		fmt.Println(msg)
		panic("kubectl")
	})

	completer := &CommandCompleter{Root: kubectl, Finder: finder}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt([]string{}),
		AutoComplete: completer,
		HistoryFile:  path.Join(homedir.HomeDir(), ".kubesh_history"),
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	sh := kubesh{
		root:   kubectl,
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
		progname: os.Args[0],
	}
	sh.addPinCommand()
	completer.Context = &sh.context

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
		args, err := tokenize(line)
		if err != nil {
			fmt.Println(err)
		} else {
			internal, err := sh.runInternalCommand(args)
			if err != nil {
				panic(err)
			}

			if !internal {
				kubectl := cmd.NewKubectlCommand(factory, os.Stdin, os.Stdout, os.Stderr)
				// TODO: what do we do with an error here? do we care?
				args, _ = applyContext(sh.context, args, kubectl)
				sh.runKubeCommand(kubectl, args)
			}
		}
		// FIXME: if the command output something w/o a trailing \n, it
		// won't show
	}
}

func (sh *kubesh) runKubeCommand(kubectl *cobra.Command, args []string) {
	defer func() {
		// Ignore any panics from kubectl
		recover()
	}()
	subcommand, _, _ := kubectl.Find(args)
	switch subcommand.Name() {
	case "proxy", "attach":
		sh.runExec(args)
	default:
		kubectl.SetArgs(args)
		kubectl.Execute()
	}
}

func (sh *kubesh) runExec(args []string) {
	cmd := exec.Command(sh.progname, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)
	go func() {
		<-signals
		fmt.Println("")
		syscall.Kill(cmd.Process.Pid, syscall.SIGINT)
	}()

	cmd.Wait()
}

func (sh *kubesh) runInternalCommand(args []string) (bool, error) {
	if len(args) > 0 {
		if f := sh.internalCommands[args[0]]; f != nil {

			return true, f(sh, args)
		}
	}

	return false, nil
}

func (sh *kubesh) addPinCommand() {
	get, _, err := sh.root.Find([]string{"get"})
	if err != nil {
		panic(err)
	}
	cmd := &cobra.Command{
		Use:       "pin",
		ValidArgs: get.ValidArgs,
		Run:       func(cmd *cobra.Command, args []string) {},
	}
	sh.root.AddCommand(cmd)
}

func prompt(context []string) string {
	path := ""
	if len(context) > 0 {
		path = fmt.Sprintf("[%v]", strings.Join(context, "/"))
	}

	return fmt.Sprintf("kubesh%v> ", path)
}
