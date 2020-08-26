// Copyright 2020 Nexus Operator and/or its authors
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

package update

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHigherVersion(t *testing.T) {
	tests := []struct {
		name      string
		thisTag   string
		otherTag  string
		want      bool
		wantError bool
	}{
		{
			"equal tags",
			"3.25.0",
			"3.25.0",
			false,
			false,
		},
		{
			"higher minor",
			"3.26.0",
			"3.25.0",
			true,
			false,
		},
		{
			"lower minor",
			"3.24.0",
			"3.25.0",
			false,
			false,
		},
		{
			"higher micro",
			"3.25.1",
			"3.25.0",
			true,
			false,
		},
		{
			"lower micro",
			"3.25.0",
			"3.25.1",
			false,
			false,
		},
		{
			"thisTag with invalid minor",
			"3..0",
			"3.25.1",
			false,
			true,
		},
		{
			"otherTag with invalid minor",
			"3.25.1",
			"3..0",
			false,
			true,
		},
		{
			"thisTag with invalid micro",
			"3.25.",
			"3.25.1",
			false,
			true,
		},
		{
			"otherTag with invalid micro",
			"3.25.1",
			"3.25.",
			false,
			true,
		},
	}

	for _, tt := range tests {
		got, err := HigherVersion(tt.thisTag, tt.otherTag)
		if got != tt.want {
			t.Errorf("%s - want: %v\tgot: %v", tt.name, tt.want, got)
		}
		if (err != nil) != tt.wantError {
			t.Errorf("%s - wantError: %v\tgot: %v", tt.name, tt.wantError, err)
		}
	}
}

func TestGetLatestMicro(t *testing.T) {
	latestMicros = make(map[int]string)
	lastQuery = time.Now()
	minor := 0
	latestMicros[minor] = "3.0.0"
	_, ok := GetLatestMicro(minor)
	assert.True(t, ok)
	_, ok = GetLatestMicro(1)
	assert.False(t, ok)
}

func TestGetLatestMinor(t *testing.T) {
	latestMicros = make(map[int]string)
	lastQuery = time.Now()
	lowerMinor := 0
	higherMinor := 1
	// first, let's test the scenario where we couldn't fetch tags
	_, err := GetLatestMinor()
	assert.NotNil(t, err)
	// now let's populate the tags and test
	latestMicros[lowerMinor] = ""
	latestMicros[higherMinor] = ""
	minor, err := GetLatestMinor()
	assert.Nil(t, err)
	assert.Equal(t, higherMinor, minor)
}

func TestParseTagsAndUpdate(t *testing.T) {
	// make sure latestMicros is blank
	latestMicros = make(map[int]string)
	validTags := []string{"latest", "3.0.0", "3.0.1", "3.1.0"}
	assert.NoError(t, parseTagsAndUpdate(validTags))
	assert.Len(t, latestMicros, 2)
	assert.Equal(t, "3.0.1", latestMicros[0])
	assert.Equal(t, "3.1.0", latestMicros[1])

	invalidMinor := []string{"3..0"}
	invalidMicro := []string{"3.25."}
	assert.Error(t, parseTagsAndUpdate(invalidMinor))
	assert.Error(t, parseTagsAndUpdate(invalidMicro))
}

func TestGetMinor(t *testing.T) {
	validTag := "3.25.0"
	invalidTag := "3..0"

	minor, err := getMinor(validTag)
	assert.Nil(t, err)
	assert.Equal(t, 25, minor)

	_, err = getMinor(invalidTag)
	assert.NotNil(t, err)
}

func TestGetMicro(t *testing.T) {
	validTag := "3.25.0"
	specialCaseTag := "3.9.0-01"
	invalidTag := "3.25."

	micro, err := getMicro(validTag)
	assert.Nil(t, err)
	assert.Equal(t, 0, micro)

	micro, err = getMicro(specialCaseTag)
	assert.Nil(t, err)
	assert.Equal(t, 1, micro)

	_, err = getMicro(invalidTag)
	assert.NotNil(t, err)
}
