package flow

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestFlow(t *testing.T) {
	suite.Run(t, new(suiteFlow))
}

type suiteFlow struct {
	suite.Suite
}
