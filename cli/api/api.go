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
// libraries.
//
// To get started, enable the Google+ API service at
// https://code.google.com/apis/console/ > Services. Then, simply fill in the
// "config.json" file with your values from the API Access section of the same
// site. You must specify the path to the config.json file using Config.
//
// For authenticated API access, the user's OAuth tokens (access token and
// refresh token) can be stored in a file. You can specify the path to this
// file by setting the TokenPath variable.
package api

import (
	"fmt"
	"io"
	"json"
	"os"

	"goauth2.googlecode.com/hg/oauth"
	"google-api-go-client.googlecode.com/hg/plus/v1"
	"google-plus-go-starter.googlecode.com/hg/noauth"
)

// config contains configuration values used to access the Google+ Platform
// APIs. These values are loaded when you call Config.
var config = struct {
	// Your unique Google API Key for simple API Access.
	APIKey string
	// Your OAuth configuration information for protected user data access.
	OAuthConfig oauth.Config
}{}

// Config must be called with the path to the API config file before NoAuthPlus
// and OAuthPlus are called.
func Config(path string) os.Error {
	return readJSON(&config, path)
}

// NoAuthPlus returns a *plus.Service which provides unauthenticated (simple)
// access to the Google+ API. It will initialize the *plus.Service for you.
//
// You must call Config before calling this function.
func NoAuthPlus() (*plus.Service, os.Error) {
	if len(config.APIKey) == 0 {
		return nil, os.NewError("APIKey missing")
	}
	t := &noauth.Transport{APIKey: config.APIKey}
	return plus.New(t.Client())
}

// TokenPath specifies the path to the file where OAuth access and refresh
// tokens will be read and written. See OAuthPlus for more information.
var TokenPath string

// OAuthPlus returns a *plus.Service which provides authenticated (OAuth) access
// to the Google+ API. It will guide the user through the OAuth dance if
// necessary and initialize the *plus.Service.
//
// If TokenPath is set, OAuthPlus will attempt to read and write the OAuth
// access and refresh tokens to the file specified to avoid forcing the user
// through the OAuth dance multiple times.
//
// You must call Config before calling this function.
func OAuthPlus() (*plus.Service, os.Error) {
	transport := &oauth.Transport{Config: &config.OAuthConfig}

	// If a path is specified, read OAuth tokens from the file.
	if len(TokenPath) > 0 {
		token := &oauth.Token{}
		if err := readJSON(token, TokenPath); err != nil {
			fmt.Fprintf(os.Stderr, "[warning] Couldn't read oauth.Token from %s: %s\n",
				TokenPath, err.String())
		} else {
			transport.Token = token
		}
	}

	if transport.Token == nil {
		// Retrieve tokens through the OAuth dance.
		if err := oauthDance(transport, os.Stdin, os.Stdout); err != nil {
			return nil, err
		}

		// If a path is specified, save the tokens to the file. 
		if len(TokenPath) > 0 {
			if err := writeJSON(transport.Token, TokenPath); err != nil {
				fmt.Fprintf(os.Stderr, "[warning] Couldn't write oauth.Token to %s: %s\n",
					TokenPath, err.String())
			}
		}
	}

	return plus.New(transport.Client())
}

// oauthDance creates a new *oauth.Token for transport by guiding the user
// through the OAuth flow. transport's Token field will be set to the new
// *oauth.Token.
func oauthDance(transport *oauth.Transport, r io.Reader, w io.Writer) os.Error {
	// Guide the user through the OAuth flow and read the authorization code.
	if _, err := fmt.Fprintln(w, "Open your browser and go to the following URL:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, transport.Config.AuthCodeURL(""), "\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "Enter the authorization code: "); err != nil {
		return err
	}
	var code string
	if _, err := fmt.Fscan(r, &code); err != nil {
		return err
	}
	// Exchange the authorization code for access and refresh tokens.
	_, err := transport.Exchange(code)
	return err
}

func readJSON(v interface{}, path string) os.Error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(v)
}

func writeJSON(v interface{}, path string) os.Error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(v)
}
