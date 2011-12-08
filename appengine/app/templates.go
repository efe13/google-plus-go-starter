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
	"http"
	"os"
	"template"
)

var templates = &template.Set{}

// init parses all *.html files in the templates directory into a single, global
// template.Set named templates.
func init() {
	// Add these stub functions now so the templates compile. The actual
	// implementations must be added for each request.
	templates.Funcs(template.FuncMap{
		"loginurl":  func(dest string) (string, os.Error) { return "", nil },
		"logouturl": func(dest string) (string, os.Error) { return "", nil },
	})
	if _, err := templates.ParseTemplateGlob("templates/*.html"); err != nil {
		panic(err)
	}
}

// requireLoginURLs is used to wrap HTTP handler functions to provide the
// appengine/user.LoginURL and appengine/user.LogoutURL functions in templates,
// under the aliases loginurl and logouturl respectively.
func requireLoginURLs(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		templates.Funcs(template.FuncMap{
			"loginurl":  func(dest string) (string, os.Error) { return user.LoginURL(c, dest) },
			"logouturl": func(dest string) (string, os.Error) { return user.LogoutURL(c, dest) },
		})

		handler(w, r)
	}
}
