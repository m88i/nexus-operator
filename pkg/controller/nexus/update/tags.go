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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/heroku/docker-registry-client/registry"
)

const (
	communityNexusRegistry     = "https://registry.hub.docker.com"
	communityNexusRepo         = "sonatype/nexus3"
	tagParseFailureFormat      = "unable to parse tag \"%s\": %v"
	unableToCheckUpdatesFormat = "Unable to check for updates: %v"
)

const (
	ttl    = time.Hour * 6
	errTTL = time.Minute
)

var (
	lastQuery    time.Time
	lastErr      time.Time
	latestMicros = make(map[int]string)
)

// HigherVersion checks if thisTag is of a higher version than otherTag
func HigherVersion(thisTag, otherTag string) (bool, error) {
	thisMinor, err := getMinor(thisTag)
	if err != nil {
		return false, fmt.Errorf(tagParseFailureFormat, thisTag, err)
	}
	otherMinor, err := getMinor(otherTag)
	if err != nil {
		return false, fmt.Errorf(tagParseFailureFormat, otherTag, err)
	}
	if thisMinor != otherMinor {
		return thisMinor > otherMinor, nil
	}

	thisMicro, err := getMicro(thisTag)
	if err != nil {
		return false, fmt.Errorf(tagParseFailureFormat, thisTag, err)
	}
	otherMicro, err := getMicro(otherTag)
	if err != nil {
		return false, fmt.Errorf(tagParseFailureFormat, otherTag, err)
	}
	return thisMicro > otherMicro, nil
}

// GetLatestMicro returns the most recent image tag within a minor (the "y" in "x.y.z").
// If the minor was not found or if we never managed to fetch any tags, the second return value is false.
func GetLatestMicro(minor int) (tag string, ok bool) {
	if time.Since(lastQuery) > ttl {
		fetchUpdates()
	}
	tag, ok = latestMicros[minor]
	return
}

// GetLatestMinor returns the most recent minor (the "y" in "x.y.z").
// If there were issues fetching the tags it returns an error.
func GetLatestMinor() (int, error) {
	if time.Since(lastQuery) > ttl {
		fetchUpdates()
	}
	if len(latestMicros) == 0 {
		return 0, fmt.Errorf("unable to fetch tags")
	}

	greatestMinor := 0
	for minor := range latestMicros {
		if minor > greatestMinor {
			greatestMinor = minor
		}
	}
	return greatestMinor, nil
}

func fetchUpdates() {
	if time.Since(lastErr) < errTTL {
		log.Debugf("Trying to fetch tags from registry again too fast, must try again later")
		return
	}

	tags, err := getTags()
	if err != nil {
		lastErr = time.Now()
		log.Errorf(unableToCheckUpdatesFormat, err)
		return
	}
	lastQuery = time.Now()

	if err = parseTagsAndUpdate(tags); err != nil {
		log.Errorf(unableToCheckUpdatesFormat, err)
		return
	}
}

func getTags() ([]string, error) {
	reg, err := registry.New(communityNexusRegistry, "", "")
	if err != nil {
		return nil, fmt.Errorf("unable to create client for registry: %v", err)
	}
	// redirect the lib's logging to ours
	reg.Logf = func(format string, args ...interface{}) {
		format = fmt.Sprintf("Registry: %s", format)
		log.Infof(format, args)
	}

	repo := communityNexusRepo
	tags, err := reg.Tags(repo)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch tags from %s: %v", repo, err)
	}
	return tags, nil
}

func parseTagsAndUpdate(tags []string) error {
	for _, candidateTag := range tags {
		if candidateTag != "latest" {
			candidateMinor, err := getMinor(candidateTag)
			if err != nil {
				return fmt.Errorf(tagParseFailureFormat, candidateTag, err)
			}
			candidateMicro, err := getMicro(candidateTag)
			if err != nil {
				return fmt.Errorf(tagParseFailureFormat, candidateTag, err)
			}
			storedTag, ok := latestMicros[candidateMinor]
			if ok {
				// we can safely ignore the error. It wouldn't be stored if it was invalid
				storedMicro, _ := getMicro(storedTag)
				if candidateMicro > storedMicro {
					latestMicros[candidateMinor] = candidateTag
				}
			} else {
				latestMicros[candidateMinor] = candidateTag
			}
		}
	}
	return nil
}

func getMinor(tag string) (int, error) {
	return strconv.Atoi(strings.Split(tag, ".")[1])
}

func getMicro(tag string) (int, error) {
	// special case for the community tag 3.9.0-01
	if tag == "3.9.0-01" {
		return 1, nil
	}
	return strconv.Atoi(strings.Split(tag, ".")[2])
}
