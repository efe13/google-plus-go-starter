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
	http.HandleFunc("/activities_get", requireLoginURLs(api.WithNoAuthPlus(activitiesGetHandler)))
}

func activitiesGetHandler(w http.ResponseWriter, r *http.Request, p *plus.Service) {
	// Get a specific activity by its ID. You can retrieve activities given their
	// IDs from any source (doesn't have to be hard-coded like it is here), such
	// as from Activities.list or Activities.search calls.
	activity, err := p.Activities.Get("z12gtjhq3qn2xxl2o224exwiqruvtda0i").Do()
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// Display the activity.
	c := appengine.NewContext(r)
	err = templates.Execute(w, "activities_get.html", map[string]interface{}{
		"user":     user.Current(c),
		"activity": activity,
	})
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
}
