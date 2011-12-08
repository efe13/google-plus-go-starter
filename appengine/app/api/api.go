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

// The api package provides helper functions for using the Google+ API client
// libraries on Google App Engine.
//
// To get started, enable the Google+ API service at
// https://code.google.com/apis/console/ > Services. Then, simply fill in the
// app/api/config.json with your values from the API Access section of the same
// site.
//
// Since the Google App Engine URL Fetch API requires a per-request context,
// you must use the *plus.Service from within an HTTP handler. This package
// provides the WithNoAuthPlus and WithOAuthPlus functions which you can use
// to wrap your HTTP handlers to provide them with fully initialized
// *plus.Services. See their documentation for usage details.
//
// For authenticated API access, the per-user OAuth tokens (access token and
// refresh token) are stored in the datastore as "oauth.Token" entities, keyed
// by the user's ID. Thus, the user must be logged into a Google account for
// any HTTP handlers that require the use of an authenticated (OAuth)
// *plus.Service. This is automatically enforced by the WithOAuthPlus wrapper
// function.
package api

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"appengine/user"
	"fmt"
	"http"
	"json"
	"os"
	"url"

	"goauth2.googlecode.com/hg/oauth"
	"google-api-go-client.googlecode.com/hg/plus/v1"
	"google-plus-go-starter.googlecode.com/hg/noauth"
)

// config contains configuration values used to access the Google+ Platform
// APIs. These values are automatically loaded by this module from
// "app/api/config.json".
var config = struct {
	// Your unique Google API Key for simple API Access.
	APIKey string
	// Your OAuth configuration information for protected user data access.
	OAuthConfig oauth.Config
	// The path in your application to which users will be redirected after they
	// allow or deny permission for your application to access their data.
	OAuthRedirectPath string
	// The scheme, hostname and port at which your application can be accessed
	// when running on the local development server.
	DevRootURL string
	// The scheme, hostname and port at which your application can be accessed
	// when running on App Engine.
	ProdRootURL string
}{}

const configPath = "app/api/config.json"

// init loads the "app/api/config.json" file into the "config" struct and registers
// an HTTP request handler function to handle OAuth redirect requests.
//
// The handler function is registered at the path defined by the
// OAuthRedirectPath" attribute declared in "app/api/config.json".
func init() {
	// Open, parse and load "app/api/config.json" into the "config" struct.
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(fmt.Sprintf("Could not open %s: %s", configPath, err.String()))
	}
	defer configFile.Close()

	if err = json.NewDecoder(configFile).Decode(&config); err != nil {
		panic(fmt.Sprintf("Could not parse %s: %s", configPath, err.String()))
	}

	// Set the OAuth redirect URL depending on whether the application is running
	// on the local development server or on App Engine.
	if appengine.IsDevAppServer() {
		config.OAuthConfig.RedirectURL = config.DevRootURL
	} else {
		config.OAuthConfig.RedirectURL = config.ProdRootURL
	}
	redirectURL, err := url.Parse(config.OAuthConfig.RedirectURL)
	if err != nil {
		panic(fmt.Sprintf("Could not parse RootURL in %s: %s", configPath, err.String()))
	}
	redirectURL.Path = config.OAuthRedirectPath
	config.OAuthConfig.RedirectURL = redirectURL.String()

	// Register the OAuth redirect URL handler.
	http.HandleFunc(config.OAuthRedirectPath, oauthHandler)
}

// requireUser is used to wrap HTTP request handlers to ensure that the user
// is logged into a Google account. If they aren't, they're redirected to a
// login page.
func requireUser(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		u := user.Current(c)

		if u == nil {
			loginURL, err := user.LoginURL(c, r.URL.RawPath)
			if err != nil {
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		handler(w, r)
	}
}

type HandlerFunc func(http.ResponseWriter, *http.Request, *plus.Service)

// WithNoAuthPlus is used to wrap HTTP request handlers which require the use of
// a *plus.Service that makes unauthenticated requests to Google APIs (i.e.
// simple API access). The *plus.Service will be initialized automatically for
// the handler.
//
// Example usage:
// 	func init() {
// 		http.HandleFunc("/hello", WithNoAuthPlus(helloHandler))
// 	}
//
// 	func helloHandler(w http.ResponseWriter, r *http.Request, p *plus.Service) {
// 		me, _ := p.People.Get("106189723444098348646").Do()
// 		fmt.Fprintf(w, "Hello, %s!", me.DisplayName)
// 	}
func WithNoAuthPlus(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		// Initialize the *plus.Service.
		t := &noauth.Transport{
			APIKey:    config.APIKey,
			Transport: &urlfetch.Transport{Context: c},
		}

		p, err := plus.New(t.Client())
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}

		// Execute the wrapped handler, providing the *plus.Service.
		handler(w, r, p)
	}
}

// WithOAuthPlus is used to wrap HTTP request handlers which require the use of
// a *plus.Service that makes authenticated (OAuth) requests to Google APIs. The
// *plus.Service will be initialized automatically for the handler, including
// having the user do the OAuth dance.
//
// Example usage:
// 	func init() {
// 		http.HandleFunc("/hello", WithOAuthPlus(helloHandler))
// 	}
//
// 	func helloHandler(w http.ResponseWriter, r *http.Request) {
// 		me, _ := Plus.People.Get("me").Do()
// 		fmt.Fprintf(w, "Hello, %s!", me.DisplayName)
// 	}
func WithOAuthPlus(handler HandlerFunc) http.HandlerFunc {
	return requireUser(func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		// Get the OAuth tokens for the current user from the datastore.
		token := &oauth.Token{}
		key := datastore.NewKey(c, "oauth.Token", user.Current(c).Id, 0, nil)
		err := datastore.Get(c, key, token)
		if err != nil && err != datastore.ErrNoSuchEntity {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}

		// If the user does not have OAuth tokens, make them do the OAuth dance.
		if err == datastore.ErrNoSuchEntity {
			// Redirect to the Google OAuth permissions page. Use the state to
			// remember where the user originally wanted to go.
			url := config.OAuthConfig.AuthCodeURL(r.URL.RawPath)
			http.Redirect(w, r, url, http.StatusFound)
		} else {
			// Initialize the *plus.Service.
			t := &oauth.Transport{
				Config:    &config.OAuthConfig,
				Token:     token,
				Transport: &urlfetch.Transport{Context: c},
			}

			p, err := plus.New(t.Client())
			if err != nil {
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}

			// Execute the wrapped handler, providing the *plus.Service.
			handler(w, r, p)

			// Save the OAuth tokens back to the datastore, in case they have been
			// updated.
			if _, err = datastore.Put(c, key, token); err != nil {
				http.Error(w, err.String(), http.StatusInternalServerError)
				return
			}
		}
	})
}

// oauthHandler handles the OAuth logic when the user is redirected from Google
// after they allow or deny permission for your application to access their
// data.
//
// This handler does the following:
// 	1. Checks for errors (e.g. if the user denied your application permission).
// 	2. Swaps the authorization code for an access token and refresh token.
// 	3. Saves the access and refresh tokens in the datastore, associated with the
// 	user. This is done so that your application will not have to ask for the
// 	user's permission again once the access token expires.
func oauthHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	trans := &oauth.Transport{
		Config:    &config.OAuthConfig,
		Transport: &urlfetch.Transport{Context: c},
	}

	if error := r.FormValue("error"); len(error) > 0 {
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	code := r.FormValue("code")
	if len(code) == 0 {
		http.Error(w, "Missing access code", http.StatusBadRequest)
		return
	}

	// Swap the access code for an access token and refresh token.
	token, err := trans.Exchange(code)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// Save the tokens for this user in the datastore.
	key := datastore.NewKey(c, "oauth.Token", user.Current(c).Id, 0, nil)
	if _, err = datastore.Put(c, key, token); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// Redirect the user back to the location they tried to access before doing
	// the OAuth dance.
	http.Redirect(w, r, r.FormValue("state"), http.StatusFound)
}
