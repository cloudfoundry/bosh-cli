package mocks

import "github.com/golang/mock/gomock"

// Shamelessly lifted from https://github.com/onsi/gomega/issues/451

type GomegaMatcher interface {
	Match(actual interface{}) (success bool, err error)
	FailureMessage(actual interface{}) (message string)
}

type matcher struct {
	GomegaMatcher
	x interface{}
}

func (m matcher) Matches(x interface{}) bool {
	m.x = x
	result, _ := m.Match(x)
	return result
}

func (m matcher) String() string {
	return m.FailureMessage(m.x)
}

func GomegaMock(gmather GomegaMatcher) gomock.Matcher {
	return matcher{gmather, nil}
}
