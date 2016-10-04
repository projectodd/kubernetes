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

type kubesh struct {
	finder     ResourceFinder
	context    []string
	lineReader *readline.Instance
	progname   string
	root       *cobra.Command
	factory    *cmdutil.Factory
}

func NewKubesh() *kubesh {
	factory := cmdutil.NewFactory(nil)
	sh := kubesh{
		factory:  factory,
		finder:   Resourceful{factory},
		progname: os.Args[0],
	}
	sh.root = sh.newRootCommand()
	completer := NewCompleter(sh.root, sh.finder, &sh.context)
	sh.lineReader, _ = readline.NewEx(&readline.Config{
		Prompt:       prompt([]string{}),
		AutoComplete: completer,
		HistoryFile:  path.Join(homedir.HomeDir(), ".kubesh_history"),
		Listener:     completer,
	})
	return &sh
}

func (sh *kubesh) Run() {
	defer sh.lineReader.Close()
	for {
		line, err := sh.lineReader.Readline()
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
			cmd := sh.newRootCommand()
			// TODO: what do we do with an error here? do we care?
			args, _ = applyContext(sh.context, args, cmd)
			sh.runKubeCommand(cmd, args)
		}
		// FIXME: if the command output something w/o a trailing \n, it
		// won't show
	}
}

func (sh *kubesh) Execute(args []string) {
	sh.root.SetArgs(args)
	sh.root.Execute()
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

func (sh *kubesh) newRootCommand() *cobra.Command {
	root := cmd.NewKubectlCommand(sh.factory, os.Stdin, os.Stdout, os.Stderr)
	get, _, err := root.Find([]string{"get"})
	if err != nil {
		panic(err)
	}
	cmd := &cobra.Command{
		Use:          "pin",
		Short:        "Pin resources for use in subsequent commands",
		Long:         pin_long,
		Example:      pin_example,
		SilenceUsage: true,
		ValidArgs:    get.ValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.GetFlagBool(cmd, "clear") {
				sh.context = []string{}
			} else {
				err := setContextCommand(sh, args)
				cmdutil.CheckErr(err)
			}
			sh.lineReader.SetPrompt(prompt(sh.context))
		},
	}
	cmd.Flags().BoolP("clear", "c", false, "Clears pinned resource")
	root.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "exit",
		Short: "Exit kubesh",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Bye!")
			os.Exit(0)
		},
	}
	root.AddCommand(cmd)
	return root
}

func prompt(context []string) string {
	path := ""
	if len(context) > 0 {
		path = fmt.Sprintf("[%v]", strings.Join(context, "/"))
	}

	return fmt.Sprintf(" kubesh%v> ", path)
}
