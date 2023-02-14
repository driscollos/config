// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"testing"

	"github.com/driscollos/config/internal/mocks"
	floatParser "github.com/driscollos/config/internal/populator/float-parser"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Unit tests", func() {
	var (
		mockController     *gomock.Controller
		mockSourcer        *mocks.MockSourcer
		mockDurationParser *mocks.MockDurationParser
		myPopulator        populator
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockSourcer = mocks.NewMockSourcer(mockController)
		mockDurationParser = mocks.NewMockDurationParser(mockController)
		myPopulator = populator{
			floatParser:    floatParser.New(),
			src:            mockSourcer,
			durationParser: mockDurationParser,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("Populating a struct", func() {
		When("a struct is provided with a string in it", func() {
			It("should populate the field appropriately", func() {
				myStruct := struct {
					Name string
				}{}

				mockSourcer.EXPECT().Get("Name").Return("Bob")

				myPopulator.Populate(&myStruct)
				Expect(myStruct.Name).To(Equal("Bob"))
			})
		})
		When("a struct is provided with an int in it", func() {
			It("should populate the field appropriately", func() {
				myStruct := struct {
					Age int
				}{}

				mockSourcer.EXPECT().Get("Age").Return("40")

				myPopulator.Populate(&myStruct)
				Expect(myStruct.Age).To(Equal(40))
			})
		})
		When("a struct is provided with a bool in it", func() {
			It("should populate the field appropriately", func() {
				myStruct := struct {
					BoolOne   bool
					BoolTwo   bool
					BoolThree bool
					BoolFour  bool
				}{}

				mockSourcer.EXPECT().Get(gomock.Any()).Return("true")
				mockSourcer.EXPECT().Get(gomock.Any()).Return("on")
				mockSourcer.EXPECT().Get(gomock.Any()).Return("yes")
				mockSourcer.EXPECT().Get(gomock.Any()).Return("1")

				myPopulator.Populate(&myStruct)
				Expect(myStruct.BoolOne).To(BeTrue())
				Expect(myStruct.BoolTwo).To(BeTrue())
				Expect(myStruct.BoolThree).To(BeTrue())
				Expect(myStruct.BoolFour).To(BeTrue())
			})
		})
		When("a struct is provided with a float32 in it", func() {
			When("the float is invalid", func() {
				It("should populate the struct with a zero", func() {
					myStruct := struct {
						Age float32
					}{}

					mockSourcer.EXPECT().Get("Age").Return("--invalid--")

					myPopulator.Populate(&myStruct)
					Expect(myStruct.Age).To(Equal(float32(0)))
				})
			})
			When("the float is valid", func() {
				It("should populate the struct with the float value", func() {
					myStruct := struct {
						MyAge float32
					}{}

					mockSourcer.EXPECT().Get("MyAge").Return("60.2")
					myPopulator.Populate(&myStruct)
					Expect(myStruct.MyAge).To(Equal(float32(60.2)))
				})
			})
		})
		When("a struct is provided with a float64 in it", func() {
			When("the float is invalid", func() {
				It("should populate the struct with a zero", func() {
					myStruct := struct {
						Age float64
					}{}

					mockSourcer.EXPECT().Get("Age").Return("--invalid--")

					myPopulator.Populate(&myStruct)
					Expect(myStruct.Age).To(Equal(float64(0)))
				})
			})
			When("the float is valid", func() {
				It("should populate the struct with the float value", func() {
					myStruct := struct {
						Age float64
					}{}

					mockSourcer.EXPECT().Get("Age").Return("40.5")

					myPopulator.Populate(&myStruct)
					Expect(myStruct.Age).To(Equal(40.5))
				})
			})
		})
		When("a struct is provided with a slice of strings inside", func() {
			It("should separate the value by comma and populate", func() {
				myStruct := struct {
					Hobbies []string
				}{}

				mockSourcer.EXPECT().Get("Hobbies").Return("Travel,Adventure")

				myPopulator.Populate(&myStruct)
				Expect(myStruct.Hobbies).To(Equal([]string{"Travel", "Adventure"}))
			})
		})
	})
})
