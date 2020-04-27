package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMutateErrors(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(handleRoot))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.NoError(t, err)
}
