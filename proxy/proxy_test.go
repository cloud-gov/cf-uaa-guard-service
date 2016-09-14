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
		sess.Values["logged"] = true

		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		assert.Equal(t, res.Code, http.StatusOK)

		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		assert.Equal(t, body, []byte("fred@queen.com"))
	})
}
