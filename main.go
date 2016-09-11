package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/cloudfoundry"
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

	rtr.HandleFunc("/{rest:.*}", fakeRouter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	loggedRouter := handlers.LoggingHandler(os.Stdout, rtr)

	err := http.ListenAndServe(":"+port, loggedRouter)
	if err != nil {
		fmt.Println(err)
	}
}

func newProxy(remote_user string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			forwardedURL := req.Header.Get(CF_FORWARDED_URL)
			url, err := url.Parse(forwardedURL)
			if err != nil {
				log.Fatalln(err.Error())
			}
			req.URL = url
			req.Host = url.Host
			req.Header.Add("X-Auth-User", remote_user)

			fmt.Println(req.Header)
		},
	}
	return proxy
}

func setProviders(callbackURL string) {
	goth.UseProviders(
		cloudfoundry.New(c.UAAUrl, c.ClientKey, c.ClientSecret, callbackURL, "openid"),
	)
}
