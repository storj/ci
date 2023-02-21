// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/zeebo/errs"
	"go.uber.org/zap"
)

// Client contains all the information to call jenkins instances.
type Client struct {
	Username string
	Token    string
	log      *zap.Logger
}

// NewClient creates a new client instance.
func NewClient(log *zap.Logger, username string, token string) Client {
	return Client{
		Username: username,
		Token:    token,
		log:      log,
	}
}

// TriggerJob calls the right Jenkins API endpoint to trigger a new job.
func (c *Client) TriggerJob(ctx context.Context, job string, parameters map[string]string) error {
	fields := make([]zap.Field, 0)
	for k, v := range parameters {
		fields = append(fields, zap.String(k, v))
	}

	fields = append(fields, zap.String("job", job))

	c.log.Info("Triggering jenkins build", fields...)
	var params []string
	for k, v := range parameters {
		params = append(params, k+"="+v)
	}
	triggerURL := fmt.Sprintf("https://build.dev.storj.io/job/%s/buildWithParameters?%s", job, strings.Join(params, "="))
	return c.jenkinsHTTPCall(ctx, triggerURL, nil)
}

func (c *Client) jenkinsHTTPCall(ctx context.Context, url string, result interface{}) error {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return errs.Wrap(err)
	}
	httpRequest.SetBasicAuth(c.Username, c.Token)
	c.log.Info("Executing HTTP request", zap.String("url", url), zap.String("user", c.Username))

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return errs.Wrap(err)
	}

	if httpResponse.StatusCode >= 300 {
		return errs.New("couldn't get gerrit message from %s, code: %d", url, httpResponse.StatusCode)
	}

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if result != nil {
		err = json.Unmarshal(body, result)
		if err != nil {
			return errs.Wrap(err)
		}
	}

	return nil
}
