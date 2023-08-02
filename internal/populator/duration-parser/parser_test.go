// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package durationParser

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Duration string parser", func() {
	var (
		mockController *gomock.Controller
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("sample strings", func() {
		When("various forms are used", func() {
			It("should parse the string correctly", func() {
				myParser := parser{}
				for key, val := range map[string]int64{
					"1s":            1000000000,
					"1 sec":         1000000000,
					"1sec":          1000000000,
					"1 second":      1000000000,
					"1second":       1000000000,
					"1m":            60000000000,
					"1 min":         60000000000,
					"1min":          60000000000,
					"1 minute":      60000000000,
					"1minute":       60000000000,
					"1h":            3600000000000,
					"1 hour":        3600000000000,
					"1d":            86400000000000,
					"1 day":         86400000000000,
					"1s1h":          3601000000000,
					"1s,1h":         3601000000000,
					"1sec,1hour":    3601000000000,
					"1 sec, 1 hour": 3601000000000,
					"1 hour 1 day":  90000000000000,
					"2 hours 1 day": 93600000000000,
					"1 week":        604800000000000,
					"2 weeks":       1209600000000000,
				} {
					Expect(myParser.Parse(key)).To(Equal(time.Duration(val)))
				}
			})
		})
	})
})
