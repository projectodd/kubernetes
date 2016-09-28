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
	"strings"

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
      pin clear

      The current pin will be shown in the prompt.
`)

func setContextCommand(sh *kubesh, args []string) error {
	if len(args) == 1 || len(args) > 3 {
		fmt.Println("Usage: " + args[0] + " (-h | clear | [TYPE [NAME]])")

		//TODO: return an error?
		return nil
	}

	switch arg := args[1]; {
	case arg == "clear":
		sh.context = []string{}
		sh.rl.SetPrompt(prompt(sh.context))

		return nil

	case arg == "-h":
		fmt.Println(help)

		return nil
	}

	resources, err := sh.finder.Lookup(args[1:])
	if err != nil {

		return err
	}

	typeOnly := len(args) == 2

	if len(resources) > 0 {
		res := resources[0]
		if typeOnly {
			sh.context = []string{res.typeName}
		} else {
			sh.context = []string{res.typeName, res.name}
		}
	}

	sh.rl.SetPrompt(prompt(sh.context))

	return nil
}

func prompt(context []string) string {
	return strings.Join(context, ":") + "> "
}

func applyContext(context []string, args []string, rootCommand *cobra.Command) ([]string, error) {
	if len(context) > 0 {
		subcmd, _, err := rootCommand.Find(args[:1])
		if err != nil {

			return args, err
		}

		// poor man's set
		resourceTypes := map[string]struct{}{}
		for _, t := range kubectl.ResourceAliases(ResourceTypes(subcmd)) {
			resourceTypes[t] = struct{}{}
		}
		newArgs := []string{}
		newArgs = append(newArgs, args[0])

		if _, ok := resourceTypes[context[0]]; ok {
			err := subcmd.ParseFlags(args)
			if err != nil {

				return args, err
			}

			nonFlagArgs := subcmd.Flags().Args()
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

			return append(newArgs, args[1:]...), nil
		}
	}

	return args, nil
}
