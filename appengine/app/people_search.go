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
	"os"

	"app/api"
)

func init() {
	http.HandleFunc("/people_search", requireLoginURLs(api.WithNoAuthPlus(peopleSearchHandler)))
}

func peopleSearchHandler(w http.ResponseWriter, r *http.Request, p *plus.Service) {
	// Retrieve the user's query (from a querystring parameter).
	query := r.FormValue("query")

	var people []*plus.Person
	var err os.Error

	// If the user entered a query, search for people (profiles) matching it.
	if len(query) > 0 {
		peopleFeed, err := p.People.Search(query).Do()
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		// The list of people resources is wrapped in a peopleFeed object.
		people = peopleFeed.Items
	}

	// Display the people.
	c := appengine.NewContext(r)
	err = templates.Execute(w, "people_search.html", map[string]interface{}{
		"user":   user.Current(c),
		"query":  query,
		"people": people,
	})
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
