// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package fileReader

import (
	"os"
)

//go:generate mockgen -destination=../../mocks/mock-file-reader.go -package=mocks . FileReader
type FileReader interface {
	Read(filename string) ([]byte, error)
}

type reader struct{}

func (r reader) Read(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
