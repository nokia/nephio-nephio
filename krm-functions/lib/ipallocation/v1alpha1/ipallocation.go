/*
  Copyright 2023 Nephio.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
  */

package v1alpha1

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

type IPAllocation interface {
	ParseKubeObject() (*fn.KubeObject, error)
}

// NewGenerator creates a new generator for the ipallocation
// It expects a raw byte slice as input representing the serialized yaml file
func NewGenerator(meta metav1.ObjectMeta, spec ipamv1alpha1.IPAllocationSpec) IPAllocation {
	return &ipalloc{
		meta: meta,
		spec: spec,
	}
}

type ipalloc struct {
	meta metav1.ObjectMeta
	spec ipamv1alpha1.IPAllocationSpec
}

func (r *ipalloc) ParseKubeObject() (*fn.KubeObject, error) {
	if len(r.meta.Annotations) == 0 {
		r.meta.Annotations = map[string]string{}
	}
	r.meta.Annotations[filters.LocalConfigAnnotation] = "true"
	ipa := &ipamv1alpha1.IPAllocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ipamv1alpha1.GroupVersion.Identifier(),
			Kind:       ipamv1alpha1.IPAllocationKind,
		},
		ObjectMeta: r.meta,
		Spec:       r.spec,
	}
	b, err := yaml.Marshal(ipa)
	if err != nil {
		return nil, err
	}
	return fn.ParseKubeObject(b)
}