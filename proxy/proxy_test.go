package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markbates/goth/gothic"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	handler := http.HandlerFunc(rootHandler)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Header.Get("X-Auth-User")))
	}))

	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	req.Header.Set(CF_FORWARDED_URL, backend.URL)

	t.Run("unauthorized", func(t *testing.T) {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusTemporaryRedirect)
		assert.Equal(t, res.Header().Get("Location"), "/auth")
	})

	t.Run("authorized", func(t *testing.T) {
		sess, err := gothic.Store.Get(req, "uaa-proxy-session")
		assert.Nil(t, err)
		sess.Values["user_email"] = "fred@queen.com"
		sess.Values["auth_token"] = "123456789"
		sess.Values["logged"] = true

		// Set some invalid value so we can be sure that it's
		// being overwritten internally.
		req.Header.Set("X-Auth-User", "auth-user-from-client")

		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusOK)

		// Be sure there's only one X-Auth-User header and
		// it's what we expect
		assert.Equal(t, len(req.Header["X-Auth-User"]), 1)
		assert.Equal(t, req.Header.Get("X-Auth-User"), "fred@queen.com")

		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		assert.Equal(t, body, []byte("fred@queen.com"))
	})
}
