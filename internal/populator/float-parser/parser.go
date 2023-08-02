// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package floatParser

import (
	"strconv"
	"strings"
)

type FloatParser interface {
	Float32(val string) (float32, error)
	Float64(val string) (float64, error)
}

type parser struct{}

func (p parser) Float32(val string) (float32, error) {
	asFloat, err := p.Float64(val)
	if err != nil {
		return 0.0, err
	}
	return float32(asFloat), nil
}

func (p parser) Float64(val string) (float64, error) {
	asInt, err := strconv.Atoi(strings.Replace(val, ".", "", -1))
	if err != nil {
		return 0.0, err
	}

	bits := strings.Split(val, ".")
	if len(bits) < 2 {
		return float64(asInt), nil
	}

	asFloat := float64(asInt)
	for x := 0; x < len(bits[1]); x++ {
		asFloat = asFloat / 10
	}
	return asFloat, nil
}
