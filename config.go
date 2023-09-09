// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package config

import (
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

	// Bool will attempt to convert the parameter whose name matches the param argument into a boolean. By default this
	// function will return FALSE
	Bool(param string) bool

	// Date will attempt to convert the parameter whose name matches the param argument into a time.Time value - if the
	// parameter is not known to the Config struct or there is an error with conversion this will be reflected in the
	// error return value
	Date(param, layout string) (time.Time, error)

	Float(param string) float64
	Int(param string) int
	Populate(container interface{}) error
	Source(path string)
	String(param string) string
}

type config struct {
	source sourcer.Sourcer
}

func (c config) Bool(param string) bool {
	switch strings.ToLower(c.source.Get(param)) {
	case "true", "yes", "on", "1":
		return true
	}
	return false
}

func (c config) Date(param, layout string) (time.Time, error) {
	return time.Parse(layout, c.source.Get(param))
}

func (c config) Float(param string) float64 {
	val, _ := strconv.ParseFloat(c.source.Get(param), 64)
	return val
}

func (c config) Int(param string) int {
	val, _ := strconv.Atoi(c.source.Get(param))
	return val
}

func (c config) Populate(container interface{}) error {
	if reflect.ValueOf(container).Kind() == reflect.Struct {
		return errors.New("pass a pointer to Populate() instead of a struct i.e. Populate(&myConfig)")
	}
	p := populator.New(c.source)
	return p.Populate(container)
}

func (c config) String(param string) string {
	return c.source.Get(param)
}

func (c config) Source(path string) {
	c.source.Source(path)
}
