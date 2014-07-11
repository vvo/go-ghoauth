// Package ghoauth provides acces_token creation for your github application

// Two steps usage:
//   1. redirect user to github auth page
//   2. create an access_token
//
// Example usage
//	var gh = ghoauth.New(&ghoauth.Config{
//		ClientId:     "$client_id",
//		ClientSecret: "$client_secret",
//		RedirectUri:  "$redirect_uri",
//		Scope:        "user:public",
//	})
//
//	// step 1: redirect users to the right github oauth page
//	// https://developer.github.com/v3/oauth/#redirect-users-to-request-github-access
//	func loginHandler(w http.ResponseWriter, r *http.Request) {
//		gh.Login(w, r)
//	}
//
//	// step 2: generate a token using github oauth access_token POST url
//	// callback does not responds to client for you, it only generates a token
//	// https://developer.github.com/v3/oauth/#github-redirects-back-to-your-site
//	func callbackHandler(w http.ResponseWriter, r *http.Request) {
//		token, err := gh.Callback(r)
//
//		if err != nil {
//			println("Something went wrong")
//		}
//
//		// token is now available and usable
//	}
package ghoauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
	"net/url"
	"path"
)

const (
	defaultBaseUrl = "https://github.com"
)

// See https://developer.github.com/v3/oauth/#web-application-flow
// to understand how to fill the properties
type Config struct {
	ClientId     string
	ClientSecret string

	RedirectUri string

	// Comma separated list of scopes https://developer.github.com/v3/oauth/#scopes
	Scope string

	// Defaults to defaultBaseUrl
	BaseUrl string
}

type oauthFlow struct {
	*Config
}

var states map[string]bool = make(map[string]bool)

// Redirects to https://github.com/login/oauth/authorize
func (f *oauthFlow) Login(w http.ResponseWriter, r *http.Request) {
	redirect, _ := url.Parse(f.Config.BaseUrl)
	redirect.Path = path.Join(redirect.Path, "login/oauth/authorize")

	q := redirect.Query()
	q.Set("client_id", f.ClientId)

	state := uuid.NewV4().String()
	states[state] = true

	q.Set("state", state)

	if f.RedirectUri != "" {
		q.Set("redirect_uri", f.RedirectUri)
	}

	if f.Scope != "" {
		q.Set("scope", f.Scope)
	}

	redirect.RawQuery = q.Encode()

	http.Redirect(w, r, redirect.String(), http.StatusTemporaryRedirect)
}

// Creates and return an access_token
func (f *oauthFlow) Callback(r *http.Request) (string, error) {

	if states[r.FormValue("state")] != true {
		return "", errors.New("Provided `state` was not found")
	}

	delete(states, r.FormValue("state"))

	type access_tokenResponse struct {
		AccessToken      string `json:"access_token"`
		Scope            string `json:"scope"`
		TokenType        string `json:"token_type"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		ErrorUri         string `json:"error_uri"`
	}

	u, _ := url.Parse(f.BaseUrl)
	if u.Path == "" {
		u.Path = "/"
	}

	u.Path = path.Join(u.Path, "login/oauth/access_token")

	q := u.Query()
	q.Set("client_id", f.ClientId)
	q.Set("client_secret", f.ClientSecret)
	q.Set("code", r.FormValue("code"))
	u.RawQuery = q.Encode()

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Accept": {"application/json"},
		},
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	js := access_tokenResponse{}
	err = json.NewDecoder(res.Body).Decode(&js)

	if err != nil {
		println(res.Status)
		return "", err
	}

	if js.AccessToken == "" {
		return "", fmt.Errorf(
			"GitHub error: %s, error_description: %s, error_uri: %s",
			js.Error, js.ErrorDescription, js.ErrorUri)
	}

	return js.AccessToken, nil
}

func New(config *Config) *oauthFlow {
	if config.BaseUrl == "" {
		config.BaseUrl = defaultBaseUrl
	}

	return &oauthFlow{config}
}