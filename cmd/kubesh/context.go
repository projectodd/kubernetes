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

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubectl"
)

var help = dedent.Dedent(`
      Pin to a resource or resource type.

      Pinning causes the shell to remember the given resource and/or
      resource type, and apply it to commands as appropriate, allowing
      you to leave the resource type and/or name out of other command
      invocations.

      # Pin to pods
      pin pods

      # Pin to a particular pod
      pin pod nginx-1234-asdf

      # Clear the pin
      pin

      # Show this help
      pin -h

      The current pin will be shown in the prompt.
`)

func setContextCommand(sh *kubesh, args []string) error {
	if len(args) > 3 || (len(args) > 1 && args[1] == "-h") {
		fmt.Println("Invalid arguments.")
		fmt.Println(help)

		return nil
	}
	if len(args) == 1 {
		sh.context = []string{}
		fmt.Println("Pin cleared.")
	} else {
		ctxargs := append(sh.context, args[1:]...)
		resources, err := sh.finder.Lookup(ctxargs)
		if err != nil {
			return err
		}
		if len(resources) > 0 {
			res := resources[0]
			if len(ctxargs) == 1 {
				sh.context = []string{res.typeName}
			} else {
				sh.context = []string{res.typeName, res.name}
			}
		}
	}
	sh.rl.SetPrompt(prompt(sh.context))
	return nil
}

func applyContext(context []string, args []string, rootCommand *cobra.Command) ([]string, error) {
	newArgs := []string{}
	newArgs = append(newArgs, args[0])

	if len(context) > 0 {
		subcmd, _, err := rootCommand.Find(args[:1])
		if err != nil {
			return args, err
		}
		cmd := Command{subcmd}

		// poor man's set
		resourceTypes := map[string]struct{}{}
		for _, t := range kubectl.ResourceAliases(cmd.ResourceTypes()) {
			resourceTypes[t] = struct{}{}
		}

		if len(resourceTypes) == 0 {
			t, ok := commandTakesResourceName(args[0])
			if ok && t == context[0] && len(context) > 1 {
				newArgs = append(newArgs, context[1])
			}

			// top is a special snowflake
			if args[0] == "top" && (context[0] == "pods" || context[0] == "nodes") {
				newArgs = append(newArgs, context[0])
			}
		} else if _, ok := resourceTypes[context[0]]; ok {
			nonFlagArgs := cmd.NonFlags(args)
			// the subcommand is an arg, so we'll always have at least one
			switch l := len(nonFlagArgs); {
			case l == 1:
				newArgs = append(newArgs, context...)
			case l == 2:
				// could be a resource type or a resource name
				arg := nonFlagArgs[1]
				if _, ok := resourceTypes[arg]; !ok {
					// treat as a resource name, use the resource type from the context
					newArgs = append(newArgs, context[0])
				}
			}
		}
	}

	return append(newArgs, args[1:]...), nil
}

func commandTakesResourceName(cmd string) (string, bool) {
	switch cmd {
	case "attach", "exec", "logs", "port-forward":
		return "pods", true
	case "cordon", "drain", "uncordon":
		return "nodes", true
	}

	return "", false
}
