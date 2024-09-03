package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	config := Config{}
	srv := NewServer(config)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	srv.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestPrefixOfInternalRoutes(t *testing.T) {
	prefix := "__@@@__"
	config := Config{
		InternalRoutePrefix: prefix,
	}
	srv := NewServer(config)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/%s/ping", prefix),
		nil,
	)
	srv.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
