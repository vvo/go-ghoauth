// Package ghoauth provides acces_token creation for your github application
//
// Two steps:
//	1. redirect user to github auth page
//	2. create an access_token
//
// Example:
//	package main
//
//	import (
//		"github.com/vvo/go-ghoauth"
//		"net/http"
//	)
//
//	var ghflow = ghoauth.New(&ghoauth.Config{
//		ClientId:     "82c041166d565cf2f149",
//		ClientSecret: "6458b989c06d8106f468fd47106d8654e3887d5b",
//		RedirectUri:  "http://localhost:8080/callback",
//		Scope:        "user:public",
//	})
//
//	func main() {
//		// step 1: redirect users to the right github oauth page
//		// https://developer.github.com/v3/oauth/#redirect-users-to-request-github-access
//		http.HandleFunc("/login", ghflow.Login)
//		http.HandleFunc("/callback", callback)
//
//		http.HandleFunc("/", root)
//		panic(http.ListenAndServe(":8080", nil))
//	}
//
//	func callback(w http.ResponseWriter, r *http.Request) {
//		// step 2: github calls you using redirect_uri
//		// we generate a token using github oauth access_token POST url
//		// callback does not responds to client for you, it only generates a token
//		// https://developer.github.com/v3/oauth/#github-redirects-back-to-your-site
//		token, err := ghflow.Callback(r)
//
//		if err != nil {
//			w.Write([]byte("UhOh!"))
//		}
//
//		// token is now available and usable
//		w.Write([]byte("Here's your access_token budy: " + token))
//	}
//
//	func root(w http.ResponseWriter, r *http.Request) {
//		w.Write([]byte("<a href=/login>login</a>"))
//	}
//
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

// refers to https://developer.github.com/v3/oauth/#github-redirects-back-to-your-site
// "If the states don’t match, the request has been created
// by a third party and the process should be aborted."
var ErrStateNotFound = errors.New("Provided `state` was not found")

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
		return "", ErrStateNotFound
	}

	delete(states, r.FormValue("state"))

	type access_tokenResponse struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
		GithubError
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
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	js := access_tokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&js)

	if err != nil {
		return "", err
	}

	if js.AccessToken == "" {
		return "", &GithubError{js.RawError, js.ErrorDescription, js.ErrorUri}
	}

	return js.AccessToken, nil
}

func New(config *Config) *oauthFlow {
	if config.BaseUrl == "" {
		config.BaseUrl = defaultBaseUrl
	}

	return &oauthFlow{config}
}

type GithubError struct {
	RawError         string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

func (e *GithubError) Error() string {
	return fmt.Sprintf(
		"GitHub error: %s, error_description: %s, error_uri: %s",
		e.RawError, e.ErrorDescription, e.ErrorUri,
	)
}
