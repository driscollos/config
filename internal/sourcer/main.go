// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

import (
	fileReader "github.com/driscollos/config/internal/sourcer/file-reader"
	terminalReader "github.com/driscollos/config/internal/sourcer/terminal-reader"
)

func New() Sourcer {
	s := sourcer{}
	s.readers.file = fileReader.New()
	s.readers.terminal = terminalReader.New()

	s.sources.files = []string{
		"build/config.yml",
		"build/config.json",
		"config/config.yml",
		"config/config.json",
		"config.yml",
		"config.json",
		"env.yml",
		"env.json",
		"config.local.yml",
		"config.local.json",
		"env.local.yml",
		"env.local.json",
	}

	s.sources.useCommandLine = true
	s.sources.useEnvironment = true
	return &s
}
