package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/markbates/goth/gothic"
)

// Check if the user is logged in, otherwise forward to login page.
func rootHandler(res http.ResponseWriter, req *http.Request) {
	s, _ := gothic.Store.Get(req, "uaa-proxy-session")
	if s.Values["logged"] != true {
		http.Redirect(res, req, "/auth/cloudfoundry", http.StatusTemporaryRedirect)
		return
	}

	newProxy(s.Values["user_email"].(string)).ServeHTTP(res, req)
}

// Handle callbacks from oauth.
func callbackHandler(res http.ResponseWriter, req *http.Request) {

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Fprintln(res, err)
		return
	}

	s, err := gothic.Store.Get(req, "uaa-proxy-session")
	s.Values["user_email"] = user.Email
	s.Values["logged"] = true
	gothic.Store.Save(req, res, s)

	http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
}

// This is a bit of nastiness: because of how route services work
// you have to read the url that is calling the service from a
// request header.
// This can probably be switched to change the request before hitting mux.
func fakeRouter(res http.ResponseWriter, req *http.Request) {
	forwardedURL := req.Header.Get(CF_FORWARDED_URL)
	if forwardedURL != "" {
		url, _ := url.Parse(forwardedURL)

		switch url.Path {
		case "/auth/cloudfoundry/callback":
			req.URL.RawQuery = "provider=cloudfoundry&" + url.RawQuery + req.URL.RawQuery
			callbackHandler(res, req)
		case "/auth/cloudfoundry":
			req.URL.RawQuery = "provider=cloudfoundry&" + url.RawQuery + req.URL.RawQuery

			setProviders("https://" + url.Host + "/auth/cloudfoundry/callback")

			gothic.BeginAuthHandler(res, req)
		default:
			rootHandler(res, req)
		}
	} else {
		rootHandler(res, req)
	}
}
