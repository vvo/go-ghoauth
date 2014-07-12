package main

import (
	"github.com/vvo/go-ghoauth"
	"net/http"
	"net/url"
	"encoding/json"
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

	// we can use the token by requesting user informations
	u, _ := url.Parse("https://api.github.com/user")

	// Use the token with the Authorization header
	req := &http.Request{
		Method: "GET",
		URL: u,
		Header: http.Header{
			"Accept": {"application/json"},
			"Authorization": {"token " + token},
		},
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil || res.StatusCode != 200 {
		w.Write([]byte("\nError while requestion user informations: " + res.Status))
		return
	}

	defer res.Body.Close()

	type userResponse struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Location   string `json:"location"`
	}

	js := userResponse{}
	err = json.NewDecoder(res.Body).Decode(&js)

	w.Write([]byte("\nYour email is " + js.Email))
	w.Write([]byte("\nAnd your name is " + js.Name))
	w.Write([]byte("\nYour location is " + js.Location))
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<a href=/login>login</a>"))
}
