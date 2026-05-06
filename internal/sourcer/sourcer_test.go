// Copyright 2022 John Driscoll (https://github.com/codebyjdd)
// This code is licensed under the MIT license
// Please see LICENSE.md

package sourcer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driscollos/config/internal/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Sourcer Unit Tests", func() {
	var (
		mockController     *gomock.Controller
		mockFileReader     *mocks.MockFileReader
		mockTerminalReader *mocks.MockTerminalReader
		mySourcer          sourcer
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
		mySourcer.sources.files = []string{"test.yaml", "test.json"}
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
				defer os.Unsetenv("Scores_One")
				defer os.Unsetenv("Scores_Two")
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
		When("an environment variable uses the documented normalized key format", func() {
			It("should use the normalized value", func() {
				const path = "Classes.Computer Science.Pupils.0.Name"
				Expect(os.Setenv("CLASSES_COMPUTER_SCIENCE_PUPILS_0_NAME", "Steve")).To(Succeed())
				defer os.Unsetenv("CLASSES_COMPUTER_SCIENCE_PUPILS_0_NAME")

				mySourcer.sources.files = nil
				mockTerminalReader.EXPECT().Get(path).Return("", errors.New("not_found"))

				Expect(mySourcer.Get(path)).To(Equal("Steve"))
			})
		})

		When("a source is specified manually and", func() {
			When("there is data from a variable set on the terminal", func() {
				It("should favour the data from the terminal", func() {
					mockFileReader.EXPECT().Read("override.yaml").Return([]byte("Name: fromFile"), nil)
					mockTerminalReader.EXPECT().Get("Name").Return("fromTerminal", nil)
					Expect(os.Setenv("Name", "fromEnv")).ToNot(HaveOccurred())
					defer os.Unsetenv("Name")
					mySourcer.Source("override.yaml")
					Expect(mySourcer.Get("Name")).To(Equal("fromTerminal"))
				})
			})
			When("there is no data from the terminal but there is an environment variable", func() {
				It("should favour data from the environment variable", func() {
					mockFileReader.EXPECT().Read("override.yaml").Return([]byte("Name: fromFile"), nil)
					mockTerminalReader.EXPECT().Get("Name").Return("", errors.New("not found"))
					Expect(os.Setenv("Name", "fromEnv")).ToNot(HaveOccurred())
					defer os.Unsetenv("Name")
					mySourcer.Source("override.yaml")
					Expect(mySourcer.Get("Name")).To(Equal("fromEnv"))
				})
			})
			When("there is no data from the terminal or the environment but there is something in the file the end user specified", func() {
				It("should use the data from the file the user specified", func() {
					mockFileReader.EXPECT().Read("override.yaml").Return([]byte("Name: fromFile"), nil)
					mockTerminalReader.EXPECT().Get("Name").Return("", errors.New("not found"))
					Expect(os.Setenv("Name", "")).ToNot(HaveOccurred())
					defer os.Unsetenv("Name")
					mySourcer.Source("override.yaml")
					Expect(mySourcer.Get("Name")).To(Equal("fromFile"))
				})
			})
		})

		When("there is only one source file and", func() {
			When("the file reader is unable to read the file", func() {
				It("should return blank when asked to Get a variable", func() {
					mockFileReader.EXPECT().Read("mysource.yml").Return(nil, errors.New("some-error"))
					mySourcer.Source("mysource.yml")
					Expect(mySourcer.Get("Name")).To(Equal(""))
				})
			})
			When("the relevant parser is unable to parse the yaml file", func() {
				It("should return blank when asked to Get a variable", func() {
					mockFileReader.EXPECT().Read("mysource.yml").Return([]byte(`--not-valid--`), nil)
					mySourcer.Source("mysource.yml")
					Expect(mySourcer.Get("Name")).To(Equal(""))
				})
			})
			When("the relevant parser is unable to parse the json file", func() {
				It("should return blank when asked to Get a variable", func() {
					mockFileReader.EXPECT().Read("mysource.json").Return([]byte(`--not-valid--`), nil)
					mySourcer.Source("mysource.json")
					Expect(mySourcer.Get("Name")).To(Equal(""))
				})
			})
			When("the filename of the source file has an unknown extension", func() {
				It("should return blank when asked to Get a variable", func() {
					mockFileReader.EXPECT().Read("mysource.unknown").Return([]byte(`--not-valid--`), nil)
					mySourcer.Source("mysource.unknown")
					Expect(mySourcer.Get("Name")).To(Equal(""))
				})
			})
			When("the filename of the source file has no extension", func() {
				It("should return blank when asked to Get a variable", func() {
					mockFileReader.EXPECT().Read("mysource").Return([]byte(`--not-valid--`), nil)
					mySourcer.Source("mysource")
					Expect(mySourcer.Get("Name")).To(Equal(""))
				})
			})
		})

		When("hot reload sees a changed file that becomes invalid", func() {
			It("should keep the last good values and avoid firing callbacks", func() {
				dir, err := os.MkdirTemp("/tmp", "config-hot-reload-*")
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(dir)

				file := filepath.Join(dir, "config.yml")
				Expect(os.WriteFile(file, []byte("Name: before\n"), 0o600)).To(Succeed())

				reloadSourcer := New().(*sourcer)
				reloadSourcer.Source(file)
				Expect(reloadSourcer.Get("Name")).To(Equal("before"))

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				callbacks := 0
				reloadSourcer.HotReload(ctx, func() {
					callbacks++
				})

				time.Sleep(1100 * time.Millisecond)
				Expect(os.WriteFile(file, []byte(":\n"), 0o600)).To(Succeed())

				time.Sleep(1500 * time.Millisecond)
				Expect(reloadSourcer.Get("Name")).To(Equal("before"))
				Expect(callbacks).To(Equal(0))
			})
		})
	})
})
