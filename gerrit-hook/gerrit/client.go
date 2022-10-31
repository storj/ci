// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/zeebo/errs"
	"go.uber.org/zap"
)

var gerritBaseURL = "https://review.dev.storj.io"

// Client is a Gerrit Rest client.
type Client struct {
	token string
	log   *zap.Logger
}

// NewClient creates a new Gerrit REST client.
func NewClient(log *zap.Logger, token string) Client {
	return Client{
		token: token,
		log:   log,
	}
}

func (g *Client) doAPICall(ctx context.Context, url string, request interface{}, result interface{}) error {
	var requestBody io.Reader
	method := "GET"
	if request != nil {
		method = "POST"

		c, err := json.Marshal(request)
		if err != nil {
			return errs.Wrap(err)
		}

		requestBody = strings.NewReader(string(c))
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return errs.Wrap(err)
	}
	httpRequest.SetBasicAuth("gerrit-trigger", g.token)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return errs.Wrap(err)
	}

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode >= 300 {
		return errs.New("couldn't get gerrit message from %s, code: %d, %s", url, httpResponse.StatusCode, body)
	}

	if result != nil {
		// XSSI prevention chars are removed here
		err = json.Unmarshal(body[5:], result)
		if err != nil {
			return errs.Wrap(err)
		}
	}

	return nil
}

// GetCommitMessage retrieves the last long commit message of a gerrit patch.
func (g *Client) GetCommitMessage(ctx context.Context, changesetID string, commit string) (string, error) {
	c := Commit{}
	url := fmt.Sprintf("%s/changes/%s/revisions/%s/commit", gerritBaseURL, changesetID, commit)
	err := g.doAPICall(ctx, url, nil, &c)
	if err != nil {
		return "", err
	}
	return c.Message, nil

}

// AddReview adds a new review (comment + vote) to a change.
func (g *Client) AddReview(ctx context.Context, changesetID string, revision string, comment string, tag string) error {
	i := ReviewInput{
		Message: comment,
		Tag:     tag,
	}
	url := fmt.Sprintf("%s/a/changes/%s/revisions/%s/review", gerritBaseURL, changesetID, revision)
	err := g.doAPICall(ctx, url, &i, nil)
	return err
}

// QueryChanges search for changes based on search expression.
func (g *Client) QueryChanges(ctx context.Context, condition string) (Changes, error) {
	c := Changes{}
	url := fmt.Sprintf("%s/changes/?q=%s", gerritBaseURL, url.QueryEscape(condition))
	err := g.doAPICall(ctx, url, nil, &c)
	return c, err
}

// GetChange returns with one change based on identifier.
func (g *Client) GetChange(ctx context.Context, change string) (Change, error) {
	c := Change{}
	url := fmt.Sprintf("%s/changes/%s/?o=LABELS&o=CURRENT_REVISION&o=MESSAGES", gerritBaseURL, change)
	err := g.doAPICall(ctx, url, nil, &c)
	return c, err
}
