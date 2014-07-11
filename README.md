# go-ghoauth [![godoc](https://godoc.org/github.com/vvo/go-ghoauth?status.svg)](https://godoc.org/github.com/vvo/go-ghoauth)

Provides easy two steps flow to create an `access_token`
for your wonderfull github application.

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
  http.HandleFunc("/login", ghflow.Login)
  http.HandleFunc("/callback", callback)
  http.HandleFunc("/", root)
  panic(http.ListenAndServe(":8080", nil))
}

func callback(w http.ResponseWriter, r *http.Request) {
  token, err := ghflow.Callback(r)

  if err != nil {
    w.Write([]byte("UhOh!"))
  }

  w.Write([]byte("Here's your access_token budy: " + token))
}

func root(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("<a href=/login>login</a>"))
}
```