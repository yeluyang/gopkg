package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUtils(t *testing.T) {
	suite.Run(t, new(TestSuiteUtils))
}

type TestSuiteUtils struct {
	suite.Suite
}

func (s *TestSuiteUtils) TestForcePanicWithNilError() {
	// Should not panic when error is nil
	s.NotPanics(func() {
		ForcePanic(nil)
	})
}

// Note: ForcePanic with non-nil error spawns a goroutine that panics,
// which cannot be caught in tests without crashing the test process.
// The behavior is intentional - it forces a crash to surface critical errors.
