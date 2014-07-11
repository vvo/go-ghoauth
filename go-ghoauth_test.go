package ghoauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	client_id     = "82c041166d565cf2f149"
	client_secret = "6458b989c06d8106f468fd47106d8654e3887d5b"
	redirect_uri  = "http://localhost:8080/callback"
	scope         = "user,repo:status"
	ghcode        = "d21kld21ldk21"
	ghresponse    = `{"access_token":"e72e16c7e42f292c6912e7710c838347ae178b4a", "scope":"user,repo:status", "token_type":"bearer"}`
	access_token  = "e72e16c7e42f292c6912e7710c838347ae178b4a"
)

func TestLogin(t *testing.T) {
	f := New(&Config{
		ClientId:     client_id,
		ClientSecret: client_secret,
		RedirectUri:  redirect_uri,
		Scope:        scope,
	})

	r, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	f.Login(w, r)

	if w.Code != 307 {
		t.Error("Redirect code must be 307")
	}

	redirect, _ := url.Parse(w.Header().Get("Location"))

	if redirect.Host != "github.com" {
		t.Error("It must redirects to github.com")
	}

	if redirect.Scheme != "https" {
		t.Error("Must use https")
	}

	if redirect.Path != "/login/oauth/authorize" {
		t.Error("Github oauth path does not matches")
	}

	q := redirect.Query()

	if q.Get("client_id") != client_id {
		t.Error("Bad `client_id` in redirect querystring")
	}

	if q.Get("redirect_uri") != redirect_uri {
		t.Error("Bad `redirect_uri` in redirect querystring")
	}

	if q.Get("scope") != scope {
		t.Error("Bad `scope` in redirect querystring")
	}

	if len(q.Get("state")) != 36 {
		t.Error("Bad `state` in redirect querystring")
	}

	// test state uniqness between calls
	r, _ = http.NewRequest("GET", "/", nil)
	f.Login(w, r)
	redirect, _ = url.Parse(w.Header().Get("Location"))

	if q.Get("state") == redirect.Query().Get("state") {
		t.Error("States should differ between calls")
	}
}

func TestCallback(t *testing.T) {
	f := New(&Config{
		ClientId:     client_id,
		ClientSecret: client_secret,
		RedirectUri:  redirect_uri,
		Scope:        scope,
	})

	// fake github response server
	fakeGithub := func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("client_id") != client_id {
			t.Error("provided client_id differs")
		}

		if r.FormValue("client_secret") != client_secret {
			t.Error("provided client_secret differs")
		}

		if r.FormValue("code") != ghcode {
			t.Error("provided code differs from github")
		}

		w.Write([]byte(ghresponse))
	}

	ts := httptest.NewServer(http.HandlerFunc(fakeGithub))
	defer ts.Close()

	f.BaseUrl = ts.URL

	// step one, /login
	r, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	f.Login(w, r)
	redirectUrl, _ := url.Parse(w.Header().Get("Location"))
	generatedState := redirectUrl.Query().Get("state")

	// construct github fake redirect request to our server
	callbackUrl, _ := url.Parse("/callback")
	q := callbackUrl.Query()
	q.Set("state", "bad state")
	q.Set("code", ghcode)

	callbackUrl.RawQuery = q.Encode()

	r, _ = http.NewRequest("GET", callbackUrl.String(), nil)
	_, err := f.Callback(r)

	// state differs from what we know, error
	if err == nil {
		t.Error("We must get an error on token mismatch")
	}

	q.Set("state", generatedState)

	callbackUrl.RawQuery = q.Encode()

	r, _ = http.NewRequest("GET", callbackUrl.String(), nil)
	token, err := f.Callback(r)

	if err != nil {
		t.Error(err)
	}

	if token != access_token {
		t.Error("Token mismatch")
	}
}
