// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	durationParser "github.com/driscollos/config/internal/populator/duration-parser"
	floatParser "github.com/driscollos/config/internal/populator/float-parser"
	"github.com/driscollos/config/internal/sourcer"
)

type Populator interface {
	Populate(dest interface{}) error
}

type populator struct {
	src            sourcer.Sourcer
	floatParser    floatParser.FloatParser
	durationParser durationParser.DurationParser
}

func (p populator) Populate(dest interface{}) error {
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return errors.New(ErrorNotPointer)
	}

	t := reflect.TypeOf(dest).Elem()
	v := reflect.ValueOf(dest).Elem()
	return p.populate(t, v, "")
}

func (p populator) populate(t reflect.Type, v reflect.Value, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		name := strings.Trim(fmt.Sprintf("%s_%s", prefix, t.Field(i).Name), "_")
		if len(ft.Tag.Get("src")) > 0 {
			name = ft.Tag.Get("src")
		}

		value := p.src.Get(name)
		if len(value) < 1 {
			value = ft.Tag.Get("default")
		}

		isRequired := false
		switch strings.ToLower(ft.Tag.Get("required")) {
		case "yes", "1", "true", "on":
			isRequired = true
		}

		if len(value) < 1 && isRequired {
			return errors.New(fmt.Sprintf("missing required value : %s", name))
		}

		switch ft.Type.Kind() {
		case reflect.Map:
			f.Set(reflect.MakeMap(ft.Type))
			for _, key := range p.findKeys(p.src.Get(name)) {
				f.SetMapIndex(reflect.ValueOf(key), reflect.New(reflect.New(f.Type().Elem()).Elem().Type()).Elem())
				if reflect.New(f.Type().Elem()).Elem().Kind() == reflect.Struct {
					inner := reflect.New(reflect.New(f.Type().Elem()).Elem().Type()).Elem()
					if err := p.populate(reflect.New(f.Type().Elem()).Elem().Type(), inner, fmt.Sprintf("%s_%s", name, key)); err != nil {
						return err
					}
					f.SetMapIndex(reflect.ValueOf(key), inner)
				}
			}
		case reflect.Slice:
			bits := strings.Split(value, ",")
			for i, bit := range bits {
				bits[i] = strings.Replace(bit, `"`, "", -1)
			}

			switch f.Type().Elem().Kind() {
			case reflect.Uint8:
				f.Set(reflect.ValueOf([]byte(value)))
			case reflect.String:
				f.Set(reflect.ValueOf(bits))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if ft.Type.Name() == "Duration" {
					durationBits := make([]time.Duration, 0)
					for _, bit := range bits {
						duration, err := p.durationParser.Parse(bit)
						if err == nil {
							durationBits = append(durationBits, duration)
						}
					}
					if isRequired && len(durationBits) < 1 {
						return errors.New(fmt.Sprintf("missing required value : %s", name))
					}
					f.Set(reflect.ValueOf(durationBits))
				} else {
					intBits := make([]int, 0)
					for _, bit := range bits {
						converted, err := strconv.Atoi(bit)
						if err == nil {
							intBits = append(intBits, converted)
						}
					}
					if isRequired && len(intBits) < 1 {
						return errors.New(fmt.Sprintf("missing required value : %s", name))
					}
					f.Set(reflect.ValueOf(intBits))
				}
			case reflect.Float32, reflect.Float64:
				floatBits := make([]float64, 0)
				for _, bit := range bits {
					fVal, err := p.floatParser.Float64(bit)
					if err == nil {
						floatBits = append(floatBits, fVal)
					}
				}
				if isRequired && len(floatBits) < 1 {
					return errors.New(fmt.Sprintf("missing required value : %s", name))
				}
				f.Set(reflect.ValueOf(floatBits))
			case reflect.Bool:
				boolBits := make([]bool, 0)
				for _, bit := range bits {
					if len(bit) < 1 {
						continue
					}
					switch strings.ToLower(bit) {
					case "1", "yes", "true", "on", "y", "t", "ok":
						boolBits = append(boolBits, true)
					default:
						boolBits = append(boolBits, false)
					}
				}
				f.Set(reflect.ValueOf(boolBits))
			case reflect.Struct:
				sliceCount := p.getSliceCount(value)
				if isRequired && sliceCount < 1 {
					return errors.New(fmt.Sprintf("missing required value : %s", name))
				}
				for i := 0; i < sliceCount; i++ {
					inner := reflect.New(reflect.New(f.Type().Elem()).Elem().Type()).Elem()
					if err := p.populate(inner.Type(), inner, fmt.Sprintf("%s_%d", name, i)); err != nil {
						return err
					}
					f.Set(reflect.Append(f, inner))
				}
			}

		case reflect.Chan:
			f.Set(reflect.MakeChan(ft.Type, 0))
		case reflect.Struct:
			if err := p.populate(ft.Type, f, name); err != nil {
				return err
			}
		case reflect.Ptr:
			fv := reflect.New(ft.Type.Elem())
			if err := p.populate(ft.Type.Elem(), fv.Elem(), prefix); err != nil {
				return err
			}
			f.Set(fv)
		case reflect.String:
			f.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ft.Type.Name() == "Duration" {
				duration, err := p.durationParser.Parse(value)
				if err == nil {
					f.SetInt(int64(duration))
				}
			} else {
				converted, err := strconv.Atoi(value)
				if err == nil {
					if converted == 0 && isRequired {
						return errors.New(fmt.Sprintf("missing required value : %s", name))
					}
					f.SetInt(int64(converted))
				}
			}
		case reflect.Float32, reflect.Float64:
			fVal, _ := p.floatParser.Float64(value)
			if fVal == 0 && isRequired {
				return errors.New(fmt.Sprintf("missing required value : %s", name))
			}

			f.SetFloat(fVal)
		case reflect.Bool:
			switch strings.ToLower(value) {
			case "1", "yes", "true", "on", "y", "t", "ok":
				f.SetBool(true)
			default:
				f.SetBool(false)
			}
		default:
		}
	}
	return nil
}

func (p populator) findKeys(src string) []string {
	if len(src) < 1 {
		return nil
	}
	src = fmt.Sprintf("{%s}", src)

	container := make(map[string]interface{})
	json.Unmarshal([]byte(src), &container)
	keys := make([]string, 0)
	for key, _ := range container {
		keys = append(keys, key)
	}
	return keys
}

func (p populator) getSliceCount(raw string) int {
	container := make([]interface{}, 0)
	err := json.Unmarshal([]byte(fmt.Sprintf("[%s]", raw)), &container)
	if err != nil {
		return 0
	}
	return len(container)
}
