package sdk

import (
	"fmt"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

var (
	_e error               = &Error{}
	_  types.GomegaMatcher = HaveOccurredWithStatusCode(0)
	_  types.GomegaMatcher = HaveOccurredWithErrorMsg("")
)

// Error is the error returned by the dashboard.
type Error struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"-"`
	// Error_ is the error details, it's exclusive with Payload.
	Error_ string `json:"error,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("status code: %d, error: %s", e.StatusCode, e.Error_)
}

type (
	HaveOccurredWithStatusCode   int
	HaveOccurredWithErrorMsg     string
	HaveOccurredContainsErrorMsg string
)

func (c HaveOccurredWithStatusCode) Match(actual any) (success bool, err error) {
	match, err := gomega.HaveOccurred().Match(actual)
	if !match {
		return match, err
	}
	match, err = gomega.BeAssignableToTypeOf(_e).Match(actual)
	if !match {
		return match, err
	}
	return actual.(*Error).StatusCode == int(c), nil
}

func (c HaveOccurredWithStatusCode) FailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred with StatusCode", c)
}

func (c HaveOccurredWithStatusCode) NegatedFailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred but not with StatusCode", c)
}

func (m HaveOccurredWithErrorMsg) Match(actual any) (success bool, err error) {
	match, err := gomega.HaveOccurred().Match(actual)
	if !match {
		return match, err
	}
	match, err = gomega.BeAssignableToTypeOf(_e).Match(actual)
	if !match {
		return match, err
	}
	return gomega.Equal(string(m)).Match(actual.(*Error).Error_)
}

func (m HaveOccurredWithErrorMsg) FailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred with ErrorMsg", m)
}

func (m HaveOccurredWithErrorMsg) NegatedFailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred but not with ErrorMsg", m)
}

func (m HaveOccurredContainsErrorMsg) Match(actual any) (success bool, err error) {
	match, err := gomega.HaveOccurred().Match(actual)
	if !match {
		return match, err
	}
	match, err = gomega.BeAssignableToTypeOf(_e).Match(actual)
	if !match {
		return match, err
	}
	return gomega.ContainSubstring(string(m)).Match(actual.(*Error).Error_)
}

func (m HaveOccurredContainsErrorMsg) FailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred contains ErrorMsg", m)
}

func (m HaveOccurredContainsErrorMsg) NegatedFailureMessage(actual any) (message string) {
	return format.Message(actual, "*Error have occurred but not contains ErrorMsg", m)
}
