// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package gerrit

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCommitMessage(t *testing.T) {
	g := Client{}
	msg, err := g.GetCommitMessage(context.Background(), "storj%2Fvelero-plugin~master~I6d20b5a8605a99740834df326ad26e646eae206e", "9288388465675dd98e30f30e2575c25d3e9f8880")
	assert.NoError(t, err)
	assert.Contains(t, msg, "The commit contains almost a working")
}

func Test_addReview(t *testing.T) {
	e := os.Getenv("GERRIT_HOOK_TOKEN")
	g := Client{
		token: e,
	}
	err := g.AddReview(context.Background(), "Ic7a5aeb29a0972d43df018fdbac44256ab74e763", "1", "test comment", "")
	assert.NoError(t, err)
}
