// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package terminalReader

import (
	"errors"
	"os"
	"strings"
)

//go:generate mockgen -destination=../../mocks/mock-terminal-reader.go -package=mocks . TerminalReader
type TerminalReader interface {
	Get(key string) (string, error)
}

type terminalReader struct {
	args map[string]string
}

func (t *terminalReader) Get(key string) (string, error) {
	val, exists := t.args[key]
	if !exists {
		return "", errors.New("not found")
	}
	return val, nil
}

func (t *terminalReader) parse() {
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

	t.args = arguments
}
