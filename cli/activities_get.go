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

package main

import (
	"flag"
	"fmt"
	"os"
	"template"

	"google-plus-go-starter.googlecode.com/hg/cli/api"
)

// Flags are parsed in main.go.
var activityId *string = flag.String("activityId", "z12gtjhq3qn2xxl2o224exwiqruvtda0i",
	"The ID of the activity to display in the activities.get action. Must be the ID of a *public* activity.")

// ActivitiesGet fetches and displays a specific public Google+ activity using
// unauthenticated (simple) API access.
func ActivitiesGet() os.Error {
	// Get the *plus.Service.
	// Getting specific public activities doesn't require OAuth.
	p, err := api.NoAuthPlus()
	if err != nil {
		return err
	}

	fmt.Printf("Getting activity with ID %q...\n", *activityId)

	// Get a specific public activity.
	activity, err := p.Activities.Get(*activityId).Do()
	if err != nil {
		return err
	}

	// Display the activity.
	return activitiesGetTemplate.Execute(os.Stdout, activity)
}

var activitiesGetTemplate = template.Must(template.New("activities.get").Parse(`
Author: {{.Actor.DisplayName}}
Content: {{.Object.Content}}
Attachment: {{$attachment := index .Object.Attachments 0}}{{$attachment.Url}}

`))
