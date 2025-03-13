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

// Client is a Gerrit Rest client.
type Client struct {
	baseurl string
	user    string
	token   string
	log     *zap.Logger
}

// NewClient creates a new Gerrit REST client.
func NewClient(log *zap.Logger, baseurl, user, token string) Client {
	return Client{
		baseurl: baseurl,
		user:    user,
		token:   token,
		log:     log,
	}
}

func (g *Client) doAPICall(ctx context.Context, url string, request interface{}) ([]byte, error) {
	var requestBody io.Reader
	method := "GET"
	if request != nil {
		method = "POST"

		c, err := json.Marshal(request)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		requestBody = strings.NewReader(string(c))
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	httpRequest.SetBasicAuth(g.user, g.token)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode >= 300 {
		return nil, errs.New("couldn't get gerrit message from %s, code: %d, %s", url, httpResponse.StatusCode, body)
	}
	return body, nil
}

func (g *Client) doJsonAPICall(ctx context.Context, url string, request interface{}, result interface{}) error {
	body, err := g.doAPICall(ctx, url, request)
	if err != nil {
		return errs.Wrap(err)
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

// GetCommit retrieves the commit of a gerrit patch.
func (g *Client) GetCommit(ctx context.Context, changesetID string, commit string) (Commit, error) {
	c := Commit{}
	url := fmt.Sprintf("%s/a/changes/%s/revisions/%s/commit", g.baseurl, changesetID, commit)
	err := g.doJsonAPICall(ctx, url, nil, &c)
	if err != nil {
		return Commit{}, err
	}
	return c, nil
}

// GetContent returns with a specific version of a file.
func (g *Client) GetContent(ctx context.Context, project string, ref string, file string) (string, error) {
	url := fmt.Sprintf("%s/a/projects/%s/branches/%s/files/%s/content", g.baseurl, url.QueryEscape(project), url.QueryEscape(ref), url.QueryEscape(file))
	body, err := g.doAPICall(ctx, url, nil)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// AddReview adds a new review (comment + vote) to a change.
func (g *Client) AddReview(ctx context.Context, changesetID string, revision string, comment string, tag string) error {
	i := ReviewInput{
		Message: comment,
		Tag:     tag,
	}
	url := fmt.Sprintf("%s/a/changes/%s/revisions/%s/review", g.baseurl, changesetID, revision)
	err := g.doJsonAPICall(ctx, url, &i, nil)
	return err
}

// QueryChanges search for changes based on search expression.
func (g *Client) QueryChanges(ctx context.Context, condition string) (Changes, error) {
	c := Changes{}
	url := fmt.Sprintf("%s/a/changes/?q=%s", g.baseurl, url.QueryEscape(condition))
	err := g.doJsonAPICall(ctx, url, nil, &c)
	return c, err
}

// GetChange returns with one change based on identifier.
func (g *Client) GetChange(ctx context.Context, change string) (Change, error) {
	c := Change{}
	url := fmt.Sprintf("%s/a/changes/%s/?o=LABELS&o=CURRENT_REVISION&o=MESSAGES", g.baseurl, change)
	err := g.doJsonAPICall(ctx, url, nil, &c)
	return c, err
}
