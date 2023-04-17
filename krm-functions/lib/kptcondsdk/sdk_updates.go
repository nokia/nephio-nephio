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

	kptv1 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1"
	kptfilelibv1 "github.com/nephio-project/nephio/krm-functions/lib/kptfile/v1"
	corev1 "k8s.io/api/core/v1"
)

// handleUpdate sets the condition and resource based on the action
// action: create/update/delete
// kind: own/for/watch
func (r *sdk) handleUpdate(a action, kind gvkKind, refs []*corev1.ObjectReference, obj *object, status kptv1.ConditionStatus, msg string, ignoreOwnKind bool) {
	// set the condition
	r.setConditionInKptFile(a, kind, refs, status, msg)
	// update resource
	if a == actionDelete {
		obj.obj.SetAnnotation(FnRuntimeDelete, "true")
	}
	// set resource
	if ignoreOwnKind {
		r.setObjectInResourceList(kind, refs, obj)
	} else {
		if obj.ownKind == ResourceKindFull {
			r.setObjectInResourceList(kind, refs, obj)
		}
	}
}

func (r *sdk) deleteConditionInKptFile(kind gvkKind, refs []*corev1.ObjectReference) {
	if !IsRefsValid(refs) {
		return
	}
	forRef := refs[0]
	if len(refs) == 1 {
		// delete condition
		r.kptf.DeleteCondition(kptfilelibv1.GetConditionType(forRef))
		// update the status back in the inventory
		r.inv.delete(&gvkKindCtx{gvkKind: kind}, []corev1.ObjectReference{*forRef})
	} else {
		objRef := refs[1]
		// delete condition
		r.kptf.DeleteCondition(kptfilelibv1.GetConditionType(objRef))
		// update the status back in the inventory
		r.inv.delete(&gvkKindCtx{gvkKind: kind}, []corev1.ObjectReference{*forRef, *objRef})
	}
}

func (r *sdk) setConditionInKptFile(a action, kind gvkKind, refs []*corev1.ObjectReference, status kptv1.ConditionStatus, msg string) {
	if !IsRefsValid(refs) {
		return
	}
	forRef := refs[0]
	if len(refs) == 1 {
		c := kptv1.Condition{
			Type:    kptfilelibv1.GetConditionType(forRef),
			Status:  status,
			Message: fmt.Sprintf("%s %s", a, msg),
		}
		r.kptf.SetConditions(c)
	} else {
		objRef := refs[1]
		c := kptv1.Condition{
			Type:    kptfilelibv1.GetConditionType(objRef),
			Status:  status,
			Reason:  fmt.Sprintf("%s.%s", kptfilelibv1.GetConditionType(&r.cfg.For), forRef.Name),
			Message: fmt.Sprintf("%s %s", a, msg),
		}
		r.kptf.SetConditions(c)
		// update the condition status back in the inventory
		r.inv.set(&gvkKindCtx{gvkKind: kind}, []corev1.ObjectReference{*forRef, *objRef}, &c, false)
	}
}

func (r *sdk) setObjectInResourceList(kind gvkKind, refs []*corev1.ObjectReference, obj *object) {
	if !IsRefsValid(refs) {
		return
	}
	forRef := refs[0]
	if len(refs) == 1 {
		r.rl.SetObject(&obj.obj)
		// update the resource status back in the inventory
		r.inv.set(&gvkKindCtx{gvkKind: kind}, []corev1.ObjectReference{*forRef}, &obj.obj, false)
	} else {
		objRef := refs[1]
		r.rl.SetObject(&obj.obj)
		// update the resource status back in the inventory
		r.inv.set(&gvkKindCtx{gvkKind: kind}, []corev1.ObjectReference{*forRef, *objRef}, &obj.obj, false)
	}
}

func IsRefsValid(refs []*corev1.ObjectReference) bool {
	if len(refs) == 0 || (len(refs) == 1 && refs[0] == nil) || (len(refs) == 2 && refs[0] == nil && refs[1] == nil) {
		return false
	}
	return true
}
