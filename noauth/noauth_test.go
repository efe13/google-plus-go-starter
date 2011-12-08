// Copyright 2011 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package noauth

import (
	"http"
	"os"
	"regexp"
	"testing"
)

type RoundTripTest struct {
	key, in, out string
}

var RoundTripTests = []RoundTripTest{
	// Expect an error if you don't specify an API key
	RoundTripTest{key: "", in: "https://www.example.com/", out: ""},
	// Happy case
	RoundTripTest{key: "abc", in: "https://www.example.com/", out: `^https://www\.example\.com/\?key=abc$`},
	RoundTripTest{key: "abc", in: "example.com", out: `^example\.com\?key=abc$`},
	RoundTripTest{key: "abc", in: "https://www.example.com/?query=foo", out: `^https://www\.example\.com/\?(query=foo&key=abc|key=abc&query=foo)$`},
	// Duplicate "key" querystring parameter
	RoundTripTest{key: "abc", in: "https://www.example.com/?key=def", out: `^https://www\.example\.com/\?(key=def&key=abc|key=abc&key=def)$`},
	// URL encode the "key"
	RoundTripTest{key: "abc?", in: "https://www.example.com/", out: `^https://www\.example\.com/\?key=abc%3F$`},
}

type fakeRoundTripper struct{}

func (t *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, os.Error) {
	return nil, nil
}

func TestRoundTrip(t *testing.T) {
	for _, r := range RoundTripTests {
		// Create the Transport being tested.
		transport := &Transport{
			APIKey:    r.key,
			Transport: &fakeRoundTripper{},
		}

		// Create the HTTP request from the input URL.
		req, err := http.NewRequest("FAKE", r.in, nil)
		if err != nil {
			t.Error(err)
		}

		// Exercise the code being tested.
		_, err = transport.RoundTrip(req)

		// If r.out is the empty string, the testcase expects an error.
		if len(r.out) == 0 {
			if err != nil {
				// Continue to the next test case.
				continue
			} else {
				t.Error("Expected an error.")
			}
		}

		// If we made it here, the testcase does not expect an error.
		if err != nil {
			t.Error(err)
		}

		// Compare how the Transport transformed the HTTP request to how we expected
		// it to.
		out := req.URL.String()
		if ok, err := regexp.MatchString(r.out, out); !ok || err != nil {
			t.Errorf("key %q and URL %q expected %q but got %q", r.key, r.in, r.out, out)
		}
	}
}
