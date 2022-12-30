// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

func New() Sourcer {
	s := sourcer{}
	s.sources.files = []string{"build/config.yml", "config/config.yml", "config.yml", "env.yml", "config.local.yml", "env.local.yml"}
	s.sources.useCommandLine = true
	s.sources.useEnvironment = true
	return &s
}
