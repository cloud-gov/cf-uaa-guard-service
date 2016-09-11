package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/markbates/goth/gothic"
)

const (
	CF_FORWARDED_URL = "X-Cf-Forwarded-Url"
)

type Config struct {
	CookieSecret       string `envconfig:"cookie_secret"`
	DefaultCallbackUrl string `envconfig:"default_callback_url"`
	UAAUrl             string `envconfig:"uaa_url"`
	ClientKey          string `envconfig:"client_key"`
	ClientSecret       string `envconfig:"client_secret"`
}

var c Config

func init() {
	err := envconfig.Process("guard", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	gothic.Store = sessions.NewCookieStore([]byte(c.CookieSecret))
}

func main() {
	setProviders(c.DefaultCallbackUrl)

	rtr := mux.NewRouter()

	rtr.HandleFunc("/auth/{provider}/callback", callbackHandler)
	rtr.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler)
	rtr.HandleFunc("/{rest:.*}", rootHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	proxyRouter := UrlChanger(rtr)

	loggedRouter := handlers.LoggingHandler(os.Stdout, proxyRouter)

	err := http.ListenAndServe(":"+port, loggedRouter)
	if err != nil {
		fmt.Println(err)
	}
}

// Set url based on the CF_FORWARDED_URL Header
func UrlChanger(h http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		forwardedURL := req.Header.Get(CF_FORWARDED_URL)
		if forwardedURL != "" {
			url, _ := url.Parse(forwardedURL)
			req.URL = url

			// Add provider for goth, need to find a better way
			if strings.HasPrefix(url.Path, "/auth") {
				req.URL.RawQuery = "provider=cloudfoundry&" + req.URL.RawQuery
			}
		}
		h.ServeHTTP(res, req)
	}

	return http.HandlerFunc(fn)
}
