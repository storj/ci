// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/storj/ci/gerrit-hook/gerrit"
	"github.com/storj/ci/gerrit-hook/github"
	"github.com/storj/ci/gerrit-hook/jenkins"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
)

// main is a binary which can be copied to gerrit's hooks directory and can act based on the give parameters.
func main() {

	cfg := zap.NewDevelopmentConfig()

	// directory to collect events for debug
	logDir := "/tmp/gerrit-hook-log"
	if _, err := os.Stat(logDir); err == nil {
		cfg.OutputPaths = append(cfg.OutputPaths, path.Join(logDir, "hook.log"))
	}

	log, _ := cfg.Build()

	viper.SetConfigName("config")
	viper.AddConfigPath(path.Join(path.Base(os.Args[0])))
	viper.AddConfigPath("$HOME/.gerrit-hook")
	viper.AddConfigPath("$HOME/.config/gerrit-hook")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Warn("Reading configuration files are failed. Hope you use environment variables (JENKINS_USER, JENKINS_TOKEN, GITHUB_TOKEN)", zap.Error(err))
	}

	viper.SetConfigName("gerrit-hook")

	j := jenkins.NewClient(log.Named("jenkins"), viper.GetString("jenkins-user"), viper.GetString("jenkins-token"))

	g := github.NewClient(log.Named("github"), viper.GetString("github-token"))

	gr := gerrit.NewClient(log.Named("gerrit"), viper.GetString("gerrit-token"))

	// arguments are defined by gerrit hook system, usually (but not only) --key value about the build
	argMap := map[string]string{}
	for p := 1; p < len(os.Args); p++ {
		if len(os.Args) > p && !strings.HasPrefix(os.Args[p+1], "--") {
			argMap[os.Args[p][2:]] = os.Args[p+1]
			p++
		}
	}

	// directory to collect events for debug
	debugDir := "/tmp/gerrit-hook-debug"
	if _, err := os.Stat(debugDir); err == nil {
		filename := fmt.Sprintf("%s-%d.txt", time.Now().Format("20060102-150405"), rand.Int())
		err := os.WriteFile(path.Join(debugDir, filename), []byte(strings.Join(os.Args, "\n")), 0644)
		if err != nil {
			log.Error("couldn't write out debug information", zap.Error(err))
		}
	}
	// binary is symlinked to site/hooks under the name of default hook name:
	// https://gerrit.googlesource.com/plugins/hooks/+/HEAD/src/main/resources/Documentation/config.md
	action := path.Base(os.Args[0])

	// helping local development
	if os.Getenv("GERRIT_HOOK_ARGFILE") != "" {
		argMap, action, err = readArgFile(os.Getenv("GERRIT_HOOK_ARGFILE"))
		if err != nil {
			log.Error("couldn't write out debug information", zap.Error(err))
		}
	}

	log.Debug("Hook is called",
		zap.String("action", action),
		zap.String("project", argMap["project"]),
		zap.String("change", argMap["change"]),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	switch action {
	case "patchset-created":
		// github comment sync is enabled for all projects
		err := github.AddComment(ctx, gr, argMap["project"], argMap["change"], argMap["commit"], argMap["change-url"], argMap["patchset"], g.PostGithubComment)
		if err != nil {
			log.Error("Couldn't add github PR comment", zap.Error(err))
		}

		if enabledProject(argMap["project"]) {
			err = jenkins.TriggeredByAnyChange(ctx, log, j, gr, argMap["project"], argMap["change"], argMap["commit"])
			if err != nil {
				log.Error("Couldn't trigger new jenkins build", zap.Error(err))
			}
		}
	case "comment-added":
		if enabledProject(argMap["project"]) {
			err := jenkins.TriggeredByComment(ctx, log, j, gr, argMap["project"], argMap["change"], argMap["commit"], argMap["comment"])
			if err != nil {
				log.Error("Couldn't trigger new jenkins build", zap.Error(err))
			}
		}
	case "ref-updated":
		if enabledProject(argMap["project"]) {
			// in case of wip -> ready / ready -> wip state change, newrev=refs/changes/02/7902/meta

			parts := strings.Split(argMap["refname"], "/")
			if len(parts) == 5 && parts[4] == "meta" {
				change, err := gr.GetChange(context.Background(), parts[3])
				if err != nil {
					log.Error("ref-updated event but change couldn't be found", zap.Error(err))
				}

				err = jenkins.TriggeredByAnyChange(ctx, log, j, gr, change.Project, change.ChangeID, change.CurrentRevision)
				if err != nil {
					log.Error("Couldn't trigger new jenkins build", zap.Error(err))
				}
			}
		}
	default:
		// we are not interested about other type of hooks, even if they are delivered.
	}

}

func enabledProject(s string) bool {
	for _, k := range viper.GetStringSlice("projects") {
		if k == s {
			return true
		}
	}
	return false
}

func readArgFile(fileName string) (argMap map[string]string, action string, err error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return argMap, action, errs.Wrap(err)
	}

	argMap = make(map[string]string)
	action = ""
	key := ""
	value := ""
	for _, line := range strings.Split(string(content), "\n") {
		if action == "" {
			action = path.Base(line)
			continue
		}
		if strings.HasPrefix(line, "--") {
			if key != "" {
				argMap[key] = value
			}
			key = line[2:]
			value = ""
		} else {
			if value != "" {
				value += "\n"
			}
			value += line
		}
	}
	if key != "" {
		argMap[key] = value
	}
	return argMap, action, nil
}
