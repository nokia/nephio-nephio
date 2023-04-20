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

package parser

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

const (
	// errors
	errKubeObjectNotInitialized = "KubeObject not initialized"
)

type Parser[T1 any] interface {
	// GetKubeObject returns the present kubeObject
	GetKubeObject() *fn.KubeObject
	// GetGoStruct returns a go struct representing the present KRM resource
	GetGoStruct() (T1, error)
	// GetStringValue is a generic utility function that returns a string from
	// a string slice representing the path in the yaml doc
	GetStringValue(fields ...string) string
	// GetIntValue is a generic utility function that returns a int from
	// a string slice representing the path in the yaml doc
	GetIntValue(fields ...string) int
	// GetBoolValue is a generic utility function that returns a bool from
	// a string slice representing the path in the yaml doc
	GetBoolValue(fields ...string) bool
	// GetStringMap is a generic utility function that returns a map[string]string from
	// a string slice representing the path in the yaml doc
	GetStringMap(fields ...string) map[string]string
	// SetNestedString is a generic utility function that sets a string on
	// a string slice representing the path in the yaml doc
	SetNestedString(s string, fields ...string) error
	// SetNestedInt is a generic utility function that sets a int on
	// a string slice representing the path in the yaml doc
	SetNestedInt(s int, fields ...string) error
	// SetNestedBool is a generic utility function that sets a bool on
	// a string slice representing the path in the yaml doc
	SetNestedBool(s bool, fields ...string) error
	// SetNestedMap is a generic utility function that sets a map[string]string on
	// a string slice representing the path in the yaml doc
	SetNestedMap(s map[string]string, fields ...string) error
	// DeleteNestedField is a generic utility function that deletes
	// a string slice representing the path from the yaml doc
	DeleteNestedField(fields ...string) error
}

// NewFromKubeObject creates a new parser interface
// It expects a *fn.KubeObject as input representing the serialized yaml file
func NewFromKubeObject[T1 any](o *fn.KubeObject) Parser[T1] {
	return (*obj[T1])(o)
}

// NewFromYaml creates a new parser interface
// It expects raw byte slice as input representing the serialized yaml file
func NewFromYaml[T1 any](b []byte) (Parser[T1], error) {
	o, err := fn.ParseKubeObject(b)
	if err != nil {
		return nil, err
	}
	return NewFromKubeObject[T1](o), nil
}

// NewFromGoStruct creates a new parser interface
// It expects a go struct representing the interface krm resource
func NewFromGoStruct[T1 any](x any) (Parser[T1], error) {
	o, err := fn.NewFromTypedObject(x)
	if err != nil {
		return nil, err
	}
	return NewFromKubeObject[T1](o), err
}

type obj[T1 any] fn.KubeObject

// GetKubeObject returns the present kubeObject
func (r *obj[T1]) GetKubeObject() *fn.KubeObject {
	return (*fn.KubeObject)(r)
}

// GetGoStruct returns a go struct representing the present KRM resource
func (r *obj[T1]) GetGoStruct() (T1, error) {
	var x T1
	err := r.As(&x)
	return x, err
}

// GetStringValue is a generic utility function that returns a string from
// a string slice representing the path in the yaml doc
func (r *obj[T1]) GetStringValue(fields ...string) string {
	s, _, _ := r.SubObject.NestedString(fields...)
	return s
}

// GetIntValue is a generic utility function that returns a int from
// a string slice representing the path in the yaml doc
func (r *obj[T1]) GetIntValue(fields ...string) int {
	i, _, _ := r.SubObject.NestedInt(fields...)
	return i
}

// GetBoolValue is a generic utility function that returns a bool from
// a string slice representing the path in the yaml doc
func (r *obj[T1]) GetBoolValue(fields ...string) bool {
	b, _, _ := r.SubObject.NestedBool(fields...)
	return b
}

// GetStringMap is a generic utility function that returns a map[string]string from
// a string slice representing the path in the yaml doc
func (r *obj[T1]) GetStringMap(fields ...string) map[string]string {
	m, _, _ := r.SubObject.NestedStringMap(fields...)
	return m
}

// SetNestedString is a generic utility function that sets a string on
// a string slice representing the path in the yaml doc
func (r *obj[T1]) SetNestedString(s string, fields ...string) error {
	return r.SubObject.SetNestedField(s, fields...)
}

// SetNestedInt is a generic utility function that sets a int on
// a string slice representing the path in the yaml doc
func (r *obj[T1]) SetNestedInt(s int, fields ...string) error {
	return r.SubObject.SetNestedInt(s, fields...)
}

// SetNestedBool is a generic utility function that sets a bool on
// a string slice representing the path in the yaml doc
func (r *obj[T1]) SetNestedBool(s bool, fields ...string) error {
	return r.SubObject.SetNestedBool(s, fields...)
}

// SetNestedMap is a generic utility function that sets a map[string]string on
// a string slice representing the path in the yaml doc
func (r *obj[T1]) SetNestedMap(s map[string]string, fields ...string) error {
	return r.SubObject.SetNestedStringMap(s, fields...)
}

// DeleteNestedField is a generic utility function that deletes
// a string slice representing the path from the yaml doc
func (r *obj[T1]) DeleteNestedField(fields ...string) error {
	_, err := r.SubObject.RemoveNestedField(fields...)
	return err
}
