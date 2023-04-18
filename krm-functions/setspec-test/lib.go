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

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func SetSpec[T any](obj *fn.KubeObject, newSpec *T) error {
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
