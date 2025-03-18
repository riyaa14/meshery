package design

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
)

var update = flag.Bool("update", false, "update golden files")

func TestDesignListInt(t *testing.T) {
	utils.SetupContextEnv(t)

	// testContext := utils.NewTestHelper(t)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Not able to get current working directory")
	}
	currDir := filepath.Dir(filename)
	fixturesDir := filepath.Join(currDir, "fixtures")

	// test scenarios for fetching data
	tests := []struct {
		Name             string
		Args             []string
		Token            string
		ExpectedResponse string
		ExpectError      bool
	}{
		{
			Name:             "Fetch Pattern List",
			Args:             []string{"list", "--page", "2"},
			ExpectedResponse: "list.design.output.golden",
			Token:            filepath.Join(fixturesDir, "token.golden"),
			ExpectError:      false,
		},
		{
			Name:             "Fetch Pattern List with Local provider",
			Args:             []string{"list", "--page", "1"},
			ExpectedResponse: "list.design.local.output.golden",
			Token:            filepath.Join(fixturesDir, "local.token.golden"),
			ExpectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			utils.TokenFlag = tt.Token

			testdataDir := filepath.Join(currDir, "testdata")
			golden := utils.NewGoldenFile(t, tt.ExpectedResponse, testdataDir)

			// Grab console prints
			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			_ = utils.SetupMeshkitLoggerTesting(t, false)
			DesignCmd.SetArgs(tt.Args)
			DesignCmd.SetOut(rescueStdout)
			err := DesignCmd.Execute()
			if err != nil {
				// if we're supposed to get an error
				if tt.ExpectError {
					// write it in file
					if *update {
						golden.Write(err.Error())
					}
					expectedResponse := golden.Load()

					utils.Equals(t, expectedResponse, err.Error())
					return
				}
				t.Fatal(err)
			}

			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = rescueStdout

			actualResponse := string(out)

			if *update {
				golden.Write(actualResponse)
			}
			expectedResponse := golden.Load()
			expectedResponse = trimLastNLinesInt(expectedResponse, 2)

			cleanedActualResponse := utils.CleanStringFromHandlePagination(actualResponse)
			cleanedExpectedResponse := utils.CleanStringFromHandlePagination(expectedResponse)

			utils.Log.Debugf("actual", cleanedActualResponse)
			utils.Log.Debugf("expected", cleanedExpectedResponse)

			utils.Equals(t, cleanedExpectedResponse, cleanedActualResponse)
		})
		t.Log("List Design test Passed")
	}
}

func trimLastNLinesInt(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return ""
	}
	return strings.Join(lines[:len(lines)-n], "\n")
}
