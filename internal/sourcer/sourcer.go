// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	fileReader "github.com/driscollos/config/internal/sourcer/file-reader"
	terminalReader "github.com/driscollos/config/internal/sourcer/terminal-reader"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -destination=../mocks/mock-data-sourcer.go -package=mocks . Sourcer
type Sourcer interface {
	Get(path string) string
	Source(path string)
}

type sourcer struct {
	readers struct {
		file     fileReader.FileReader
		terminal terminalReader.TerminalReader
	}
	sources struct {
		files          []string
		useCommandLine bool
		useEnvironment bool
	}
	isSetup bool
	values  []map[string]interface{}
}

func (s *sourcer) setup() error {
	if s.isSetup {
		return nil
	}

	s.values = make([]map[string]interface{}, 0)
	for _, file := range s.sources.files {
		bytes, err := s.readers.file.Read(file)
		if err != nil {
			if len(s.sources.files) == 1 {
				return fmt.Errorf("could not read from source file : %s", file)
			}
			continue
		}
		s.loadFromSource(file, bytes)
	}
	s.isSetup = true
	return nil
}

func (s *sourcer) loadFromSource(filename string, source []byte) error {
	data := make(map[string]interface{})
	bits := strings.Split(filename, ".")
	if len(bits) < 2 {
		return errors.New(ErrorUnknownFileFormat)
	}

	switch bits[len(bits)-1] {
	case "yml", "yaml":
		if err := yaml.Unmarshal(source, &data); err != nil {
			return err
		}
	case "json":
		if err := json.Unmarshal(source, &data); err != nil {
			return err
		}
	default:
		return errors.New(ErrorUnknownFileFormat)
	}

	s.values = append(s.values, data)
	return nil
}

func (s *sourcer) Source(path string) {
	s.sources.useCommandLine = false
	s.sources.useEnvironment = false
	s.sources.files = []string{path}
	s.isSetup = false
}

func (s *sourcer) Get(path string) string {
	s.setup()
	var retVal interface{}

	if s.sources.useCommandLine {
		argVal, err := s.readers.terminal.Get(path)
		if err == nil {
			return argVal
		}
	}

	if s.sources.useEnvironment {
		if len(os.Getenv(strings.Replace(path, " ", "_", -1))) > 0 {
			return os.Getenv(strings.Replace(path, " ", "_", -1))
		}
	}

	for _, source := range s.values {
		val := s.get(source, path)
		if val != nil {
			retVal = val
		}
	}

	if retVal == nil {
		return ""
	}

	switch reflect.TypeOf(retVal).Kind() {
	case reflect.Slice, reflect.Map:
		bytes, _ := json.Marshal(retVal)
		return string(bytes)[1 : len(string(bytes))-1]
	}
	return strings.TrimSpace(fmt.Sprintf("%v", retVal))
}

func (s *sourcer) get(source map[string]interface{}, path string) interface{} {
	s.setup()
	bits := strings.Split(path, "_")
	if len(bits) == 1 {
		data := source[path]
		return data
	}

	extract, exists := source[bits[0]]
	if !exists {
		return nil
	}

	bits = bits[1:]
	for pos, part := range bits {
		if extract == nil {
			continue
		}

		if pos == len(bits)-1 {
			if reflect.TypeOf(extract).Kind() == reflect.Slice {
				partAsInt, err := strconv.Atoi(part)
				if err != nil {
					return nil
				}
				if len(part) < (partAsInt + 1) {
					return nil
				}
				return extract.([]interface{})[partAsInt]
			}
			if extract == nil || reflect.TypeOf(extract).Kind() != reflect.Map {
				return nil
			}
			return extract.(map[string]interface{})[part]
		}

		partAsInt, err := strconv.Atoi(part)
		if err == nil && reflect.TypeOf(extract).Kind() == reflect.Slice {
			if len(extract.([]interface{})) > partAsInt {
				extract = extract.([]interface{})[partAsInt]
				continue
			}
		}

		if reflect.TypeOf(extract).Kind() != reflect.Map {
			return nil
		}
		extract = extract.(map[string]interface{})[part]
	}
	return nil
}
