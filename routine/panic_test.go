package routine

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestPanic(t *testing.T) {
	suite.Run(t, new(TestSuitePanic))
}

type TestSuitePanic struct {
	suite.Suite
}

func (s *TestSuitePanic) TestPanicCrashesProcess() {
	if os.Getenv("TEST_FORCE_PANIC") == "1" {
		Panic("unrecoverable panic test")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPanic/TestPanicCrashesProcess")
	cmd.Env = append(os.Environ(), "TEST_FORCE_PANIC=1")
	out, err := cmd.CombinedOutput()

	s.Error(err, "process should have exited with non-zero status")
	output := string(out)
	s.Contains(output, "unrecoverable panic test")
	s.Contains(output, "panic_test.go")
	s.T().Logf("panic output:\n%s", output)
}

func (s *TestSuitePanic) TestPanicCannotBeRecovered() {
	if os.Getenv("TEST_FORCE_PANIC") == "1" {
		defer func() {
			recover() // this should NOT save the process
		}()
		Panic("should not be recoverable")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPanic/TestPanicCannotBeRecovered")
	cmd.Env = append(os.Environ(), "TEST_FORCE_PANIC=1")
	out, err := cmd.CombinedOutput()

	s.Error(err, "process should have crashed despite recover()")
	output := string(out)
	s.Contains(output, "should not be recoverable")
	s.T().Logf("panic output (with recover attempt):\n%s", output)
}
