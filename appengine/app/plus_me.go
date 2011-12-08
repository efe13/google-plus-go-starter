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

package app

import (
	"appengine"
	"appengine/user"
	"google-api-go-client.googlecode.com/hg/plus/v1"
	"http"

	"app/api"
)

func init() {
	http.HandleFunc("/plus_me", requireLoginURLs(api.WithOAuthPlus(plusMeHandler)))
}

func plusMeHandler(w http.ResponseWriter, r *http.Request, p *plus.Service) {
	// Since p is an authenticated (OAuth) *plus.Service (from the
	// api.WithOAuthPlus wrapper above), we can make People requests referencing
	// the special identifier "me". "me" is used to indicate the authenticated
	// user (i.e. the one who went through the OAuth flow).
	me, err := p.People.Get("me").Do()
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// Display the user's profile.
	c := appengine.NewContext(r)
	err = templates.Execute(w, "plus_me.html", map[string]interface{}{
		"user": user.Current(c),
		"me":   me,
	})
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
