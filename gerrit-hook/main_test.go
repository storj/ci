// Copyright (C) 2024 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/storj/ci/gerrit-hook/gerrit"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestJenkinsInstancesConfig(t *testing.T) {
	viper.Reset()
	configBytes := bytes.NewBufferString(`
jenkins:
   public:
      url: test
   private:
      url: test2
`)
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(configBytes)
	instances := jenkinsInstances()

	require.NoError(t, err)

	require.Len(t, instances, 2)
	require.Equal(t, "test", instances["public"].URL)
	require.Equal(t, "test2", instances["private"].URL)
}

func TestReadTriggerConfig(t *testing.T) {
	t.Skip("Real integration test. Requires Gerrit user")
	log := zaptest.NewLogger(t)
	gr := gerrit.NewClient(log, os.Getenv("GERRIT_BASEURL"), os.Getenv("GERRIT_USER"), os.Getenv("GERRIT_TOKEN"))
	config := ReadTriggerConfig(context.Background(), log, gr, "storj/storj")
	fmt.Println(config)
	require.Equal(t, "storj-gerrit-verify", config.Verify)
	require.Equal(t, "storj-gerrit-premerge", config.PreMerge)
	require.Equal(t, "public", config.Jenkins)
}
