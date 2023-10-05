// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package jenkins

import (
	"context"
	"fmt"
	"strings"

	"github.com/storj/ci/gerrit-hook/gerrit"
	"go.uber.org/zap"
)

// TriggeredByComment handles new comment event and checks if build should be triggered.
func TriggeredByComment(ctx context.Context, log *zap.Logger, jc Client, gc gerrit.Client, project string, changeID string, commit string, comment string) error {
	change, err := gc.GetChange(ctx, changeID)
	if err != nil {
		return err
	}

	var buildType string

	comment = strings.ToLower(comment)
	if strings.Contains(comment, "run jenkins verify") {
		buildType = "verify"
	} else if strings.Contains(comment, "run jenkins pre-merge") || strings.Contains(comment, "run jenkins premerge") {
		buildType = "premerge"
	} else if strings.Contains(comment, "run jenkins") {
		if change.LabelMax("Verified") == 1 {
			buildType = "premerge"
		} else {
			buildType = "verify"
		}
	} else {
		// all other comments including new reviews: check current state and trigger if build is required
		return TriggeredByAnyChange(ctx, log, jc, gc, project, changeID, comment)
	}

	if change.Private {
		return nil
	}

	err = gc.AddReview(ctx, changeID, change.CurrentRevision, fmt.Sprintf("triggering build %s...", buildType), createTag(buildType))
	if err != nil {
		return err
	}

	job := jenkinsProject(project) + "-gerrit-" + buildType
	err = jc.TriggerJob(ctx, job, map[string]string{"GERRIT_REF": commit})
	if err != nil {
		return err
	}
	return nil

}

// TriggeredByAnyChange checks if build should be triggered in case of repository is changed.
func TriggeredByAnyChange(ctx context.Context, log *zap.Logger, jc Client, gc gerrit.Client, project string, changeID string, commit string) error {
	change, err := gc.GetChange(ctx, changeID)
	if err != nil {
		return err
	}

	if change.Private || change.WorkInProgress {
		return nil
	}

	// most important: check the current state

	var buildType string
	if change.LabelMax("Verified") == 0 {
		buildType = "verify"
	}

	if change.LabelMax("Verified") == 1 && change.LabelCount("Code-Review", 2) > 1 && change.LabelMin("Code-Review") > -2 {
		buildType = "premerge"
	}

	if buildType == "" {
		return nil
	}

	if alreadyTriggered(change, buildType, change.Revisions[change.CurrentRevision].Number) {
		log.Debug("Build is already triggered", zap.String("type", buildType), zap.String("change", changeID), zap.String("revision", change.CurrentRevision))
		return nil
	}

	err = gc.AddReview(ctx, changeID, change.CurrentRevision, fmt.Sprintf("triggering build %s...", buildType), createTag(buildType))
	if err != nil {
		return err
	}

	job := jenkinsProject(project) + "-gerrit-" + buildType
	err = jc.TriggerJob(ctx, job, map[string]string{"GERRIT_REF": commit})
	if err != nil {
		return err
	}
	return nil
}

func alreadyTriggered(change gerrit.Change, buildType string, revision int) bool {
	expectedTag := createTag(buildType)
	for _, m := range change.Messages {
		if m.Tag == expectedTag && m.RevisionNumber == revision {
			return true
		}
	}
	return false
}

func createTag(buildType string) string {
	// this supposed to be the same what Jenkins uses to combine messages and show only the last one
	// content after ~ is ignored during the comparison
	return "autogenerated:gerrit-integration~" + buildType
}

func jenkinsProject(project string) string {
	parts := strings.Split(project, "/")
	return parts[len(parts)-1]
}
