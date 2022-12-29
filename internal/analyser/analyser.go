package analyser

import (
	"github.com/driscollos/config/internal/structs"
	"reflect"
	"strings"
)

type Analyser interface {
	Analyse(thing interface{}) []structs.FieldDefinition
}

type analyser struct{}

func (a analyser) Analyse(thing interface{}) []structs.FieldDefinition {
	v := reflect.ValueOf(thing)
	t := reflect.TypeOf(thing)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	definitions := make([]structs.FieldDefinition, 0)
	for i := 0; i < v.NumField(); i++ {
		def := structs.FieldDefinition{
			Name:         v.Type().Field(i).Name,
			Tags:         t.Field(i).Tag,
			DefaultValue: t.Field(i).Tag.Get("default"),
			Type:         v.Field(i).Type().String(),
		}

		switch strings.ToLower(t.Field(i).Tag.Get("required")) {
		case "true", "yes", "1", "on":
			def.Required = true
		}

		if v.Field(i).Kind().String() == "struct" && v.Field(i).Type().String() != "time.Time" {
			def.Type = "struct"
			def.Nested = a.Analyse(v.Field(i).Interface())
		}
		definitions = append(definitions, def)
	}
	return definitions
}
