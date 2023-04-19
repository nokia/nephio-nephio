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
	"reflect"
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
	oldNode := yamlNodeOf(obj.UpsertMap(field))
	err := obj.SetNestedField(value, field)
	if err != nil {
		return err
	}
	newNode := yamlNodeOf(obj.GetMap(field))

	RestoreFieldOrder(oldNode, newNode)
	DeepCopyComments(oldNode, newNode)
	return nil
}

func ShallowCopyComments(src, dst *yaml.Node) {
	dst.HeadComment = src.HeadComment
	dst.LineComment = src.LineComment
	dst.FootComment = src.FootComment
}

func DeepCopyComments(src, dst *yaml.Node) {
	if src.Kind != dst.Kind {
		return
	}
	ShallowCopyComments(src, dst)
	if dst.Kind == yaml.MappingNode {
		if (len(src.Content)%2 != 0) || (len(dst.Content)%2 != 0) {
			panic("unexpected number of children for YAML map")
		}
		for i := 0; i < len(dst.Content); i += 2 {
			dstKeyNode := dst.Content[i]
			key, ok := asString(dstKeyNode)
			if !ok {
				continue
			}

			j, ok := findKey(src, key)
			if !ok {
				continue
			}
			srcKeyNode, srcValueNode := src.Content[j], src.Content[j+1]
			dstValueNode := dst.Content[i+1]
			ShallowCopyComments(srcKeyNode, dstKeyNode)
			DeepCopyComments(srcValueNode, dstValueNode)
		}
	}
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

		srcValueNode := src.Content[i+1]
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
