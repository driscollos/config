// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package config

import (
	"context"
	"errors"
	"github.com/driscollos/config/internal/populator"
	"github.com/driscollos/config/internal/sourcer"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func New() Config {
	return config{
		source: sourcer.New(),
	}
}

// Config will parse terminal arguments, environment variables or configuration sourced from json or yaml files in order to
// understand the configuration of your application or service. This configuration can be retrieved either by calling the
// access methods (which will attempt to convert the requested value to their respective data type) or by passing a struct
// to Populate - which will populate the matching fields of your configuration struct.
type Config interface {

	// Bool will attempt to convert the parameter whose name matches the param argument into a boolean. The default
	// return value is FALSE
	Bool(param string) bool

	// Date will attempt to convert the parameter whose name matches the param argument into a time.Time value - if the
	// parameter is not known to the Config struct or there is an error with conversion this will be reflected in the
	// error return value
	Date(param, layout string) (time.Time, error)

	// Float will attempt to convert the parameter whose name matches the param argument into a float64 value. The default
	// return value is 0
	Float(param string) float64

	// HotReload will attempt to check for changes any files used as a source of information. If changes are detected, the
	// config library will reload the information from that file. Populate actions with a mix of structs to be repopulated
	// with the new information, and functions without parameters to be called in response to the change eg myfunc() {}
	HotReload(ctx context.Context, actions ...interface{})

	// Int will attempt to convert the parameter whose name matches the param argument into an int value. The default
	// return value is 0
	Int(param string) int

	// Populate will attempt to match the fields in the container (struct) argument to the parameters known to the Config
	// struct. It will populate as many fields as it can, coverting them to the correct types. If there are any errors during
	// population this will be reflected in the error return variable - this includes failing to populate fields which are marked
	// as required:"true" in struct tags
	Populate(container interface{}) error

	// Source explicitly specifies a file to be used as a source of information. Information from higher priority sources such as
	// the terminal or environment variables is still considered
	Source(path string)

	// String will attempt to convert the parameter whose name matches the param argument into a string value. The default
	// return value is ""
	String(param string) string
}

type config struct {
	source sourcer.Sourcer
}

// Bool will attempt to convert the parameter whose name matches the param argument into a boolean. The default
// return value is FALSE
func (c config) Bool(param string) bool {
	switch strings.ToLower(c.source.Get(param)) {
	case "true", "yes", "on", "1":
		return true
	}
	return false
}

// Date will attempt to convert the parameter whose name matches the param argument into a time.Time value - if the
// parameter is not known to the Config struct or there is an error with conversion this will be reflected in the
// error return value
func (c config) Date(param, layout string) (time.Time, error) {
	return time.Parse(layout, c.source.Get(param))
}

// Float will attempt to convert the parameter whose name matches the param argument into a float64 value. The default
// return value is 0
func (c config) Float(param string) float64 {
	val, _ := strconv.ParseFloat(c.source.Get(param), 64)
	return val
}

// HotReload will attempt to check for changes any files used as a source of information. If changes are detected, the
// config library will reload the information from that file. Populate actions with a mix of structs to be repopulated
// with the new information, and functions without parameters to be called in response to the change eg myfunc() {}
func (c config) HotReload(ctx context.Context, actions ...interface{}) {
	structs := make([]interface{}, 0)
	funcs := make([]interface{}, 0)

	for _, action := range actions {
		t := reflect.TypeOf(action)
		if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			structs = append(structs, action)
			continue
		}
		if t.Kind() == reflect.Func && t.NumIn() == 0 {
			funcs = append(funcs, action)
		}
	}

	c.source.HotReload(ctx, func() {
		for _, myStruct := range structs {
			_ = populator.New(c.source).Populate(myStruct)
		}
		for _, myFunc := range funcs {
			if f, ok := myFunc.(func()); ok {
				f()
			}
		}
	})
}

// Int will attempt to convert the parameter whose name matches the param argument into an int value. The default
// return value is 0
func (c config) Int(param string) int {
	val, _ := strconv.Atoi(c.source.Get(param))
	return val
}

// Populate will attempt to match the fields in the container (struct) argument to the parameters known to the Config
// struct. It will populate as many fields as it can, coverting them to the correct types. If there are any errors during
// population this will be reflected in the error return variable - this includes failing to populate fields which are marked
// as required:"true" in struct tags
func (c config) Populate(container interface{}) error {
	if reflect.ValueOf(container).Kind() == reflect.Struct {
		return errors.New("pass a pointer to Populate() instead of a struct i.e. Populate(&myConfig)")
	}
	p := populator.New(c.source)
	return p.Populate(container)
}

// String will attempt to convert the parameter whose name matches the param argument into a string value. The default
// return value is ""
func (c config) String(param string) string {
	return c.source.Get(param)
}

// String will attempt to convert the parameter whose name matches the param argument into a string value. The default
// return value is ""
func (c config) Source(path string) {
	c.source.Source(path)
}
