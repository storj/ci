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
	log       *zap.Logger
	instances map[string]ClientConfig
}

// ClientConfig contains all the information to connect to a specific instance.
type ClientConfig struct {
	User  string
	Token string
	URL   string
}

// NewClient creates a new client instance.
func NewClient(log *zap.Logger, instances map[string]ClientConfig) Client {
	return Client{
		log:       log,
		instances: instances,
	}
}

// TriggerJob calls the right Jenkins API endpoint to trigger a new job.
func (c *Client) TriggerJob(ctx context.Context, instance string, job string, parameters map[string]string) error {
	fields := make([]zap.Field, 0)
	for k, v := range parameters {
		fields = append(fields, zap.String(k, v))
	}

	fields = append(fields, zap.String("job", job))
	fields = append(fields, zap.String("commit", parameters["GERRIT_REF"]))
	fields = append(fields, zap.String("instance", instance))

	c.log.Info("Triggering jenkins build", fields...)
	var params []string
	for k, v := range parameters {
		params = append(params, k+"="+v)
	}
	triggerURL := fmt.Sprintf("/job/%s/buildWithParameters?%s", job, strings.Join(params, "="))
	return c.jenkinsHTTPCall(ctx, instance, triggerURL, nil)
}

func (c *Client) jenkinsHTTPCall(ctx context.Context, instance string, relURL string, result interface{}) error {
	jenkins, found := c.instances[instance]
	if !found {
		return errs.New("No such registered Jenkins instance: %s", instance)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Trim(jenkins.URL, "/")+relURL, nil)
	if err != nil {
		return errs.Wrap(err)
	}
	httpRequest.SetBasicAuth(jenkins.User, jenkins.Token)
	c.log.Info("Executing HTTP request", zap.String("url", jenkins.URL), zap.String("user", jenkins.User))

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return errs.Wrap(err)
	}

	if httpResponse.StatusCode >= 300 {
		return errs.New("couldn't trigger jenkins job %s, code: %d", jenkins.URL, httpResponse.StatusCode)
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
