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
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type Resourceful struct {
	Factory *cmdutil.Factory // can't extend non-local types
}

type Resource struct {
	Type string
	Name string
}

type ResourceFinder interface {
	// takes a type, or type and resource name, returning a slice
	// of Resources for each record returned by the api
	Lookup(args []string) ([]Resource, error)
}

func (rf Resourceful) Lookup(args []string) ([]Resource, error) {
	f := *rf.Factory
	cmdNamespace, _, err := f.DefaultNamespace()
	if err != nil {

		return nil, err
	}

	mapper, typer := f.Object()
	resourceMapper := &resource.Mapper{
		ObjectTyper:  typer,
		RESTMapper:   mapper,
		ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
		Decoder:      f.Decoder(true),
	}
	r := resource.NewBuilder(mapper, typer, resourceMapper.ClientMapper, resourceMapper.Decoder).
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

	obj, err := resource.AsVersionedObject(infos, true, unversioned.GroupVersion{}, f.JSONEncoder())
	if err != nil {
		return nil, err
	}

	filterFuncs := f.DefaultResourceFilterFunc()
	filterOpts := &kubectl.PrintOptions{ShowAll: false}
	_, items, err := cmdutil.FilterResourceList(obj, filterFuncs, filterOpts)
	if err != nil {
		return nil, err
	}

	ret := make([]Resource, 0, len(items))
	for _, item := range items {
		info, err := resourceMapper.InfoForObject(item, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, Resource{info.Mapping.Resource, info.Name})
	}

	return ret, nil
}
