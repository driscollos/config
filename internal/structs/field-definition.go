// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package structs

import "reflect"

type FieldDefinition struct {
	Name         string
	DefaultValue string
	Type         string
	Tags         reflect.StructTag
	Nested       []FieldDefinition
	Required     bool
}
