// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import "testing"

func TestSanitizeString(t *testing.T) {
	// The tests below are written based solely on the function's documentation comment
	// which specifies that the function should:
	//  - replace unsupported characters with underscores,
	//  - suppress (remove) leading and trailing underscores,
	//  - convert uppercase to lowercase and separate words with underscores (camelCase -> snake_case),
	//    taking acronyms into account (e.g. DNSResolution -> dns_resolution),
	//  - and produce a string that matches the regexp: ^[a-z0-9]+(_[a-z0-9]+)*$
	//
	// Each test case below declares the expected output according to those rules.
	tests := []struct {
		name  string
		in    string
		want  string
		valid bool
	}{
		{"empty", "", "", false},
		{"already valid lowercase", "simple", "simple", true},
		{"space -> underscore (inside)", "foo bar", "foo_bar", false},
		{"leading/trailing unsupported trimmed", "-foo-", "foo", false},
		{"camelCase -> snake_case", "fooBar", "foo_bar", false},
		{"uppercase at end", "fooB", "foo_b", false},
		{"acronym then word", "DNSResolution", "dns_resolution", false},
		{"acronym with separator", "DNS-Resolution", "dns_resolution", false},
		{"numbers preserved and camel boundary", "foo123Bar", "foo123_bar", false},
		{"multiple consecutive unsupported collapsed", "a--b", "a_b", false},
		{"existing underscore in middle preserved", "foo_bar", "foo_bar", true},
		{"leading underscore removed", "_foo", "foo", false},
		{"multiple leading underscore removed", "___foo", "foo", false},
		{"trailing underscore removed", "foo_", "foo", false},
		{"multiple trailing underscore removed", "foo____", "foo", false},
		{"leading and trailing underscore removed", "_foo_", "foo", false},
		{"multiple leading and trailing underscore removed", "___foo______", "foo", false},
		{"all underscores -> empty", "_____", "", false},
		{"all unsupported -> empty", "----", "", false},
		{"non-ascii replaced/removed and trimmed", "ümlaut", "mlaut", false},
		{"emoji treated as unsupported", "a😊b", "a_b", false},
		{"HTTP + Request -> http_request", "HTTPRequest", "http_request", false},
		{"mixed acronym and words", "someHTTPServer", "some_http_server", false},
		{"two-letter acronym", "ID", "id", false},
		{"single capitalized word", "Foo", "foo", false},
		{"punctuation between words", "a.b,c", "a_b_c", false},
		{"XMLHttpRequest complex acronym", "XMLHttpRequest", "xml_http_request", false},
		{"underscore then uppercase then underscore", "_A_B_", "a_b", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, valid := sanitizeString(tc.in)
			if got != tc.want || valid != tc.valid {
				t.Fatalf("sanitizeString(%q) = %q, %t; want %q, %t", tc.in, got, valid, tc.want, tc.valid)
			}
		})
	}
}
