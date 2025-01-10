// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package experimental_test

import (
	"fmt"
	"testing"

	"github.com/google/yamlfmt/formatters/experimental"
	"github.com/google/yamlfmt/internal/assert"
	"github.com/google/yamlfmt/internal/multilinediff"
)

func newFormatter(config *experimental.Config) *experimental.ExperimentalFormatter {
	return &experimental.ExperimentalFormatter{
		Config: config,
	}
}

func TestExperimentalFormatter(t *testing.T) {
	testCases := []struct {
		name        string
		config      *experimental.Config
		initialYaml string
		expected    string
	}{
		{
			name: "initial test case",
			initialYaml: `a:
    b: 1`,
			expected: `a:
  b: 1
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.config
			if config == nil {
				config = experimental.DefaultConfig()
			}
			f := newFormatter(config)
			result, err := f.Format([]byte(tc.initialYaml))
			assert.NilErr(t, err)
			fmt.Println(string(result))
			if d, found := multilinediff.Diff(string(result), tc.expected, "\n"); found > 0 {
				t.Fatal(d)
			}
		})
	}
}
