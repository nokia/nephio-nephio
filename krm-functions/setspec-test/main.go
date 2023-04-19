/*
 Copyright 2023 The Nephio Authors.

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

package main

import (
	"context"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	nf "github.com/nephio-project/api/nf_requirements/v1alpha1"
	iflib "github.com/nephio-project/nephio/krm-functions/lib/interface/v1alpha1"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ fn.Runner = &SetSpecFn{}

type SetSpecFn struct {
}

// Run is the main function logic.
func (r *SetSpecFn) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	for _, obj := range items.Where(fn.IsGroupVersionKind(nf.InterfaceGroupVersionKind)) {

		if obj.GetName() == "currentsdk" {

			// current SDK's way of handling spec manipulations

			// create a go interface to the Kubernetes resource
			iface := iflib.NewFromKubeObject(obj)
			if iface == nil {
				results.Errorf("something went wrong") // NOTE: no error to give details
				return false
			}

			// read data from the K8s resource one field at a time (separate Get method to all Spec fields)
			_ = iface.GetCNIType()

			// write data to the K8s resource one field at a time (separate Set method to all Spec fields)
			// keep in mind to use DeleteXXX instead of SetXXX if you want to set it to its empty value
			err := iface.SetCNIType(nf.CNITypeIPVLAN)
			if err != nil {
				results.ErrorE(err)
				return false
			}
			err = iface.SetNetworkInstanceName("test")
			if err != nil {
				results.ErrorE(err)
				return false
			}
			err = iface.DeleteAttachmentType()
			if err != nil {
				results.ErrorE(err)
				return false
			}

		} else {

			// Isti's proposal to handle "Spec" field manipulations:

			// read the Interface into the API go struct
			var iface nf.Interface
			err := obj.As(&iface)
			if err != nil {
				results.ErrorE(err)
				return false
			}

			// manipulate the go struct
			iface.Spec.CNIType = nf.CNITypeIPVLAN
			iface.Spec.NetworkInstance.Name = "test"
			iface.Spec.AttachmentType = ""

			// write back changes in "Spec" to the KubeObject, keeping the comments
			err = SetSpec(obj, &iface.Spec)
			if err != nil {
				results.ErrorE(err)
				return false
			}

			// Reason to change:
			//   No need to create an interface for each K8s resource and 3 methods for each resource field,
			//   and keep the interface and the API go struct in sync later

			// Notes:
			// - this proposal is almost equivalent in terms of keeping the comments with the current SDK
			// - both versions will keep all comments in Release 1 supported scenarios
			// - the implementation of SetSpec() is a temporary fix, the proper place to implement this is here:
			//    https://github.com/GoogleContainerTools/kpt-functions-sdk/blob/e8e9cb3c3ae2a19c22f52701d1542cf24e541df0/go/fn/internal/map.go#L138
			//   at the line that reads: "// TODO: Copy comments?"  :))
			//   Unfortunately different CLA (+ my company's policy) prevents me to contribute there
			// - naturally SetSpec() should be generalized and used for Status, as well

		}
	}

	// process IPAllocations for further tests
	for _, obj := range items.Where(fn.IsGroupVersionKind(ipamv1alpha1.IPAllocationGroupVersionKind)) {
		// read the Interface into the API go struct
		var ipalloc ipamv1alpha1.IPAllocation
		err := obj.As(&ipalloc)
		if err != nil {
			results.ErrorE(err)
			return false
		}

		// manipulate the go struct
		if ipalloc.Spec.Selector == nil {
			ipalloc.Spec.Selector = &v1.LabelSelector{
				MatchLabels: map[string]string{
					"new-test-label-1": "data",
				},
			}
		}
		ipalloc.Spec.Selector.MatchLabels["new-test-label-2"] = "foo"
		if len(ipalloc.Spec.Labels) == 0 {
			ipalloc.Spec.Labels = map[string]string{"new-test-label-3": "val"}
		} else {
			for k := range ipalloc.Spec.Labels {
				delete(ipalloc.Spec.Labels, k)
			}
		}
		ipalloc.Spec.Prefix = "" // "delete" Prefix

		// write back changes in "Spec" to the KubeObject, keeping the comments
		err = SetSpec(obj, &ipalloc.Spec)
		if err != nil {
			results.ErrorE(err)
			return false
		}

		// write Status
		ipalloc.Status.AllocatedPrefix = "1.1.1.1"
		ipalloc.Status.Gateway = "1.1.1.255"
		err = SetStatus(obj, &ipalloc.Status)
		if err != nil {
			results.ErrorE(err)
			return false
		}

	}

	return true
}

func main() {
	runner := fn.WithContext(context.Background(), &SetSpecFn{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}
