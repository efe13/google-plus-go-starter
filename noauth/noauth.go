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

// The noauth package provides support for making unauthenticated HTTP requests
// to Google APIs (i.e. simple API access). It provides an interface similar to
// goauth2.googlecode.com/hg/oauth.
//
// Example usage:
// 	t := &noauth.Transport{APIKey: YOUR_API_KEY}
// 	c := t.Client()
// 	c.Post(...)
package noauth

import (
	"http"
	"os"
)

// Transport implements http.RoundTripper. When configured with a valid API key,
// it can be used to make unauthenticated HTTP requests to Google APIs.
//
// 	t := &noauth.Transport{APIKey: YOUR_API_KEY}
// 	c := t.Client()
//  r, err := c.Get("https://www.googleapis.com/plus/v1/people?query=Vic")
//
// It will automatically append the API key as a querystring parameter.
type Transport struct {
	// APIKey is your unique simple API access key from
	// https://code.google.com/apis/console > API Access > Simple API Access
	APIKey string
	// Transport is the HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	// (It should never be a noauth.Transport.)
	Transport http.RoundTripper
}

// Client returns an *http.Client that can make unauthenticated requests to
// Google APIs (i.e. simple API access).
//
// This client can and should be reused when making multiple API requests.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// RoundTrip executes a single HTTP transaction appending the Transport's API
// key as a querystring parameter.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, os.Error) {
	if t.APIKey == "" {
		return nil, os.NewError("No APIKey supplied")
	}

	newReq := *req
	u := newReq.URL
	q := u.Query()
	q.Add("key", t.APIKey)
	u.RawQuery = q.Encode()

	return t.transport().RoundTrip(&newReq)
}
