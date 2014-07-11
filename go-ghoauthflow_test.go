package ghoauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestLogin(t *testing.T) {
	f := New(&Config{
		ClientId:     "99cf2e4846d90af7a650",
		ClientSecret: "d9yqi4rr6jjyoeqcn405m99zj84nmj70j2x8h6ky",
		RedirectUri:  "http://yaw-dawg.com/sup",
		Scope:        "user,repo:status",
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

	if q.Get("client_id") != "99cf2e4846d90af7a650" {
		t.Error("Bad `client_id` in redirect querystring")
	}

	if q.Get("redirect_uri") != "http://yaw-dawg.com/sup" {
		t.Error("Bad `redirect_uri` in redirect querystring")
	}

	if q.Get("scope") != "user,repo:status" {
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
		ClientId:     "99cf2e4846d90af7a650",
		ClientSecret: "d9yqi4rr6jjyoeqcn405m99zj84nmj70j2x8h6ky",
		RedirectUri:  "http://yaw-dawg.com/sup",
		Scope:        "user,repo:status",
	})

	// code sent by github in response to first step
	ghCode := "d21kld21ldk21"

	// fake github response server
	fakeGithub := func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("client_id") != f.ClientId {
			t.Error("provided client_id differs")
		}

		if r.FormValue("client_secret") != f.ClientSecret {
			t.Error("provided client_secret differs")
		}

		if r.FormValue("code") != ghCode {
			t.Error("provided code differs from github")
		}

		json := `{"access_token":"e72e16c7e42f292c6912e7710c838347ae178b4a", "scope":"user,repo:status", "token_type":"bearer"}`
		w.Write([]byte(json))
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

	expectedToken := "e72e16c7e42f292c6912e7710c838347ae178b4a"

	// construct github fake redirect request to our server
	callbackUrl, _ := url.Parse("/callback")
	q := callbackUrl.Query()
	q.Set("state", "bad state")
	q.Set("code", ghCode)

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

	if token != expectedToken {
		t.Error("Token mismatch")
	}
}
