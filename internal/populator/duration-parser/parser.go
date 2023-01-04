// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package durationParser

import (
	"strconv"
	"strings"
	"time"
)

//go:generate mockgen -destination=../../mocks/mock-duration-parser.go -package=mocks . DurationParser
type DurationParser interface {
	Parse(duration string) (time.Duration, error)
}

type parser struct{}

func (p parser) Parse(durationStr string) (time.Duration, error) {
	var section string
	var number int
	var descriptor string
	var duration time.Duration
	var err error
	mode := "number"

	for x := 0; x < len(durationStr); x++ {
		bit := durationStr[x : x+1]
		_, err = strconv.Atoi(bit)
		switch err {
		case nil:
			switch mode {
			case "number":
				section += bit
			case "descriptor":
				duration += p.analyseSection(number, descriptor)
				section = bit
				descriptor = ""
				number = 0
				mode = "number"
			}
		default:
			switch mode {
			case "number":
				number, err = strconv.Atoi(section)
				if err != nil || number < 1 {
					section = bit
					descriptor = ""
					number = 0
					mode = "number"
					continue
				}
				mode = "descriptor"
				descriptor = bit
			case "descriptor":
				descriptor += bit
			}
		}
	}

	switch mode {
	case "descriptor":
		duration += p.analyseSection(number, descriptor)
	}
	return duration, nil
}

func (p parser) analyseSection(number int, descriptor string) time.Duration {
	if number < 1 {
		return 0
	}

	switch strings.TrimSpace(strings.ReplaceAll(descriptor, ",", "")) {
	case "d", "day", "days":
		return (time.Hour * 24) * time.Duration(number)
	case "h", "hr", "hour", "hrs", "hours":
		return time.Hour * time.Duration(number)
	case "m", "min", "minute", "mins", "minutes":
		return time.Minute * time.Duration(number)
	case "s", "sec", "second", "secs", "seconds":
		return time.Second * time.Duration(number)
	}
	return 0
}
