package sourcer

import (
	"errors"
	"github.com/driscollos/config/internal/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
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
		mockTerminalReader *mocks.MockTerminalReader
		mySourcer sourcer
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockFileReader = mocks.NewMockFileReader(mockController)
		mockTerminalReader = mocks.NewMockTerminalReader(mockController)
		mySourcer = sourcer{}
		mySourcer.sources.useCommandLine = true
		mySourcer.sources.useEnvironment = true
		mySourcer.readers.file = mockFileReader
		mySourcer.readers.terminal = mockTerminalReader
		mySourcer.sources.files = []string{"test.yaml","test.json"}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("Data sourcer", func() {
		When("a yaml file is processed", func() {
			It("should understand the contents of the yaml file correctly", func() {
				mockTerminalReader.EXPECT().Get(gomock.Any()).Return("", errors.New("not_found")).Times(4)
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
				mockTerminalReader.EXPECT().Get(gomock.Any()).Return("", errors.New("not_found")).Times(4)
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

		When("a value exists in a file, an enivornment variable and the terminal", func() {
			It("should prioritise the three sources appropriately", func() {
				mockTerminalReader.EXPECT().Get("Scores_One").Return("1", nil)
				mockTerminalReader.EXPECT().Get("Scores_Two").Return("", errors.New("not_found"))
				mockTerminalReader.EXPECT().Get("Scores_Three").Return("", errors.New("not_found"))
				os.Setenv("Scores_One", "2")
				os.Setenv("Scores_Two", "2")
				mockFileReader.EXPECT().Read("test.yaml").Return([]byte(strings.TrimSpace(`
Scores:
  One: 3
  Two: 3
  Three: 3
				`)), nil)
				mockFileReader.EXPECT().Read("test.json").Return(nil, errors.New("file not found"))
				Expect(mySourcer.Get("Scores_One")).To(Equal("1"))
				Expect(mySourcer.Get("Scores_Two")).To(Equal("2"))
				Expect(mySourcer.Get("Scores_Three")).To(Equal("3"))
			})
		})

		When("a source is specified manually", func() {
			It("should use this over all other sources", func() {
				mockFileReader.EXPECT().Read("override.yaml").Return([]byte(strings.TrimSpace(`
Name: Bob
				`)), nil)
				mySourcer.Source("override.yaml")
				Expect(mySourcer.Get("Name")).To(Equal("Bob"))
			})
		})
	})
})
