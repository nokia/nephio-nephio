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
	"fmt"
	"os"
	"reflect"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn/internal"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const ownerRefAnnotationKey = "fnruntime.nephio.org/owner"
const deletionAnnotationKey = "fnruntime.nephio.org/delete"

func OwnedBy(parent *fn.KubeObject) func(*fn.KubeObject) bool {
	ownerRef := fmt.Sprintf("alma, %v", parent.GetId())
	return func(obj *fn.KubeObject) bool {
		return obj.GetAnnotation(ownerRefAnnotationKey) == ownerRef
	}
}

func IsOrphanOwnedby(gvk schema.GroupVersionKind) func(*fn.KubeObject) bool {
	ownerRef := fmt.Sprintf("alma, %v", gvk)
	return func(obj *fn.KubeObject) bool {
		return obj.GetAnnotation(ownerRefAnnotationKey) == ownerRef
	}
}

func FlagForDeletion(obj *fn.KubeObject) {
	obj.SetAnnotation(deletionAnnotationKey, "true")
}

func CreateOrUpdate(owned fn.KubeObjects, newObj interface{}, updateFn func() error) error {
	return nil
}

func SetSpecKeepComments(o *fn.SubObject, val interface{}) error {
	field := "spec"

	err := func() error {
		if val == nil {
			return fmt.Errorf("the passed-in object must not be nil")
		}
		if o == nil {
			return fmt.Errorf("the object doesn't exist")
		}
		kind := reflect.ValueOf(val).Kind()
		if kind == reflect.Ptr {
			kind = reflect.TypeOf(val).Elem().Kind()
		}

		switch kind {
		case reflect.Struct, reflect.Map:
			m, err := internal.TypedObjectToMapVariant(val)
			if err != nil {
				return err
			}
			node := m.Node()
			for _, n := range node.Content {
				fmt.Fprintf(os.Stderr, "++ %v, %v, %v", n.FootComment, n.LineComment, n.HeadComment)
			}
			// node2 := o.GetMap("status")
			return o.GetInternalMap().SetNestedMap(m, field)

		default:
			return fmt.Errorf("unhandled kind %s", kind)
		}
	}()
	if err != nil {
		return fmt.Errorf("unable to set %v at fields %v with error: %w", val, field, err)
	}
	return nil
}
