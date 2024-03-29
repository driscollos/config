// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
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
	Map          struct {
		KeyType string
		ValType string
		Nested  []FieldDefinition
	}
	Required bool
}
