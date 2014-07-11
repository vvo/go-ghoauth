# go-ghoauth [![godoc](https://godoc.org/github.com/vvo/go-ghoauth?status.svg)](https://godoc.org/github.com/vvo/go-ghoauth)

Provides easy two steps flow to create an `access_token`
for your wonderfull github application.

[![wercker status](https://app.wercker.com/status/701ab65200a2288626c33231ed44b07f/m "wercker status")](https://app.wercker.com/project/bykey/701ab65200a2288626c33231ed44b07f)

## Example

[example/webapp.go](example/webapp.go)
```go
package main

import (
  "github.com/vvo/go-ghoauth"
  "net/http"
)

var ghflow = ghoauth.New(&ghoauth.Config{
  ClientId:     "82c041166d565cf2f149",
  ClientSecret: "6458b989c06d8106f468fd47106d8654e3887d5b",
  RedirectUri:  "http://localhost:8080/callback",
  Scope:        "user:public",
})

func main() {
  // step 1: redirect users to the right github oauth page
  // https://developer.github.com/v3/oauth/#redirect-users-to-request-github-access
  http.HandleFunc("/login", ghflow.Login)
  http.HandleFunc("/callback", callback)
  
  http.HandleFunc("/", root)
  panic(http.ListenAndServe(":8080", nil))
}

func callback(w http.ResponseWriter, r *http.Request) {
  // step 2: github calls you using redirect_uri
  // we generate a token using github oauth access_token POST url
  // callback does not responds to client for you, it only generates a token
  // https://developer.github.com/v3/oauth/#github-redirects-back-to-your-site
  token, err := ghflow.Callback(r)

  if err != nil {
    w.Write([]byte("UhOh!"))
  }

  // token is now available and usable
  w.Write([]byte("Here's your access_token budy: " + token))
}

func root(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("<a href=/login>login</a>"))
}
```