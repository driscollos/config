package populator

import (
	"github.com/driscollos/config/internal/analyser"
	"github.com/driscollos/config/internal/sourcer"
)

func New(source sourcer.Sourcer) Populator {
	return populator{
		analyser: analyser.New(),
		sourcer:  source,
	}
}
