package sourcer

import (
	"errors"
	"github.com/driscollos/config/internal/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Cron Blacklist Update Handler", func() {
	var (
		mockController *gomock.Controller
		mockFileReader *mocks.MockFileReader
		mySourcer sourcer
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockFileReader = mocks.NewMockFileReader(mockController)
		mySourcer = sourcer{
			fileReader: mockFileReader,
		}
		mySourcer.sources.files = []string{"test.yaml","test.json"}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("Data sourcer", func() {
		When("a yaml file is processed", func() {
			It("should understand the contents of the yaml file correctly", func() {
				mockFileReader.EXPECT().Read("test.yaml").Return([]byte(strings.TrimSpace(`
Name: Bob
Hobbies:
  Sports:
    First: Skating
    Best: Running
Age: 41
				`)), nil)
				mockFileReader.EXPECT().Read("test.json").Return(nil, errors.New("file not found"))
				Expect(mySourcer.Get("Name")).To(Equal("Bob"))
				Expect(mySourcer.Get("Hobbies_Sports_First")).To(Equal("Skating"))
				Expect(mySourcer.Get("Hobbies_Sports_Best")).To(Equal("Running"))
				Expect(mySourcer.Get("Age")).To(Equal("41"))
			})
		})

		When("a json file is processed", func() {
			It("should understand the contents of the json file correctly", func() {
				mockFileReader.EXPECT().Read("test.json").Return([]byte(strings.TrimSpace(`
{"Name": "Bob", "Age": 41, "Hobbies": {"Sports": {"First": "Skating", "Best": "Running"}}}
				`)), nil)
				mockFileReader.EXPECT().Read("test.yaml").Return(nil, errors.New("file not found"))
				Expect(mySourcer.Get("Name")).To(Equal("Bob"))
				Expect(mySourcer.Get("Hobbies_Sports_First")).To(Equal("Skating"))
				Expect(mySourcer.Get("Hobbies_Sports_Best")).To(Equal("Running"))
				Expect(mySourcer.Get("Age")).To(Equal("41"))
			})
		})
	})
})
