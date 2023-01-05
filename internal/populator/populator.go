// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"fmt"
	"github.com/driscollos/config/internal/analyser"
	durationParser "github.com/driscollos/config/internal/populator/duration-parser"
	"github.com/driscollos/config/internal/sourcer"
	"github.com/driscollos/config/internal/structs"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Populator interface {
	Populate(container interface{}) error
}

type populator struct {
	analyser       analyser.Analyser
	sourcer        sourcer.Sourcer
	durationParser durationParser.DurationParser
}

func (p populator) Populate(container interface{}) error {
	def := p.analyser.Analyse(container)
	v := reflect.ValueOf(container)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return p.populate("", def, v)
}

func (p populator) populate(path string, def []structs.FieldDefinition, container reflect.Value) error {
	for i, field := range def {
		if field.Type == "struct" {
			if err := p.populate(strings.TrimLeft(fmt.Sprintf("%s_%s", path, field.Name), "_"), field.Nested, container.Field(i)); err != nil {
				return err
			}
			continue
		}
		if err := p.populateField(strings.TrimLeft(fmt.Sprintf("%s_%s", path, field.Name), "_"), field, container.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func (p populator) populateField(path string, def structs.FieldDefinition, container reflect.Value) error {
	val := p.findVal(path)
	if len(val) < 1 || val == "" {
		if len(def.DefaultValue) < 1 && def.Required {
			return fmt.Errorf("field '%s' is required but has no defined or default value", strings.Replace(path, "_", ".", -1))
		}
		val = def.DefaultValue
	}

	switch def.Type {
	case "string":
		container.SetString(val)
	case "int", "int8", "int16", "int32", "int64":
		asInt, _ := strconv.Atoi(val)
		container.SetInt(int64(asInt))
	case "time.Duration":
		duration, err := p.durationParser.Parse(val)
		if err == nil {
			container.SetInt(int64(duration))
		}
	case "time.Time":
		container.Set(reflect.ValueOf(p.time(def.Tags.Get("layout"), val)))
	case "float32":
		fVal, err := strconv.ParseFloat(val, 32)
		if err == nil {
			container.SetFloat(fVal)
		} else {
			container.SetFloat(0)
		}
	case "float64":
		fVal, err := strconv.ParseFloat(val, 64)
		if err == nil {
			container.SetFloat(fVal)
		} else {
			container.SetFloat(0)
		}
	case "bool":
		switch val {
		case "true", "yes", "1", "on":
			container.SetBool(true)
		default:
			container.SetBool(false)
		}
	case "[]string":
		if len(val) > 0 {
			allVals := make([]string, 0)
			for _, subVal := range strings.Split(val, ",") {
				allVals = append(allVals, subVal)
			}
			container.Set(reflect.ValueOf(allVals))
		}
	case "[]int", "[]int8", "[]int16", "[]int32", "[]int64":
		allVals := make([]int, 0)
		for _, subVal := range strings.Split(val, ",") {
			intVal, err := strconv.Atoi(subVal)
			if err == nil {
				allVals = append(allVals, intVal)
			}
		}
		container.Set(reflect.ValueOf(allVals))
	case "map[string]string":
		allVals := make(map[string]string)
		for _, subVal := range strings.Split(val, ",") {
			parts := strings.Split(strings.TrimSpace(subVal), ":")
			if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
				continue
			}
			allVals[parts[0]] = parts[1]
		}
		container.Set(reflect.ValueOf(allVals))
	case "map[string]int":
		allVals := make(map[string]int)
		for _, subVal := range strings.Split(val, ",") {
			parts := strings.Split(strings.TrimSpace(subVal), ":")
			if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
				continue
			}
			intVal, err := strconv.Atoi(parts[1])
			if err == nil {
				allVals[parts[0]] = intVal
			} else {
				allVals[parts[0]] = 0
			}
		}
		container.Set(reflect.ValueOf(allVals))
	case "map[string]int64":
		allVals := make(map[string]int64)
		for _, subVal := range strings.Split(val, ",") {
			parts := strings.Split(strings.TrimSpace(subVal), ":")
			if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
				continue
			}
			intVal, err := strconv.Atoi(parts[1])
			if err == nil {
				allVals[parts[0]] = int64(intVal)
			} else {
				allVals[parts[0]] = 0
			}
		}
		container.Set(reflect.ValueOf(allVals))
	}

	return nil
}

func (p populator) duration(val string) time.Duration {
	fragment := ""
	var duration time.Duration
	for i := 0; i < len(val); i++ {
		bit := val[i : i+1]
		_, err := strconv.Atoi(bit)
		if err != nil {
			if len(fragment) < 1 {
				continue
			}
			fragmentVal, err := strconv.Atoi(fragment)
			if err != nil {
				continue
			}
			switch bit {
			case "M":
				duration += time.Hour * time.Duration(730*fragmentVal)
			case "w":
				duration += time.Hour * time.Duration(168*fragmentVal)
			case "d":
				duration += time.Hour * time.Duration(24*fragmentVal)
			case "h":
				duration += time.Hour * time.Duration(fragmentVal)
			case "m":
				duration += time.Second * time.Duration(60*fragmentVal)
			case "s":
				duration += time.Second * time.Duration(fragmentVal)
			}
			fragment = ""
			continue
		}
		fragment += bit
	}
	return duration
}

func (p populator) time(layout, val string) time.Time {
	myTime, err := time.Parse(layout, val)
	if err != nil {
		return time.Time{}
	}
	return myTime
}

func (p populator) findVal(path string) string {
	return p.sourcer.Get(path)
}
