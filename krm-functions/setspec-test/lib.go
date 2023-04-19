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
	"reflect"
	"strings"
	"unsafe"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func SetSpec(obj *fn.KubeObject, newSpec interface{}) error {
	return SetNestedFieldKeepingComments(&obj.SubObject, newSpec, "spec")
}

func SetStatus(obj *fn.KubeObject, newStatus interface{}) error {
	return SetNestedFieldKeepingComments(&obj.SubObject, newStatus, "status")
}

func SetNestedFieldKeepingComments(obj *fn.SubObject, value interface{}, field string) error {
	oldNode := yamlNodeOf(obj.GetMap(field))
	err := obj.SetNestedField(value, field)
	if err != nil {
		return err
	}
	newNode := yamlNodeOf(obj.GetMap(field))

	RestoreFieldOrder(oldNode, newNode)
	return DeepCopyComments(oldNode, newNode)
}

func ShallowCopyComments(src, dst *yaml.Node) {
	dst.HeadComment = src.HeadComment
	dst.LineComment = src.LineComment
	dst.FootComment = src.FootComment
}

func DeepCopyComments(src, dst *yaml.Node) error {
	if src.Kind != dst.Kind {
		return nil
	}
	ShallowCopyComments(src, dst)
	if dst.Kind == yaml.MappingNode {
		children := dst.Content
		if len(children)%2 != 0 {
			panic("unexpected number of children for YAML map")
		}

		for i := 0; i < len(children); i += 2 {
			dstKeyNode := children[i]
			key, ok := asString(dstKeyNode)
			if !ok {
				continue
			}

			j, ok := findKey(src, key)
			if !ok {
				continue
			}
			srcKeyNode, srcValueNode := src.Content[j], src.Content[j+1]
			dstValueNode := children[i+1]
			ShallowCopyComments(srcKeyNode, dstKeyNode)
			err := DeepCopyComments(srcValueNode, dstValueNode)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func RestoreFieldOrder(src, dst *yaml.Node) {
	if (src.Kind != dst.Kind) || (dst.Kind != yaml.MappingNode) {
		return
	}
	if (len(src.Content)%2 != 0) || (len(dst.Content)%2 != 0) {
		panic("unexpected number of children for YAML map")
	}

	nextInDst := 0
	for i := 0; i < len(src.Content); i += 2 {
		key, ok := asString(src.Content[i])
		if !ok {
			continue
		}

		j, ok := findKey(dst, key)
		if !ok {
			continue
		}
		if j != nextInDst {
			dst.Content[j], dst.Content[nextInDst] = dst.Content[nextInDst], dst.Content[j]
			dst.Content[j+1], dst.Content[nextInDst+1] = dst.Content[nextInDst+1], dst.Content[j+1]
		}
		nextInDst += 2

		srcValueNode := children[i+1]
		dstValueNode := dst.Content[nextInDst-1]
		RestoreFieldOrder(srcValueNode, dstValueNode)
	}
}

func asString(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!str" || node.Tag == "") {
		return node.Value, true
	}
	return "", false
}

func findKey(m *yaml.Node, key string) (int, bool) {
	children := m.Content
	if len(children)%2 != 0 {
		panic("unexpected number of children for YAML map")
	}
	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]
		k, ok := asString(keyNode)
		if ok && k == key {
			return i, true
		}
	}
	return 0, false
}

// This is a temporary workaround until YAML comments (and ordering) are made accessible properly via the official SDK API
func yamlNodeOf(obj *fn.SubObject) *yaml.Node {
	internalObj := reflect.ValueOf(*obj).FieldByName("obj")
	nodePtr := internalObj.Elem().FieldByName("node")
	nodePtr = reflect.NewAt(nodePtr.Type(), unsafe.Pointer(nodePtr.UnsafeAddr())).Elem()
	return nodePtr.Interface().(*yaml.Node)
}

func SetSpec1stTry[T any](obj *fn.KubeObject, newSpec *T) error {
	specFieldName := "spec"
	err := func() error {
		if newSpec == nil {
			return fmt.Errorf("the passed-in object must not be nil")
		}
		if obj == nil {
			return fmt.Errorf("the object doesn't exist")
		}
		specObj := obj.GetMap(specFieldName)
		if specObj == nil {
			return fmt.Errorf("missing %q field", specFieldName)
		}

		var oldSpec T
		err := specObj.As(&oldSpec)
		if err != nil {
			return fmt.Errorf("parse error at %q field: %v", specFieldName, err)
		}

		newSpecVal := reflect.ValueOf(newSpec)
		if newSpecVal.Kind() == reflect.Ptr {
			newSpecVal = newSpecVal.Elem()
		}
		if newSpecVal.Kind() != reflect.Struct {
			return fmt.Errorf("unhandled kind %s for %q field", newSpecVal.Kind(), specFieldName)
		}
		oldSpecVal := reflect.ValueOf(oldSpec)
		return setStructFields(specObj, newSpecVal, oldSpecVal)
	}()
	if err != nil {
		return fmt.Errorf("unable to set %q field of K8s resource %v/%v with error: %w", specFieldName, obj.GetKind(), obj.GetName(), err)
	}
	return nil
}

// setStructFields applies the value `newStructVal` to `obj`, but only overwrites the fields
// that are different than `oldStructVal`'s
func setStructFields(obj *fn.SubObject, newStructVal, oldStructVal reflect.Value) error {
	if newStructVal.Type() != oldStructVal.Type() {
		panic("logical error: type mismatch somewhere it shouldn't happen")
	}

	for i, n := 0, newStructVal.NumField(); i < n; i++ {
		fieldName := GetJsonName(newStructVal.Type().Field(i))
		if fieldName == "" {
			continue
		}

		fieldNewVal := newStructVal.Field(i)
		fieldOldVal := oldStructVal.Field(i)

		if reflect.DeepEqual(fieldOldVal.Interface(), fieldNewVal.Interface()) {
			continue
		}
		if fieldNewVal.Kind() == reflect.String {
			// SetNestedField() doesn't handle enums correctly
			obj.SetNestedString(fieldNewVal.String(), fieldName)
			continue
		}
		if (fieldNewVal.Kind() == reflect.Ptr) &&
			!fieldNewVal.IsNil() &&
			!fieldOldVal.IsNil() {
			fieldNewVal = fieldNewVal.Elem()
			fieldOldVal = fieldOldVal.Elem()
		}
		if fieldNewVal.Kind() == reflect.Struct {
			err := setStructFields(obj.GetMap(fieldName), fieldNewVal, fieldOldVal)
			if err != nil {
				return err
			}
			continue
		}
		err := obj.SetNestedField(fieldNewVal.Interface(), fieldName)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetJsonName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("yaml")
	if jsonTag == "" {
		jsonTag = field.Tag.Get("json")
	}

	switch jsonTag {
	case "": // missing
	case "-": // not serialized
		return ""
	default:
		// TODO: handle json:",inline" by recursively calling
		parts := strings.Split(jsonTag, ",")
		return parts[0]
	}
	return ""
}
