package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	externalnode "github.com/ProxeusApp/node-go"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	e := echo.New()

	// test core
	testCore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err, "Cannot read body")
		r.Body.Close()

		n := externalnode.ExternalNode{}
		err = json.Unmarshal(body, &n)
		assert.Nil(t, err, "Cannot unmarshal body")

		assert.EqualValues(t, "http://example.com", n.Url, "Wrong service url")
		assert.EqualValues(t, "service", n.Name, "Wrong service name")
		assert.EqualValues(t, "service", n.ID, "Wrong service name")
		assert.EqualValues(t, "topsecret", n.Secret, "Wrong secret")
		assert.EqualValues(t, serviceDetail, n.Detail, "Wrong service detail")
	}))
	defer testCore.Close()

	// test target
	testTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err, "Cannot read body")
		r.Body.Close()

		content := map[string]string{}

		err = json.Unmarshal(body, &content)
		assert.Nil(t, err, "Cannot unmarshal body")

		assert.EqualValues(t, "Hello", content["test1"], "Wrong content")
		io.WriteString(w, string(body))
	}))
	defer testTarget.Close()

	h := &handler{
		proxeusURL:  testCore.URL,
		serviceName: "service",
		serviceUrl:  "http://example.com",
		jwtSecret:   "topsecret",
		targetURL:   testTarget.URL,
		headers: [][]string{
			{
				"foo", "bar",
			},
		},
	}

	err := h.register()
	if err != nil {
		t.Errorf("Expected not error but got %s", err.Error())
	}

	req := httptest.NewRequest(http.MethodPost, "/next", strings.NewReader(`{"test1":"Hello"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	if assert.NoError(t, h.next(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		result, err := ioutil.ReadAll(rec.Body)
		assert.Nil(t, err, "Cannot read result body")
		assert.Equal(t, `{"test1":"Hello"}`, string(result))
	}
}

func TestExtractHeaders(t *testing.T) {

	tests := []struct {
		title   string
		env     []string
		headers [][]string
	}{
		{
			title: "nil",
		},
		{
			title: "none",
			env: []string{
				"FOOBAR=foobar",
				"",
				"BAR=foo",
			},
		},
		{
			title: "mixed",
			env: []string{
				"FOOBAR=foobar",
				"",
				"JSON_SENDER_HEADER_foobar=ABC123",
				"BAR=foo",
				"JSON_SENDER_HEADER_barfoo=321CBA",
			},
			headers: [][]string{
				{"foobar", "ABC123"},
				{"barfoo", "321CBA"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			headers := extractHeaders(test.env)
			if !reflect.DeepEqual(headers, test.headers) {
				t.Errorf("Expected %v but got %v", test.headers, headers)
			}
		})
	}
}
