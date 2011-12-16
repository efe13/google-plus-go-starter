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
	"fmt"
	"os"
	"template"

	"google-plus-go-starter.googlecode.com/hg/cli/api"
)

// PlusMe fetches and displays the user's public Google+ profile using
// authenticated (OAuth) API access.
func PlusMe() os.Error {
	// Get the *plus.Service.
	// Associating a user with their Google+ profile requires OAuth.
	p, err := api.OAuthPlus()
	if err != nil {
		return err
	}

	fmt.Println("Getting the authenticated user's profile...")

	// Get the user's profile.
	// "me" is a special value that refers to the authenticated user.
	me, err := p.People.Get("me").Do()
	if err != nil {
		return err
	}

	// Display the user's profile.
	return plusMeTemplate.Execute(os.Stdout, me)
}

var plusMeTemplate = template.Must(template.New("plus.me").Parse(`
Name: {{.DisplayName}}
Profile: {{.Url}}
About: {{.AboutMe}}

`))
