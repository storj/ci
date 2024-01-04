// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/zeebo/errs"
)

// Client implements a minimal gerrit client.
type Client struct {
	base string
	user string
	pass string
}

// GetChangeInfos returns changes that can be potentially be automatically submitted.
func (c *Client) GetChangeInfos(ctx context.Context) (cis []*ChangeInfo, err error) {
	q := strings.Join([]string{
		"status:open",
		"label:Code-Review=2,count>=2",
		"-label:Code-Review<=-1",
		"-has:unresolved",
		"-is:wip",
	}, "+")

	err = c.query(ctx, "GET", "changes/?q="+q+"&o=CURRENT_REVISION&o=SUBMITTABLE&o=MESSAGES&o=LABELS", &cis)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	for _, ci := range cis {
		err = c.query(ctx, "GET", ci.actionURL("related"), &ci)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		err = c.query(ctx, "GET", ci.infoURL("messages"), &ci.Messages)
		if err != nil {
			return nil, errs.Wrap(err)
		}
	}

	cis = filter(cis, func(ci *ChangeInfo) bool {
		return !ci.HasVerified() || ci.Verified()
	})

	sort.Slice(cis, func(i, j int) bool {
		return cis[i].Number < cis[j].Number
	})

	return cis, nil
}

// Rebase rebases the specified change.
func (c *Client) Rebase(ctx context.Context, ci *ChangeInfo) error {
	return c.query(ctx, "POST", ci.actionURL("rebase"), nil)
}

// Submit submits the specified change.
func (c *Client) Submit(ctx context.Context, ci *ChangeInfo) error {
	return c.query(ctx, "POST", ci.actionURL("submit"), nil)
}

func (c *Client) query(ctx context.Context, method, endpoint string, into interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, c.base+"a/"+endpoint, nil)
	if err != nil {
		return errs.Wrap(err)
	}
	req.SetBasicAuth(c.user, c.pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errs.New("invalid status code %q: %d", endpoint, resp.StatusCode)
	}

	if into != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return errs.Wrap(err)
		}

		const skipPrefix = ")]}'\n"
		data = bytes.TrimPrefix(data, []byte(skipPrefix))

		if err := json.Unmarshal(data, into); err != nil {
			return errs.Wrap(err)
		}
	}

	return nil
}
