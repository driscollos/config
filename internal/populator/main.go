// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"github.com/driscollos/config/internal/analyser"
	durationParser "github.com/driscollos/config/internal/populator/duration-parser"
	"github.com/driscollos/config/internal/sourcer"
)

func New(source sourcer.Sourcer) Populator {
	return &populator{
		analyser:       analyser.New(),
		sourcer:        source,
		durationParser: durationParser.New(),
	}
}
