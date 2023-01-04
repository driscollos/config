// Copyright 2022 John Driscoll (https://github.com/jddcode)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"github.com/driscollos/config/internal/mocks"
	"github.com/driscollos/config/internal/structs"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Cron Blacklist Update Handler", func() {
	var (
		mockController     *gomock.Controller
		mockAnalyser       *mocks.MockAnalyser
		mockSourcer        *mocks.MockSourcer
		mockDurationParser *mocks.MockDurationParser
		myPopulator        populator
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyser = mocks.NewMockAnalyser(mockController)
		mockSourcer = mocks.NewMockSourcer(mockController)
		mockDurationParser = mocks.NewMockDurationParser(mockController)
		myPopulator = populator{
			analyser:       mockAnalyser,
			sourcer:        mockSourcer,
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

				fieldDefs := []structs.FieldDefinition{
					{
						Name: "Name",
						Type: "string",
					},
				}
				mockAnalyser.EXPECT().Analyse(gomock.Any()).Return(fieldDefs)
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

				fieldDefs := []structs.FieldDefinition{
					{
						Name: "Age",
						Type: "int",
					},
				}
				mockAnalyser.EXPECT().Analyse(gomock.Any()).Return(fieldDefs)
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
					BoolFive  bool
				}{}

				fieldDefs := []structs.FieldDefinition{
					{
						Name: "BoolOne",
						Type: "bool",
					},
					{
						Name: "BoolTwo",
						Type: "bool",
					},
					{
						Name: "BoolThree",
						Type: "bool",
					},
					{
						Name: "BoolFour",
						Type: "bool",
					},
				}
				mockAnalyser.EXPECT().Analyse(gomock.Any()).Return(fieldDefs)
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
	})
})
