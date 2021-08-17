// Copyright 2021 Nexus Operator and/or its authors
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

package util

import (
	"bytes"
	"fmt"
)

// AppendToStringMap adds the given key and value to the map. If map is nil, create it first
func AppendToStringMap(stringMap map[string]string, key, value string) map[string]string {
	if stringMap == nil {
		stringMap = map[string]string{}
	}
	stringMap[key] = value
	return stringMap
}

// FromMapToJavaProperties converts a given map to a Java properties file string.
// Example:
// # given myMap[string]string = { "key1": "value1", "key2": "value2" }
// # you got back:
//   key1: value1
//   key2: value2
func FromMapToJavaProperties(theMap map[string]string) string {
	if len(theMap) == 0 {
		return ""
	}
	b := new(bytes.Buffer)
	for key, value := range theMap {
		_, _ = fmt.Fprintf(b, "%s: \"%s\"\n", key, value)
	}
	return b.String()
}
