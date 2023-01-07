// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package config

import (
	"testing"

	"github.com/driscollos/config/internal/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Config Unit Tests", func() {
	var (
		mockController *gomock.Controller
		mockSourcer    *mocks.MockSourcer
		myConf         config
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockSourcer = mocks.NewMockSourcer(mockController)
		myConf = config{
			source: mockSourcer,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("Config", func() {
		When("the Bool function is called", func() {
			It("should return the correct value", func() {
				values := map[string]bool{
					"true":         true,
					"TRUE":         true,
					"yes":          true,
					"no":           false,
					"on":           true,
					"off":          false,
					"1":            true,
					"0":            false,
					"randomString": false,
				}

				for text, outcome := range values {
					mockSourcer.EXPECT().Get("Parameter").Return(text)
					Expect(myConf.Bool("Parameter")).To(Equal(outcome))
				}
			})
		})
	})
})
