// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package github

import (
	"net/http"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/zeebo/errs"
)

// Authorization implements authorization for http.Requests.
type Authorization interface {
	Add(req *http.Request) error
}

// Token implements transport for adding Personal Access Token.
type Token string

// RoundTrip adds authorization header to the request.
func (pat Token) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "token "+string(pat))
	return http.DefaultTransport.RoundTrip(req)
}

// LoadInstanceTransport creates a new instance authentication.
func LoadInstanceTransport(appID, instanceID, privateKeyPEMPath string) (http.RoundTripper, error) {
	aid, err := strconv.ParseInt(appID, 10, 64)
	if err != nil {
		return nil, errs.New("invalid appID %q", appID)
	}
	iid, err := strconv.ParseInt(instanceID, 10, 64)
	if err != nil {
		return nil, errs.New("invalid instanceID %q", instanceID)
	}

	transport, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, aid, iid, privateKeyPEMPath)
	if err != nil {
		return nil, errs.New("failed to transport: %w", err)
	}

	return transport, nil
}
