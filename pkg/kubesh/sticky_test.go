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

package kubesh_test

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func newGet() *cobra.Command {
	root := cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	get, _, _ := root.Find([]string{"get"})
	return get
}

func TestReuseCommand(t *testing.T) {
	get := newGet()
	get.ParseFlags([]string{"-v", "2"})
	flag := get.Flags().Lookup("v")
	if flag.Value.String() != "2" {
		t.Error("expected 2, got", flag.Value.String())
	}
	if get.Flags().Lookup("show-all").Changed {
		t.Error("Should not be set yet")
	}
	get.ParseFlags([]string{"-a"})
	if !get.Flags().Lookup("show-all").Changed {
		t.Error("Should be set already")
	}
	if "2" != get.Flags().Lookup("v").Value.String() {
		t.Error("Log level should still be 2")
	}
	get = newGet()
	if get.Flags().Lookup("show-all").Changed {
		t.Error("Should not be set yet")
	}
	get.DebugFlags()
}
