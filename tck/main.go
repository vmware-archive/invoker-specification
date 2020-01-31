package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/projectriff/invoker-specification/tck/framework"
	"github.com/spf13/cobra"
)

func main() {
	suites := []*framework.Suite{&framework.Request_reply, &framework.Streaming}

	configFile := ""
	config := framework.Config{}
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := toml.DecodeFile(configFile, &config); err != nil {
				return err
			}

			config.Listener = &consoleListener{}

			runner := framework.NewRunner(&config, suites)
			runner.Run()
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&config.FocusedSuites, "focused-suites", "s", nil, "names of suites to focus")
	cmd.Flags().StringSliceVarP(&config.FocusedTests, "focused-tests", "t", nil, "names of tests to focus")
	cmd.Flags().StringVarP(&configFile, "config", "c", "tck.toml", "path to config file")


	//config.FocusedTests = []string {"rr-0001"}
	cmd.Execute()

}

type consoleListener struct {
	suitePadding string
	testsPadding string
}

func (l *consoleListener) AboutToStart(suites []*framework.Suite, tests map[*framework.Suite][]*framework.Testcase) {
	suitesPrefix := 0
	testsPrefix  := 0
	for s, ts := range tests {
		if len(s.Name) > suitesPrefix {
			suitesPrefix = len(s.Name)
		}
		for _, t := range ts {
			if len(t.Name) > testsPrefix {
				testsPrefix = len(t.Name)
			}
		}
	}
	l.suitePadding = fmt.Sprintf("[%%%ds]", suitesPrefix)
	l.testsPadding = fmt.Sprintf("  [%%%ds]", testsPrefix)
}

func (l *consoleListener) SuiteStart(suite *framework.Suite) {
	fmt.Printf(l.suitePadding+" %s\n", suite.Name, suite.Description)
}
func (l *consoleListener) TechnicalError(testcase *framework.Testcase, result interface{}) {
	var red = color.New(color.FgRed).SprintFunc()
	fmt.Printf(l.testsPadding + " %s %s: %s\n", testcase.Name, red("ERROR"), testcase.Description, result)
}
func (l *consoleListener) OptionalFailure(testcase *framework.Testcase, result interface{}) {
	var yellow = color.New(color.FgHiYellow).SprintFunc()
	fmt.Printf(l.testsPadding + " %s %s: %s\n", testcase.Name, yellow(" WARN"), testcase.Description, result)
}
func (l *consoleListener) HardFailure(testcase *framework.Testcase, result interface{}) {
	var red = color.New(color.FgHiRed).SprintFunc()
	fmt.Printf(l.testsPadding + " %s %s: %s\n", testcase.Name, red(" FAIL"), testcase.Description, result)
}
func (l *consoleListener) Pass(testcase *framework.Testcase) {
	var green = color.New(color.FgHiGreen).SprintFunc()
	fmt.Printf(l.testsPadding + " %s %s\n", testcase.Name, green(" PASS"), testcase.Description)
}
