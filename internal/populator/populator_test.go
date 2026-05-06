// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package populator

import (
	"testing"
	"time"

	"github.com/driscollos/config/internal/mocks"
	floatParser "github.com/driscollos/config/internal/populator/float-parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
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
		When("a typed nil pointer is provided", func() {
			It("should return an error instead of panicking", func() {
				var myStruct *struct {
					Name string
				}

				err := myPopulator.Populate(myStruct)
				Expect(err).To(MatchError(ErrorNotPointer))
			})
		})
		When("a struct is provided with a string in it", func() {
			It("should populate the field appropriately", func() {
				myStruct := struct {
					Name string
				}{}

				mockSourcer.EXPECT().Get("Name").Return("Bob")

				err := myPopulator.Populate(&myStruct)
				Expect(myStruct.Name).To(Equal("Bob"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("a struct is provided with an int in it", func() {
			It("should populate the field appropriately", func() {
				myStruct := struct {
					Age int
				}{}

				mockSourcer.EXPECT().Get("Age").Return("40")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
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

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
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

					err := myPopulator.Populate(&myStruct)
					Expect(err).ToNot(HaveOccurred())
					Expect(myStruct.Age).To(Equal(float32(0)))
				})
			})
			When("the float is valid", func() {
				It("should populate the struct with the float value", func() {
					myStruct := struct {
						MyAge float32
					}{}

					mockSourcer.EXPECT().Get("MyAge").Return("60.2")
					err := myPopulator.Populate(&myStruct)
					Expect(err).ToNot(HaveOccurred())
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

					err := myPopulator.Populate(&myStruct)
					Expect(err).ToNot(HaveOccurred())
					Expect(myStruct.Age).To(Equal(float64(0)))
				})
			})
			When("the float is valid", func() {
				It("should populate the struct with the float value", func() {
					myStruct := struct {
						Age float64
					}{}

					mockSourcer.EXPECT().Get("Age").Return("40.5")

					err := myPopulator.Populate(&myStruct)
					Expect(err).ToNot(HaveOccurred())
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

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Hobbies).To(Equal([]string{"Travel", "Adventure"}))
			})
		})
		When("a struct is provided with a map of string slices", func() {
			It("should populate each map value from the nested source", func() {
				myStruct := struct {
					BadWords map[string][]string
				}{}

				mockSourcer.EXPECT().Get("BadWords").Return(`{"abbo":["abbot"],"anus":["uranus"]}`)
				mockSourcer.EXPECT().Get("BadWords").Return(`{"abbo":["abbot"],"anus":["uranus"]}`)
				mockSourcer.EXPECT().Get("BadWords_abbo").Return(`["abbot"]`)
				mockSourcer.EXPECT().Get("BadWords_anus").Return(`["uranus"]`)

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.BadWords).To(Equal(map[string][]string{
					"abbo": {"abbot"},
					"anus": {"uranus"},
				}))
			})
		})
		When("a struct is provided with numeric slices", func() {
			It("should populate each slice using the parsed numeric values", func() {
				myStruct := struct {
					Ints   []int
					Uints  []uint
					Floats []float64
				}{}

				mockSourcer.EXPECT().Get("Ints").Return("1,2,3")
				mockSourcer.EXPECT().Get("Uints").Return("4,5,6")
				mockSourcer.EXPECT().Get("Floats").Return("1.5,2.25,3.75")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Ints).To(Equal([]int{1, 2, 3}))
				Expect(myStruct.Uints).To(Equal([]uint{4, 5, 6}))
				Expect(myStruct.Floats).To(Equal([]float64{1.5, 2.25, 3.75}))
			})
			It("should populate numeric slices from JSON arrays", func() {
				myStruct := struct {
					Ints   []int
					Uints  []uint
					Floats []float64
				}{}

				mockSourcer.EXPECT().Get("Ints").Return("[10,21,56]")
				mockSourcer.EXPECT().Get("Uints").Return("[4,5,6]")
				mockSourcer.EXPECT().Get("Floats").Return("[1.5,2.25,3.75]")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Ints).To(Equal([]int{10, 21, 56}))
				Expect(myStruct.Uints).To(Equal([]uint{4, 5, 6}))
				Expect(myStruct.Floats).To(Equal([]float64{1.5, 2.25, 3.75}))
			})
		})
		When("a struct is provided with a map of pointer values", func() {
			It("should populate pointer scalar values without panicking", func() {
				myStruct := struct {
					Names map[string]*string
				}{}

				mockSourcer.EXPECT().Get("Names").Return(`{"primary":"Bob"}`)
				mockSourcer.EXPECT().Get("Names").Return(`{"primary":"Bob"}`)
				mockSourcer.EXPECT().Get("Names_primary").Return("Bob")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Names).To(HaveKey("primary"))
				Expect(myStruct.Names["primary"]).ToNot(BeNil())
				Expect(*myStruct.Names["primary"]).To(Equal("Bob"))
			})

			It("should populate pointer struct values without panicking", func() {
				type endpoint struct {
					Host string
				}
				myStruct := struct {
					Endpoints map[string]*endpoint
				}{}

				mockSourcer.EXPECT().Get("Endpoints").Return(`{"api":{"Host":"api.internal"}}`)
				mockSourcer.EXPECT().Get("Endpoints").Return(`{"api":{"Host":"api.internal"}}`)
				mockSourcer.EXPECT().Get("Endpoints_api").Return("")
				mockSourcer.EXPECT().Get("Endpoints_api_Host").Return("api.internal")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Endpoints).To(HaveKey("api"))
				Expect(myStruct.Endpoints["api"]).ToNot(BeNil())
				Expect(myStruct.Endpoints["api"].Host).To(Equal("api.internal"))
			})
		})
		When("a struct contains a pointer to another struct", func() {
			It("should preserve the field prefix while populating the nested struct", func() {
				type database struct {
					Host string
				}
				myStruct := struct {
					Database *database
				}{}

				mockSourcer.EXPECT().Get("Database").Return("")
				mockSourcer.EXPECT().Get("Database_Host").Return("db.internal")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.Database).ToNot(BeNil())
				Expect(myStruct.Database.Host).To(Equal("db.internal"))
			})
		})
		When("a struct is provided with a pointer to time", func() {
			It("should still populate scalar pointer fields", func() {
				myStruct := struct {
					StartsAt *time.Time
				}{}

				mockSourcer.EXPECT().Get("StartsAt").Return("2026-01-02T03:04:05Z")

				err := myPopulator.Populate(&myStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(myStruct.StartsAt).ToNot(BeNil())
				Expect(myStruct.StartsAt.UTC()).To(Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
			})
		})
	})
})
