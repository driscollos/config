// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

import (
	"encoding/json"
	"errors"
	"fmt"
	fileReader "github.com/driscollos/config/internal/sourcer/file-reader"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
)

//go:generate mockgen -destination=../mocks/mock-data-sourcer.go -package=mocks . Sourcer
type Sourcer interface {
	Get(path string) string
	Source(path string)
}

type sourcer struct {
	fileReader fileReader.FileReader
	sources    struct {
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
		bytes, err := s.fileReader.Read(file)
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
			fmt.Println("json error:", err.Error())
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
		argVal, exists := s.commandLineArgs()[path]
		if exists {
			return argVal
		}
	}

	if s.sources.useEnvironment {
		if len(os.Getenv(path)) > 0 {
			return os.Getenv(path)
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
	if reflect.TypeOf(retVal).Kind() == reflect.Slice {
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
		if pos == len(bits)-1 {
			if extract == nil || reflect.TypeOf(extract).Kind() != reflect.Map {
				return nil
			}
			return extract.(map[string]interface{})[part]
		}

		if reflect.TypeOf(extract).Kind() != reflect.Map {
			return nil
		}
		extract = extract.(map[string]interface{})[part]
	}
	return nil
}

func (s *sourcer) commandLineArgs() map[string]string {
	bits := strings.Split(strings.Join(os.Args, " "), "--")
	inEscapedArg := false
	escapedArg := ""
	escapedArgName := ""
	arguments := make(map[string]string)

	for _, bit := range bits[1:] {
		if inEscapedArg {
			if !strings.Contains(bit, "]") {
				escapedArg += "--" + bit
				continue
			} else {
				escapedArg += "--" + bit[:strings.Index(bit, "]")]
				arguments[escapedArgName] = escapedArg

				escapedArg = ""
				escapedArgName = ""
				inEscapedArg = false
				continue
			}
		}

		parts := strings.Split(bit, " ")
		if strings.Contains(bit, "[") && !strings.Contains(bit, "]") {
			inEscapedArg = true
			escapedArg = bit[strings.Index(bit, "[")+1:]
			escapedArgName = parts[0]
		} else {
			myVal := strings.Join(parts[1:], " ")
			if len(myVal) < 1 {
				myVal = "1"
			}

			myVal = strings.Trim(strings.TrimSpace(myVal), "[")
			myVal = strings.Trim(myVal, "]")
			arguments[parts[0]] = myVal
		}
	}

	return arguments
}
