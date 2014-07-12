package main

import (
	"github.com/vvo/go-ghoauth"
	"net/http"
	"github.com/google/go-github/github"
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
		return
	}

	// token is now available and usable
	w.Write([]byte("Here's your access_token budy: " + token))

	// we can use it with go-github by creating a custom httpClient that can
	// handle authentication for you by adding an Authorization header when
	// requesting GitHub API
	myCustomHttpClient := httpClient(token)
	client := github.NewClient(myCustomHttpClient)

	// list all repositories for the authenticated user
	repos, _, err := client.Repositories.List("", nil)

	w.Write([]byte("\n\nHere's your repositories list:\n"))
	for i := 0; i < len(repos); i++ {
		w.Write([]byte(*repos[i].FullName + "\n"))
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<a href=/login>login</a>"))
}

type Transport struct {
	Token string
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if t.Token != "" {
		req.Header.Set("Authorization", "token " + t.Token)
	}

	resp, err = http.DefaultTransport.RoundTrip(req)

	return
}

func httpClient(token string) *http.Client {
	t := &Transport{
		Token: token,
	}
	return &http.Client{Transport: t}
}
