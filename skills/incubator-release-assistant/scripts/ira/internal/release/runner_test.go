package release

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerWritesLiveProgressToConsoleAndLog(t *testing.T) {
	t.Setenv("IRA_RUNNER_TEST_HELPER", "1")
	logPath := filepath.Join(t.TempDir(), "command.log")
	var console bytes.Buffer
	runner := Runner{Out: &console, ProgressInterval: 5 * time.Millisecond}
	if err := runner.Run(logPath, "", os.Args[0], "-test.run=^TestRunnerHelperProcess$"); err != nil {
		t.Fatal(err)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{"console": console.String(), "log": string(logBytes)} {
		for _, required := range []string{"[IRA] START", "[IRA] COMMAND", "[IRA] WORKDIR", "[IRA] LOG", "[IRA] PID", "[IRA] RUNNING", "helper output", "[IRA] DONE"} {
			if !strings.Contains(text, required) {
				t.Errorf("%s output is missing %q:\n%s", name, required, text)
			}
		}
	}
}

func TestRunnerHelperProcess(t *testing.T) {
	if os.Getenv("IRA_RUNNER_TEST_HELPER") != "1" {
		return
	}
	time.Sleep(30 * time.Millisecond)
	fmt.Fprintln(os.Stdout, "helper output")
}
