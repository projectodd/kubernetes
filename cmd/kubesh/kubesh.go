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
	"os"

	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubesh"
)

func main() {
	cmdutil.BehaviorOnFatal(func(msg string, code int) {
		fmt.Println(msg)
		panic("kubectl")
	})

	kubesh := kubesh.NewKubesh()

	// If args are passed, just run that kubectl command directly
	progArgs := os.Args[1:]
	if len(progArgs) > 0 {
		kubesh.Execute(progArgs)
		return
	}

	fmt.Println("kubesh is an interactive interface to kubectl")
	fmt.Println("<TAB> should complete most commands and resources")
	fmt.Println("For options/flags, tab complete a dash, '--<TAB>'")
	fmt.Println("Use GNU readline key bindings for editing and history")

	kubesh.Run()
}
