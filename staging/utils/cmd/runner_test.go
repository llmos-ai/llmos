package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/llmos-ai/llmos/utils/cmd"
)

func Test_Runner(t *testing.T) {
	type input struct {
		key        string
		command    string
		checkExist bool
	}
	type output struct {
		err   error
		exist bool
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "pwd exist",
			given: input{
				command: "pwd",
			},
			expected: output{
				err: nil,
			},
		},
		{
			name: "not exist",
			given: input{
				command:    "ABCDE",
				checkExist: true,
			},
			expected: output{
				exist: false,
			},
		},
		{
			name: "true",
			given: input{
				command: "true",
			},
			expected: output{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		r := cmd.NewRunner()
		if tc.given.checkExist {
			assert.Equal(t, tc.expected.exist, r.CmdExist(tc.given.command))
		} else {
			_, err := r.Run(tc.given.command)
			assert.Equal(t, tc.expected.err, err)
		}
	}

}
