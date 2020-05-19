package main

import (
	"reflect"
	"testing"
)

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
