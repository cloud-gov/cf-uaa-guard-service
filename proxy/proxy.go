package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

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
	CookieSecret       string `envconfig:"cookie_secret" required:"true"`
	DefaultCallbackURL string `envconfig:"default_callback_url" required:"true"`
	LoginURL           string `envconfig:"login_url" required:"true"`
	ClientKey          string `envconfig:"client_key" required:"true"`
	ClientSecret       string `envconfig:"client_secret" required:"true"`
	Port               string `envconfig:"port" default:"3000"`
}

var c Config

func main() {
	err := envconfig.Process("guard", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	gothic.Store = sessions.NewCookieStore([]byte(c.CookieSecret))
	gothic.GetProviderName = func(*http.Request) (string, error) {
		return "cloudfoundry", nil
	}

	setProviders(c.DefaultCallbackURL)

	rtr := mux.NewRouter()

	rtr.HandleFunc("/auth/callback", callbackHandler)
	rtr.HandleFunc("/auth", authHandler)
	rtr.HandleFunc("/{rest:.*}", rootHandler)

	proxyRouter := ProxyForwardedURL(rtr)

	loggedRouter := handlers.LoggingHandler(os.Stdout, proxyRouter)

	err = http.ListenAndServe(":"+c.Port, loggedRouter)
	if err != nil {
		fmt.Println(err)
	}
}

// Set url based on the CF_FORWARDED_URL Header
func ProxyForwardedURL(h http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		forwardedURL := req.Header.Get(CF_FORWARDED_URL)
		if forwardedURL != "" {
			url, _ := url.Parse(forwardedURL)
			req.URL = url
		}
		h.ServeHTTP(res, req)
	}

	return http.HandlerFunc(fn)
}
