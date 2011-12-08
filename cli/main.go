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
	"sort"
	"strings"

	"google-plus-go-starter.googlecode.com/hg/cli/api"
)

type actionFunc func() os.Error

// actions maps command-line names to functions that demonstrate the use of the
// Google+ API. Users can specify which action(s) to run with the "action" flag.
var actions = map[string]actionFunc{
	"activities.get": ActivitiesGet,
	"people.search":  PeopleSearch,
	"plus.me":        PlusMe,
}

var action *string = flag.String("action", "all",
	"The action(s) to execute. One of: all, "+strings.Join(keys(actions), ", "))
var configPath *string = flag.String("configPath", "",
	"The path to the file containing API access information.")
var tokenPath *string = flag.String("tokenPath", "",
	"The path to the file where OAuth tokens will be read and written. Optional.")

func main() {
	flag.Parse()

	if len(*configPath) == 0 {
		fmt.Fprintln(os.Stderr, "You must supply the configPath flag.")
		os.Exit(1)
	}

	// Set up the API helper functions.
	if err := api.Config(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, "Could not configure API: ", err)
		os.Exit(1)
	}
	api.TokenPath = *tokenPath

	// Execute specified action(s).
	if *action == "all" {
		for _, name := range keys(actions) {
			executeAction(name)
		}
	} else {
		executeAction(*action)
	}
}

// executeAction prints the name of the action and executes it.
// It will os.Exit with an error if the action does not exist or the action
// encountered an error during execution.
func executeAction(name string) {
	fn, ok := actions[name]
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid action name: ", name)
		os.Exit(1)
	}

	// Print the action name.
	fmt.Println(name)
	fmt.Println(strings.Repeat("-", len(name)))

	// Execute the action.
	if err := fn(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// keys returns a slice containing all keys in the map in increasing order.
func keys(m map[string]actionFunc) []string {
	keys := make([]string, len(m))
	i := 0
	for key, _ := range m {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}
