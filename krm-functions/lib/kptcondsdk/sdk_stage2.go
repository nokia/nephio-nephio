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

package kptcondsdk

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	kptv1 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1"
	kptfilelibv1 "github.com/nephio-project/nephio/krm-functions/lib/kptfile/v1"
	corev1 "k8s.io/api/core/v1"
)

// generateResource updates or generates the resource when the status is declared ready
// First readiness is validated in 2 steps:
// - global readiness: when key resources are missing
// - per instance readiness: when certain parts of an instance readiness is missing
func (r *sdk) generateResource() {
	fn.Logf("generateResource isReady: %t\n", r.inv.isReady())
	if !r.inv.isReady() {
		// when the overal status is not ready delete all resources
		// TBD if we need to check the delete annotation
		readyMap := r.inv.getReadyMap()
		for _, readyCtx := range readyMap {
			if readyCtx.forObj != nil {
				if len(r.cfg.Owns) == 0 {
					r.rl.DeleteObject(readyCtx.forObj)
				}
			}
		}
		return
	}
	// the overall status is ready, so lets check the readiness map
	readyMap := r.inv.getReadyMap()
	if len(readyMap) == 0 {
		// this is when the global resource is not found
		r.handleGenerateUpdate(corev1.ObjectReference{APIVersion: r.cfg.For.APIVersion, Kind: r.cfg.For.Kind, Name: r.kptf.GetKptFile().Name}, nil, []*fn.KubeObject{})
	}
	for forRef, readyCtx := range readyMap {
		fn.Logf("generateResource readyMap: forRef %v, readyCtx: %v\n", forRef, readyCtx)
		// if the for is not ready delete the object
		if !readyCtx.ready {
			if readyCtx.forObj != nil {
				// TBD if this is the right approach -> avoids deleting interface
				if len(r.cfg.Owns) == 0 {
					r.rl.DeleteObject(readyCtx.forObj)
				}
			}
			continue
		}
		if r.cfg.GenerateResourceFn != nil {
			objs := []*fn.KubeObject{}
			for _, o := range readyCtx.owns {
				objs = append(objs, &o)
			}
			for _, o := range readyCtx.watches {
				objs = append(objs, &o)
			}
			r.handleGenerateUpdate(forRef, readyCtx.forObj, objs)
		}
	}
	// update the kptfile with the latest conditions
	kptfile, err := r.kptf.ParseKubeObject()
	if err != nil {
		fn.Log(err)
		r.rl.AddResult(err, r.rl.GetObjects()[0])
	}
	r.rl.SetObject(kptfile)
}

// handleGenerateUpdate performs the fn/controller callback and handles the response
// by updating the condition and resource in kptfile/resourcelist
func (r *sdk) handleGenerateUpdate(forRef corev1.ObjectReference, forObj *fn.KubeObject, objs []*fn.KubeObject) {
	newObj, err := r.cfg.GenerateResourceFn(forObj, objs)
	if err != nil {
		fn.Log("error generating new resource: %v", err.Error())
		if forObj != nil {
			r.rl.AddResult(err, forObj)
		} else {
			r.rl.AddResult(err, r.rl.GetObjects()[0])
		}
		return
	}
	if newObj == nil {
		fn.Logf("cannot generate resource GenerateResourceFn returned nil, for: %v\n", forRef)
		if forObj != nil {
			r.rl.AddResult(fmt.Errorf("cannot generate resource GenerateResourceFn returned nil, for: %v", forRef), forObj)
		} else {
			r.rl.AddResult(fmt.Errorf("cannot generate resource GenerateResourceFn returned nil, for: %v", forRef), r.rl.GetObjects()[0])
		}
		return
	}
	// set owner reference on the new resource if not having owns
	// as you ste it to yourself
	if len(r.cfg.Owns) == 0 {
		newObj.SetAnnotation(FnRuntimeOwner, kptfilelibv1.GetConditionType(&forRef))
	}
	// add the resource to the kptfile and updates the resource in the resourcelist
	r.handleUpdate(actionUpdate, forGVKKind, []*corev1.ObjectReference{&forRef}, &object{obj: *newObj}, kptv1.ConditionTrue, "done", true)
}
