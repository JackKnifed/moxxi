package moxxiConf

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErr_Error(t *testing.T) {
	fakeError := errors.New("fake error")
	var testData = []struct {
		in  Err
		out string
	}{
		{
			NewErr{ErrCloseFile, "/tmp/testfile", fakeError},
			"failed to close the file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrRemoveFile, "/tmp/testfile", fakeError},
			"failed to remove file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrFilePerm, "/tmp/testfile", fakeError},
			"permission denied to create file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrFileUnexpect, "/tmp/testfile", fakeError},
			"unknown error with file [/tmp/testfile] - fake error",
		}, {
			NewErr{ErrBadHost, "/tmp/testfile", nil},
			"bad hostname provided [/tmp/testfile]",
		}, {
			NewErr{ErrBadIP, "/tmp/testfile", nil},
			"bad IP provided [/tmp/testfile]",
		}, {
			NewErr{ErrNoRandom, "", nil},
			"was not given a new random domain - shutting down",
		},
	}
	for _, test := range testData {
		testOut := test.in.Error()
		assert.Equal(t, test.out, testOut, "errors should match")
	}
}

func TestHandlerLocFlag(t *testing.T) {
	var testData = []string{
		"/one",
		"/two",
		"three",
		"/four",
	}

	var expected = []string{
		"/one",
		"/two",
		"/four",
	}

	testWork := new(HandlerLocFlag)

	for _, each := range testData {
		err := testWork.Set(each)
		assert.NoError(t, err, "there should not have been a problem adding an item")
	}

	assert.Equal(t, "/one /two /four", testWork.String(), "the test input and current value of the test should be equal")

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testWork.GetOne(i), "one item from the test was incorrect")
	}

	junkTest := new(HandlerLocFlag)
	assert.Equal(t, "", junkTest.String(), "should be empty")
	junkTest.Set("/some/real/junk")
	assert.Equal(t, "/some/real/junk", junkTest.String(), "should be empty")

}
