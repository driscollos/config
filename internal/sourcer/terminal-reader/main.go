// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package terminalReader

func New() TerminalReader {
	t := terminalReader{}
	t.parse()
	return &t
}
