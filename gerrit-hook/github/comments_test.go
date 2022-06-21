// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findGithubRef(t *testing.T) {

	assertGithubRef := func(t *testing.T, expected []githubRef, text string) bool {
		refs := findGithubRefs(text)
		assert.EqualValues(t, expected, refs)
		return true
	}

	t.Run("Short reference", func(t *testing.T) {
		assertGithubRef(t, []githubRef{{"", "123"}}, `
This is a commit

Github #123!
`)
	})

	t.Run("Full URL", func(t *testing.T) {
		assertGithubRef(t, []githubRef{{"storj/common", "5616"}}, `
https://github.com/storj/common/issues/5616
`)
	})

	t.Run("Full URL inline", func(t *testing.T) {
		assertGithubRef(t, []githubRef{{"storj/sandbox", "1"}}, `no web change

https://github.com/storj/sandbox/issues/1
Change-Id: Iec94f295143f973a7e54bc5c8cf187dbc3eb2b64

`)
	})

	t.Run("Short id with org/repo", func(t *testing.T) {
		assertGithubRef(t, []githubRef{{"storj/common", "5616"}}, "foo bar storj/common#5616")
	})

	t.Run("Without any reference", func(t *testing.T) {
		assertGithubRef(t, nil, "How are you?")
	})

	t.Run("Multiple", func(t *testing.T) {
		assertGithubRef(t, []githubRef{
			{"", "123"},
			{"storj/common", "5616"},
		}, `
Github #123!
https://github.com/storj/common/issues/5616
`)
	})
}
