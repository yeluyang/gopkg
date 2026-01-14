package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"
)

func TestDynamicRatelimit(t *testing.T) {
	suite.Run(t, new(TestSuiteDynamicRatelimit))
}

type TestSuiteDynamicRatelimit struct {
	suite.Suite
}

func (s *TestSuiteDynamicRatelimit) TestNormal() {
	closed := false
	limit := rate.Limit(0)
	rl := NewDynamicLimiter(
		"test",
		100*time.Millisecond,
		func() rate.Limit {
			s.Require().False(closed)
			limit += 1
			return limit
		},
		func(l rate.Limit) {
			s.Require().False(closed)
			s.Require().Equal(limit, l)
		},
	)
	time.Sleep(time.Second)
	rl.Close()
	time.Sleep(time.Second)
	closed = true
	time.Sleep(time.Second)
	staticLimit := limit
	s.Require().NotZero(limit)
	time.Sleep(time.Second)
	s.Require().Equal(staticLimit, rl.Limit())
	s.Require().Equal(int(staticLimit), rl.Burst())
}

func (s *TestSuiteDynamicRatelimit) TestInitialLimitIsZero() {
	rl := NewDynamicLimiter(
		"test",
		100*time.Millisecond,
		func() rate.Limit {
			return 0
		},
		func(l rate.Limit) {
			s.Require().Zero(l)
		},
	)
	defer rl.Close()
	time.Sleep(time.Second)
	s.Require().Zero(rl.Limit())
	s.Require().Equal(1, rl.Burst())
}

func (s *TestSuiteDynamicRatelimit) TestLimitChangeToZero() {
	flag := true
	rl := NewDynamicLimiter(
		"test",
		100*time.Millisecond,
		func() rate.Limit {
			if flag {
				return rate.Limit(2)
			} else {
				return rate.Limit(0)
			}
		},
		func(l rate.Limit) {
			s.Require().Zero(l)
		},
	)
	defer rl.Close()

	s.Require().Equal(rate.Limit(2), rl.Limit())
	s.Require().Equal(int(rl.Limit()), rl.Burst())

	time.Sleep(time.Second)
	s.Require().Equal(rate.Limit(2), rl.Limit())
	s.Require().Equal(int(rl.Limit()), rl.Burst())

	flag = false
	time.Sleep(time.Second)
	s.Require().Zero(rl.Limit())
	s.Require().Equal(1, rl.Burst())
}
