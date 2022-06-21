// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package gerrit

// ReviewInput is a REST request for adding reviews.
type ReviewInput struct {
	Message string `json:"message"`
	Tag     string `json:"tag"`
}

// Changes a list of gerrit patches.
type Changes []Change

// Change is a descriptor for a gerrit patch.
type Change struct {
	Number           int           `json:"_number"`
	AttentionSet     struct{}      `json:"attention_set"`
	Branch           string        `json:"branch"`
	ChangeID         string        `json:"change_id"`
	Created          string        `json:"created"`
	Deletions        int           `json:"deletions"`
	HasReviewStarted bool          `json:"has_review_started"`
	Hashtags         []interface{} `json:"hashtags"`
	ID               string        `json:"id"`
	Insertions       int           `json:"insertions"`
	Mergeable        bool          `json:"mergeable"`
	MetaRevID        string        `json:"meta_rev_id"`
	WorkInProgress   bool          `json:"work_in_progress"`
	Private          bool          `json:"is_private"`
	Owner            struct {
		AccountID int `json:"_account_id"`
	} `json:"owner"`
	Project      string        `json:"project"`
	Requirements []interface{} `json:"requirements"`
	Status       string        `json:"status"`
	Subject      string        `json:"subject"`
	Revisions    map[string]struct {
		Number  int    `json:"_number"`
		Created string `json:"created"`
		Fetch   struct {
			AnonymousHTTP struct {
				Ref string `json:"ref"`
				URL string `json:"url"`
			} `json:"anonymous http"`
		} `json:"fetch"`
		Kind     string `json:"kind"`
		Ref      string `json:"ref"`
		Uploader struct {
			AccountID int `json:"_account_id"`
		} `json:"uploader"`
	} `json:"revisions"`
	SubmitRecords []struct {
		Labels []struct {
			AppliedBy struct {
				AccountID int `json:"_account_id"`
			} `json:"applied_by,omitempty"`
			Label  string `json:"label"`
			Status string `json:"status"`
		} `json:"labels"`
		RuleName string `json:"rule_name"`
		Status   string `json:"status"`
	} `json:"submit_records"`
	SubmitType             string `json:"submit_type"`
	TotalCommentCount      int    `json:"total_comment_count"`
	UnresolvedCommentCount int    `json:"unresolved_comment_count"`
	Updated                string `json:"updated"`
	Labels                 map[string]struct {
		All []struct {
			AccountID string `json:"account_id"`
			Value     int    `json:"value"`
		} `json:"all"`
	} `json:"labels"`
	Messages        []Message `json:"messages"`
	CurrentRevision string    `json:"current_revision"`
}

// Message is a comment added to review (eg. build started...).
type Message struct {
	ID             string `json:"id"`
	Tag            string `json:"tag"`
	Message        string `json:"message"`
	RevisionNumber int    `json:"_revision_number"`
}

// Commit represents a commit message.
type Commit struct {
	Message string `json:"message"`
}
