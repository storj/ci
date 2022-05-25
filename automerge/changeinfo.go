package main

import "strconv"

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
}

type SubmitRecordInfo struct {
	RuleName string                  `json:"rule_name"`
	Status   string                  `json:"status"`
	Labels   []SubmitRecordInfoLabel `json:"labels"`
}

type SubmitRecordInfoLabel struct {
	Label  string `json:"label"`
	Status string `json:"status"`
}

type ActionInfo struct {
	Method  string `json:"method"`
	Label   string `json:"label"`
	Title   string `json:"title"`
	Enabled bool   `json:"enabled"`
}

type RelatedChangeAndCommitInfo struct {
	Project         string     `json:"project"`
	ChangeId        string     `json:"change_id"`
	Commit          CommitInfo `json:"commit"`
	Change          int        `json:"_change_number"`
	Revision        int        `json:"_revision_number"`
	CurrentRevision int        `json:"_current_revision_number"`
	Status          string     `json:"status"`
}

type CommitInfo struct {
	Commit  string       `json:"commit"`
	Parents []CommitInfo `json:"parents"`
	Subject string       `json:"subject"`
}

func (ci *ChangeInfo) CanRebase() bool {
	return ci.HasNoParents() && ci.Verified()
}

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

func (ci *ChangeInfo) CanMerge() bool {
	return true &&
		ci.HasNoParents() &&
		ci.Verified() &&
		ci.Reviewed() &&
		ci.Submittable &&
		ci.Mergeable
}

func (ci *ChangeInfo) HasVerified() bool {
	for _, rec := range ci.SubmitRecords {
		if rec.RuleName != "gerrit~PrologRule" {
			continue
		}
		for _, lab := range rec.Labels {
			if lab.Label == "Verified" && lab.Status != "NEED" {
				return true
			}
		}
	}
	return false
}

func (ci *ChangeInfo) Verified() bool {
	for _, rec := range ci.SubmitRecords {
		if rec.RuleName != "gerrit~PrologRule" {
			continue
		}
		for _, lab := range rec.Labels {
			if lab.Label == "Verified" && lab.Status == "OK" {
				return true
			}
		}
	}
	return false
}

func (ci *ChangeInfo) Reviewed() bool {
	for _, rec := range ci.SubmitRecords {
		if rec.RuleName != "gerrit~PrologRule" {
			continue
		}
		for _, lab := range rec.Labels {
			if lab.Label == "Code-Review" || lab.Label == "Code-Review-2" {
				if lab.Status == "NEED" {
					return false
				}
			}
		}
	}
	return true
}

func (ci *ChangeInfo) actionURL(action string) string {
	return "changes/" + ci.Id + "/revisions/" + ci.CurrentRevision + "/" + action
}

func (ci *ChangeInfo) ViewURL(base string) string {
	return base + "c/" + ci.Project + "/+/" + strconv.Itoa(ci.Number)
}
