// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zeebo/errs"
)

var gerritBaseURL = "https://review.dev.storj.io"

// main is a binary which can be copied to gerrit's hooks directory and can act based on the give parameters.
func main() {

	// arguments are defined by gerrit hook system, usually (but not only) --key value about the build
	argMap := map[string]string{}
	for p := 1; p < len(os.Args); p++ {
		if len(os.Args) > p && !strings.HasPrefix(os.Args[p+1], "--") {
			argMap[os.Args[p][2:]] = os.Args[p+1]
			p++
		}
	}

	if path.Base(os.Args[0]) == "patchset-created" || os.Getenv("GERRIT_HOOK_ACTION") == "patchset-created" {
		err := patchsetCreated(context.Background(), argMap, postGithubComment)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "couldn't execute hook %+v\n", err.Error())
		}
	}
}

// postGithubComment adds a new comment to a github issue.
func postGithubComment(ctx context.Context, orgRepo string, issue string, message string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", orgRepo, issue)
	request := map[string]string{
		"body": message,
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return errs.Wrap(err)
	}
	err = callGithubAPIV3(ctx, "POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return err
	}
	return nil
}

// patchsetCreated handles incoming hook call by gerrit for patchset-created events.
func patchsetCreated(ctx context.Context, argMap map[string]string, postComment func(ctc context.Context, orgRepo string, issue string, message string) error) error {
	patchset, err := strconv.Atoi(argMap["patchset"])
	if err != nil {
		return errs.New("given patchset id is not a number: %s", argMap["patchset"])
	}
	message, err := getGerritMessage(ctx, argMap["change"], patchset)
	if err != nil {
		return err
	}
	previousMessage := ""
	if patchset > 0 {
		previousMessage, err = getGerritMessage(ctx, argMap["change"], patchset-1)
		if err != nil {
			return err
		}
	}

	currentRefs := findGithubRefs(message)
	oldRefs := findGithubRefs(previousMessage)
	newRefs := subtractRefs(currentRefs, oldRefs)

	for _, ref := range newRefs {
		if ref.repo == "" {
			ref.repo = argMap["project"]
		}
		comment := fmt.Sprintf("Change %s mentions this issue.", argMap["change-url"])
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

// getGerritMessage retrieves the last long commit message of a gerrit patch.
func getGerritMessage(ctx context.Context, changesetID string, patchset int) (string, error) {
	url := fmt.Sprintf("%s/changes/%s/revisions/%d/commit", gerritBaseURL, changesetID, patchset)
	httpRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", errs.Wrap(err)
	}
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return "", errs.Wrap(err)
	}

	if httpResponse.StatusCode >= 300 {
		return "", errs.New("couldn't get gerrit message from %s, code: %d", url, httpResponse.StatusCode)
	}
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", errs.Wrap(err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	var jsonContent map[string]interface{}
	// XSSI prevention chars are removed here
	err = json.Unmarshal(body[5:], &jsonContent)
	if err != nil {
		return "", errs.Wrap(err)
	}

	return jsonContent["message"].(string), nil
}

// getToken retrieves the GITHUB_TOKEN for API usage.
func getToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", errs.Wrap(err)
	}
	configFile := filepath.Join(configDir, "gerrit-hook", "github-token")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return "", fmt.Errorf("token is not defined neither with GITHUB_TOKEN nor in %s file", configFile)
	}
	tokenBytes, err := os.ReadFile(configFile)
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(bytes.TrimSpace(tokenBytes)), nil
}

// callGithubAPIV3 is a wrapper around the HTTP method call.
func callGithubAPIV3(ctx context.Context, method string, url string, body io.Reader) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errs.Wrap(err)
	}
	token, err := getToken()
	if err != nil {
		return errs.Wrap(err)
	}
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return errs.Wrap(err)
	}

	if resp.StatusCode > 299 {
		return errs.Combine(errs.New("%s url is failed (%s): %s", method, resp.Status, url), resp.Body.Close())
	}
	return resp.Body.Close()
}
