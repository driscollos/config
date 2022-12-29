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
