// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
	if dest == nil {
		return errors.New(ErrorNotPointer)
	}
	rt := reflect.TypeOf(dest)
	if rt.Kind() != reflect.Ptr {
		return errors.New(ErrorNotPointer)
	}
	rv := reflect.ValueOf(dest).Elem()
	return p.populate(rt.Elem(), rv, "")
}

func fieldPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func (p populator) getName(ft reflect.StructField, prefix string, t reflect.Type) string {
	name := strings.Trim(fmt.Sprintf("%s_%s", prefix, ft.Name), "_")
	if tag := ft.Tag.Get("src"); tag != "" {
		name = tag
	}
	return name
}

func parseRequired(tag string) bool {
	switch strings.ToLower(tag) {
	case "yes", "1", "true", "on":
		return true
	default:
		return false
	}
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "yes", "true", "on", "y", "t", "ok":
		return true
	default:
		return false
	}
}

func (p populator) maybeDecodeBase64(value, mode string) string {
	switch strings.ToLower(mode) {
	case "true":
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return ""
		}
		return string(decoded)
	case "optional":
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err == nil {
			return string(decoded)
		}
		return value
	default:
		return value
	}
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	z := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), z.Interface())
}

func elemType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

func newOf(t reflect.Type) reflect.Value {
	return reflect.New(elemType(t)).Elem()
}

func (p populator) setScalarFromString(f reflect.Value, ft reflect.StructField, value string, required bool, name string) error {
	switch f.Kind() {
	case reflect.String:
		f.SetString(value)
		if required && value == "" {
			return fmt.Errorf(ErrorMissingRequiredValue, name)
		}
	case reflect.Bool:
		f.SetBool(parseBool(value))
		if required && !parseBool(value) && strings.TrimSpace(value) == "" {
			return fmt.Errorf(ErrorMissingRequiredValue, name)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if ft.Type.Name() == "Duration" {
			if value == "" {
				if required {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				return nil
			}
			dur, err := p.durationParser.Parse(value)
			if err != nil {
				if required {
					return fmt.Errorf("%s: invalid duration %q: %w", name, value, err)
				}
				return nil
			}
			f.SetInt(int64(dur))
			return nil
		}
		if strings.TrimSpace(value) == "" {
			if required {
				return fmt.Errorf(ErrorMissingRequiredValue, name)
			}
			return nil
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			if required {
				return fmt.Errorf("%s: invalid int %q: %w", name, value, err)
			}
			return nil
		}
		f.SetInt(i)
		if required && f.Int() == 0 {
			return fmt.Errorf(ErrorMissingRequiredValue, name)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if strings.TrimSpace(value) == "" {
			if required {
				return fmt.Errorf(ErrorMissingRequiredValue, name)
			}
			return nil
		}
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			if required {
				return fmt.Errorf("%s: invalid uint %q: %w", name, value, err)
			}
			return nil
		}
		f.SetUint(u)
		if required && f.Uint() == 0 {
			return fmt.Errorf(ErrorMissingRequiredValue, name)
		}
	case reflect.Float32, reflect.Float64:
		if strings.TrimSpace(value) == "" {
			if required {
				return fmt.Errorf(ErrorMissingRequiredValue, name)
			}
			return nil
		}
		fv, err := p.floatParser.Float64(value)
		if err != nil {
			if required {
				return fmt.Errorf("%s: invalid float %q: %w", name, value, err)
			}
			return nil
		}
		f.SetFloat(fv)
		if required && f.Float() == 0 {
			return fmt.Errorf(ErrorMissingRequiredValue, name)
		}
	default:
	}
	return nil
}

func parseCSV(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, bit := range raw {
		bit = strings.TrimSpace(strings.Trim(bit, `"`))
		if bit == "" {
			continue
		}
		out = append(out, bit)
	}
	return out
}

func (p populator) parseSliceStrings(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var arr []string
	if json.Unmarshal([]byte(raw), &arr) == nil {
		return arr, nil
	}
	return parseCSV(raw), nil
}

func (p populator) populate(t reflect.Type, v reflect.Value, prefix string) error {
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		f := v.Field(i)

		if ft.PkgPath != "" {
			continue
		}
		if !f.CanSet() {
			continue
		}

		name := p.getName(ft, prefix, t)
		value := p.src.Get(name)
		if value == "" {
			value = ft.Tag.Get("default")
		}

		value = p.maybeDecodeBase64(value, ft.Tag.Get("base64"))
		required := parseRequired(ft.Tag.Get("required"))

		switch f.Kind() {
		case reflect.Map:
			f.Set(reflect.MakeMap(f.Type()))
			keys, err := p.findKeys(p.src.Get(name))
			if err != nil {
				if required {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				continue
			}
			elemT := f.Type().Elem()
			for _, k := range keys {
				keyV := reflect.ValueOf(k).Convert(f.Type().Key())
				elemV := newOf(elemT)

				if elemV.Kind() == reflect.Struct {
					if err := p.populate(elemT, elemV, fmt.Sprintf("%s_%s", name, k)); err != nil {
						return err
					}
				}
				f.SetMapIndex(keyV, elemV)
			}

		case reflect.Slice:
			if f.Type().Elem().Kind() == reflect.Uint8 {
				f.SetBytes([]byte(value))
				if required && len(value) == 0 {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				continue
			}

			if f.Type().Elem().Kind() == reflect.Struct {
				count := p.getSliceCount(value)
				if required && count < 1 {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				for idx := 0; idx < count; idx++ {
					elem := newOf(f.Type().Elem())
					if err := p.populate(f.Type().Elem(), elem, fmt.Sprintf("%s_%d", name, idx)); err != nil {
						return err
					}
					f.Set(reflect.Append(f, elem))
				}
				continue
			}

			strs, _ := p.parseSliceStrings(value)
			if required && len(strs) == 0 {
				return fmt.Errorf(ErrorMissingRequiredValue, name)
			}

			switch f.Type().Elem().Kind() {
			case reflect.String:
				rv := reflect.MakeSlice(f.Type(), 0, len(strs))
				for _, s := range strs {
					rv = reflect.Append(rv, reflect.ValueOf(s))
				}
				f.Set(rv)

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if f.Type().Elem().Name() == "Duration" {
					out := reflect.MakeSlice(f.Type(), 0, len(strs))
					for _, s := range strs {
						if s == "" {
							continue
						}
						d, err := p.durationParser.Parse(s)
						if err == nil {
							out = reflect.Append(out, reflect.ValueOf(d))
						}
					}
					if required && out.Len() == 0 {
						return fmt.Errorf(ErrorMissingRequiredValue, name)
					}
					f.Set(out)
				} else {
					out := reflect.MakeSlice(f.Type(), 0, len(strs))
					for _, s := range strs {
						i, err := strconv.ParseInt(s, 10, 64)
						if err == nil {
							out = reflect.Append(out, reflect.New(f.Type().Elem()).Elem().Convert(reflect.TypeOf(i)))
						}
					}
					if required && out.Len() == 0 {
						return fmt.Errorf(ErrorMissingRequiredValue, name)
					}
					f.Set(out)
				}

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				out := reflect.MakeSlice(f.Type(), 0, len(strs))
				for _, s := range strs {
					u, err := strconv.ParseUint(s, 10, 64)
					if err == nil {
						out = reflect.Append(out, reflect.New(f.Type().Elem()).Elem().Convert(reflect.TypeOf(u)))
					}
				}
				if required && out.Len() == 0 {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				f.Set(out)

			case reflect.Float32, reflect.Float64:
				out := reflect.MakeSlice(f.Type(), 0, len(strs))
				for _, s := range strs {
					fv, err := p.floatParser.Float64(s)
					if err == nil {
						out = reflect.Append(out, reflect.New(f.Type().Elem()).Elem().Convert(reflect.TypeOf(fv)))
					}
				}
				if required && out.Len() == 0 {
					return fmt.Errorf(ErrorMissingRequiredValue, name)
				}
				f.Set(out)

			case reflect.Bool:
				out := reflect.MakeSlice(f.Type(), 0, len(strs))
				for _, s := range strs {
					out = reflect.Append(out, reflect.ValueOf(parseBool(s)))
				}
				f.Set(out)
			}

		case reflect.Struct:
			if err := p.populate(f.Type(), f, name); err != nil {
				return err
			}

		case reflect.Ptr:
			elemT := f.Type().Elem()
			elemV := reflect.New(elemT).Elem()
			if elemV.Kind() == reflect.Struct {
				if err := p.populate(elemT, elemV, prefix); err != nil {
					return err
				}
				if !isZero(elemV) || required {
					ptr := reflect.New(elemT)
					ptr.Elem().Set(elemV)
					f.Set(ptr)
				}
			} else {
				if value != "" || required {
					ptr := reflect.New(elemT)
					if err := p.setScalarFromString(ptr.Elem(), ft, value, required, name); err != nil {
						return err
					}
					f.Set(ptr)
				}
			}

		default:
			if err := p.setScalarFromString(f, ft, value, required, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p populator) findKeys(src string) ([]string, error) {
	if strings.TrimSpace(src) == "" {
		return nil, errors.New(ErrorSourceIsBlank)
	}

	raw := strings.TrimSpace(src)
	if !strings.HasPrefix(raw, "{") {
		raw = "{" + raw + "}"
	}

	container := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &container); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(container))
	for k := range container {
		keys = append(keys, k)
	}
	return keys, nil
}

func (p populator) getSliceCount(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	var arr []interface{}
	if json.Unmarshal([]byte(raw), &arr) == nil {
		return len(arr)
	}
	return len(parseCSV(raw))
}
