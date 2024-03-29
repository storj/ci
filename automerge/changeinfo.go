// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/zeebo/errs"
)

// ChangeInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-info.
type ChangeInfo struct {
	Id              string                       `json:"id"`
	ChangeId        string                       `json:"change_id"`
	Project         string                       `json:"project"`
	Status          string                       `json:"status"`
	Created         string                       `json:"created"`
	Updated         string                       `json:"updated"`
	Submittable     bool                         `json:"submittable"`
	Mergeable       bool                         `json:"mergeable"`
	Number          int                          `json:"_number"`
	SubmitRecords   []SubmitRecordInfo           `json:"submit_records"`
	CurrentRevision string                       `json:"current_revision"`
	Related         []RelatedChangeAndCommitInfo `json:"changes"`
	Messages        []ChangeMessageInfo
	Labels          map[string]ApprovalInfo
}

// SubmitRecordInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#submit-record-info.
type SubmitRecordInfo struct {
	RuleName string                  `json:"rule_name"`
	Status   string                  `json:"status"`
	Labels   []SubmitRecordInfoLabel `json:"labels"`
}

// SubmitRecordInfoLabel corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#submit-record-info-label.
type SubmitRecordInfoLabel struct {
	Label  string `json:"label"`
	Status string `json:"status"`
}

// ActionInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#action-info.
type ActionInfo struct {
	Method  string `json:"method"`
	Label   string `json:"label"`
	Title   string `json:"title"`
	Enabled bool   `json:"enabled"`
}

// RelatedChangeAndCommitInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#related-change-and-commit-info.
type RelatedChangeAndCommitInfo struct {
	Project         string     `json:"project"`
	ChangeId        string     `json:"change_id"`
	Commit          CommitInfo `json:"commit"`
	Change          int        `json:"_change_number"`
	Revision        int        `json:"_revision_number"`
	CurrentRevision int        `json:"_current_revision_number"`
	Status          string     `json:"status"`
}

// CommitInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#commit-info.
type CommitInfo struct {
	Commit  string       `json:"commit"`
	Parents []CommitInfo `json:"parents"`
	Subject string       `json:"subject"`
}

// ChangeMessageInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-message-info.
type ChangeMessageInfo struct {
	Id         string      `json:"id"`
	Tag        string      `json:"tag"`
	Author     AccountInfo `json:"author"`
	RealAuthor AccountInfo `json:"real_author"`
	Date       gerritTime  `json:"date"`
	Message    string      `json:"message"`
	Revision   int         `json:"_revision_number"`
}

// ApprovalInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#approval-info.
type ApprovalInfo struct {
	All   []ApprovalInfoVote `json:"all"`
	Value int                `json:"value"`
}

// ApprovalInfoVote corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#approval-info-vote.
type ApprovalInfoVote struct {
	AccountId int        `json:"_account_id"`
	Date      gerritTime `json:"date"`
	Value     int        `json:"value"`
	Tag       string     `json:"tag"`
}

// AccountInfo corresponds to https://gerrit-review.googlesource.com/Documentation/rest-api-accounts.html#account-info.
type AccountInfo struct {
	Id int `json:"_account_id"`
}

type gerritTime time.Time

// UnmarshalJSON implements json unmarshaler.
func (g *gerritTime) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errs.New("invalid date: %q", data)
	}
	t, err := time.Parse("2006-01-02 15:04:05.000000000", string(data[1:len(data)-1]))
	if err != nil {
		return err
	}
	*g = gerritTime(t)
	return nil
}

// CanRebase checks whether it makes sense to rebase this commit.
func (ci *ChangeInfo) CanRebase() bool {
	return ci.HasNoParents() && (ci.Verified() || (!ci.HasVerified() && ci.NotBuilding()))
}

// HasNoParents checks whether this has any uncommitted parents.
func (ci *ChangeInfo) HasNoParents() bool {
	parents := make(map[string]struct{})
	for _, rel := range ci.Related {
		if rel.Commit.Commit == ci.CurrentRevision {
			for _, p := range rel.Commit.Parents {
				parents[p.Commit] = struct{}{}
			}
		}
	}
	for _, rel := range ci.Related {
		if _, ok := parents[rel.Commit.Commit]; ok {
			if rel.Status == "NEW" {
				return false
			}
		}
	}
	return true
}

// CanMerge checks whether the change can be submitted.
func (ci *ChangeInfo) CanMerge() bool {
	return true &&
		ci.HasNoParents() &&
		ci.Verified() &&
		ci.Submittable &&
		ci.Mergeable
}

// HasVerified checks whether the change has the verified labels.
func (ci *ChangeInfo) HasVerified() bool {
	for _, rec := range ci.SubmitRecords {
		if rec.RuleName != "gerrit~PrologRule" && rec.RuleName != "gerrit~DefaultSubmitRule" {
			continue
		}
		for _, lab := range rec.Labels {
			if lab.Label == "Verified" && lab.Status != "NEED" {
				return true
			}
		}
	}

	for _, vote := range ci.Labels["Verified"].All {
		if vote.Value == -1 || vote.Value == 2 {
			return true
		}
	}

	return false
}

// Verified checks whether the conditions for verifications have been met.
func (ci *ChangeInfo) Verified() bool {
	// if it has verified as a submit record, then it's verified
	for _, rec := range ci.SubmitRecords {
		if rec.RuleName != "gerrit~PrologRule" && rec.RuleName != "gerrit~DefaultSubmitRule" {
			continue
		}
		for _, lab := range rec.Labels {
			if lab.Label == "Verified" && lab.Status == "OK" {
				return true
			}
		}
	}

	mostPositive := 0
	for _, vote := range ci.Labels["Verified"].All {
		if vote.Value < 0 {
			return false
		}
		if vote.Value > mostPositive {
			mostPositive = vote.Value
		}
	}

	// if the last build is over an hour hold and we have a single
	// +1 vote, then assume that it failed.
	if lastBuild := ci.LatestBuildStarted(); mostPositive == 1 &&
		(!lastBuild.IsZero() && time.Since(lastBuild) > time.Hour) {
		return false
	}

	return mostPositive >= 2
}

// NotBuilding checks whether the current change is being built by jenkins at the moment.
func (ci *ChangeInfo) NotBuilding() bool {
	if ci.HasVerified() {
		return true
	}
	if time.Since(ci.LatestBuildStarted()) > 1*time.Hour {
		return true
	}
	return false
}

func (ci *ChangeInfo) infoURL(kind string) string {
	return "changes/" + ci.Id + "/" + kind
}

func (ci *ChangeInfo) actionURL(action string) string {
	return "changes/" + ci.Id + "/revisions/" + ci.CurrentRevision + "/" + action
}

// ViewURL returns the link for seeing a specific commit.
func (ci *ChangeInfo) ViewURL(base string) string {
	return base + "c/" + ci.Project + "/+/" + strconv.Itoa(ci.Number)
}

// LatestBuildStarted returns the most recent build starting time.
func (ci *ChangeInfo) LatestBuildStarted() time.Time {
	largestRev := 0
	var latest time.Time
	for _, msg := range ci.Messages {
		switch {
		case false,

			// new style messages
			strings.HasPrefix(msg.Tag, "autogenerated:gerrit-integration") &&
				strings.Contains(msg.Message, "triggering build"),

			// old style messages
			strings.HasPrefix(msg.Tag, "autogenerated:jenkins-gerrit-trigger") &&
				strings.Contains(msg.Message, "Build Started"):

			if msg.Revision > largestRev {
				largestRev = msg.Revision
				latest = time.Time(msg.Date)
			} else if msg.Revision == largestRev && time.Time(msg.Date).After(latest) {
				latest = time.Time(msg.Date)
			}
		}
	}

	return latest
}
