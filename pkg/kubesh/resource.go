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
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type Resourceful struct {
	Factory *cmdutil.Factory // can't extend non-local types
}

type ResourceInfo struct {
	typeName string
	name     string
}

type ResourceFinder interface {
	// takes a type, or type and resource name, returning a slice
	// of ResourceInfo for each record returned by the api
	Lookup(args []string) ([]ResourceInfo, error)
}

func (rf Resourceful) Lookup(args []string) ([]ResourceInfo, error) {
	f := rf.Factory
	cmdNamespace, _, err := f.DefaultNamespace()
	if err != nil {

		return nil, err
	}

	mapper, typer := f.Object()
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		NamespaceParam(cmdNamespace).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return nil, err
	}

	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}

	ret := make([]ResourceInfo, 0, len(infos))
	for _, i := range infos {
		ret = append(ret, ResourceInfo{i.Mapping.Resource, i.Name})
	}

	return ret, nil
}
