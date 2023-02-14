// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	durationParser "github.com/driscollos/config/internal/populator/duration-parser"
	floatParser "github.com/driscollos/config/internal/populator/float-parser"
	"github.com/driscollos/config/internal/sourcer"
)

func New() Populator {
	return populator{
		src:            sourcer.New(),
		floatParser:    floatParser.New(),
		durationParser: durationParser.New(),
	}
}
