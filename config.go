// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package config

import (
	"errors"
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

type Config interface {
	Bool(param string) bool
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
