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
	"encoding/json"
	"crypto/tls"
)

const (
	CF_FORWARDED_URL = "X-Cf-Forwarded-Url"
	CALLBACK_ROUTE = "/auth/cloudfoundry/callback"
)

type Config struct {
	CookieSecret       string `envconfig:"cookie_secret" required:"true"`
	VcapAppRaw         string `envconfig:"vcap_application"`
	InsecureSkipVerify bool   `envconfig:"insecure_skip_verify"`
	LoginURL           string `envconfig:"login_url" required:"true"`
	ClientKey          string `envconfig:"client_key" required:"true"`
	ClientSecret       string `envconfig:"client_secret" required:"true"`
	Port               string `envconfig:"port" default:"3000"`
	RequireSsl         bool   `envconfig:"require_ssl" default:"true"`
}
type VcapApplication struct {
	Uris []string `json:"application_uris"`
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
	callbackUrl, err := CallbackUrl(c.RequireSsl)
	if err != nil {
		log.Fatal(err.Error())
	}
	setProviders(callbackUrl)

	rtr := mux.NewRouter()

	rtr.HandleFunc("/auth/callback", callbackHandler)
	rtr.HandleFunc("/auth", authHandler)
	rtr.HandleFunc("/{rest:.*}", rootHandler)

	proxyRouter := ProxyForwardedURL(rtr)

	loggedRouter := handlers.LoggingHandler(os.Stdout, proxyRouter)

	err = http.ListenAndServe(":" + c.Port, loggedRouter)
	if err != nil {
		fmt.Println(err)
	}
}
func CallbackUrl(requireSsl bool) (string, error) {
	var vcapApp VcapApplication
	err := json.Unmarshal([]byte(c.VcapAppRaw), &vcapApp)
	if err != nil {
		return "", err
	}
	uri := vcapApp.Uris[0]
	if requireSsl {
		return "https://" + uri + CALLBACK_ROUTE, nil
	}
	return "http://" + uri + CALLBACK_ROUTE, nil
}
func DefaultHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify},
		Proxy: http.ProxyFromEnvironment,
	}
	return &http.Client{Transport: tr}
}
// Set url based on the CF_FORWARDED_URL Header
func ProxyForwardedURL(h http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		forwardedURL := req.Header.Get(CF_FORWARDED_URL)
		if forwardedURL != "" {
			req.URL, _ = url.Parse(forwardedURL)
		}
		h.ServeHTTP(res, req)
	}

	return http.HandlerFunc(fn)
}
