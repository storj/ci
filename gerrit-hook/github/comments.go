// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/storj/ci/gerrit-hook/gerrit"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
)

// Client is a simplified REST Api client for specific Github functionalities.
type Client struct {
	client *http.Client
	log    *zap.Logger
}

// NewClient creates a new Github REST client.
func NewClient(log *zap.Logger, client *http.Client) Client {
	return Client{
		client: client,
		log:    log,
	}
}

// AddComment handles incoming hook call by gerrit for patchset-created events.
func AddComment(ctx context.Context, gr gerrit.Client, project string, change string, commit string, changeURL string, patchset string, postComment func(ctc context.Context, orgRepo string, issue string, message string) error) error {

	fullCommit, err := gr.GetCommit(ctx, change, commit)
	if err != nil {
		return err
	}

	previousFullCommit := gerrit.Commit{}

	// The first patchset is numbered 1, hence we shouldn't request the previous message for it.
	// This only happens on initial push.
	if patchset != "" && patchset != "0" && patchset != "1" {
		p, err := strconv.Atoi(patchset)
		if err != nil {
			return err
		}

		previousFullCommit, err = gr.GetCommit(ctx, change, strconv.Itoa(p-1))
		if err != nil {
			return err
		}
	}

	currentRefs := findGithubRefs(fullCommit.Message)
	oldRefs := findGithubRefs(previousFullCommit.Message)
	newRefs := subtractRefs(currentRefs, oldRefs)

	for _, ref := range newRefs {
		if ref.repo == "" {
			ref.repo = project
		}
		comment := fmt.Sprintf("Change [%s](%s) mentions this issue.", fullCommit.Subject, changeURL)
		if err := postComment(ctx, ref.repo, ref.issue, comment); err != nil {
			return err
		}
	}
	return nil
}

type githubRef struct {
	repo  string
	issue string
}

// findGithubRefs tries to find references to a github issues / pull request.
func findGithubRefs(message string) (refs []githubRef) {
	issuePattern := regexp.MustCompile(`([a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+)?#(\d+)`)
	urlPattern := regexp.MustCompile(`https://github.com/([a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+)/(?:pull|issues)/(\d+)`)
	for _, line := range strings.Split(message, "\n") {
		matches := issuePattern.FindStringSubmatch(line)
		if matches != nil {
			refs = append(refs, githubRef{repo: matches[1], issue: matches[2]})
		}
		matches = urlPattern.FindStringSubmatch(line)
		if matches != nil {
			refs = append(refs, githubRef{repo: matches[1], issue: matches[2]})
		}
	}
	return refs
}

func subtractRefs(currentRefs, oldRefs []githubRef) []githubRef {
	newRefs := []githubRef{}
nextRef:
	for _, current := range currentRefs {
		for _, old := range oldRefs {
			if current == old {
				continue nextRef
			}
		}
		newRefs = append(newRefs, current)
	}
	return newRefs
}

// callGithubAPIV3 is a wrapper around the HTTP method call.
func (g *Client) callGithubAPIV3(ctx context.Context, method string, url string, body io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errs.Wrap(err)
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := g.client.Do(req)
	if err != nil {
		return errs.Wrap(err)
	}

	response, _ := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > 299 {
		return errs.New("%s url is failed (%s), url: %s, response: %q", method, resp.Status, url, response)
	}
	return nil
}

// PostGithubComment adds a new comment to a github issue.
func (g *Client) PostGithubComment(ctx context.Context, orgRepo string, issue string, message string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", orgRepo, issue)
	request := map[string]string{
		"body": message,
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return errs.Wrap(err)
	}
	err = g.callGithubAPIV3(ctx, "POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return err
	}
	return nil
}
