package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/zeebo/errs"
)

type Client struct {
	base string
	user string
	pass string
}

func (c *Client) GetChangeInfos() (cis []*ChangeInfo, err error) {
	q := strings.Join([]string{
		"status:open",
		"label:Code-Review>1",
		"-label:Code-Review<0",
		"-has:unresolved",
	}, "+")

	err = c.query("GET", "changes/?q="+q+"&o=CURRENT_REVISION&o=SUBMITTABLE", &cis)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	cis = filter(cis, (*ChangeInfo).Reviewed)
	cis = filter(cis, func(ci *ChangeInfo) bool {
		return !ci.HasVerified() || ci.Verified()
	})

	for _, ci := range cis {
		err = c.query("GET", ci.actionURL("related"), &ci)
		if err != nil {
			return nil, errs.Wrap(err)
		}
	}

	sort.Slice(cis, func(i, j int) bool {
		return cis[i].Number < cis[j].Number
	})

	return cis, nil
}

func (c *Client) Rebase(ci *ChangeInfo) error {
	return c.query("POST", ci.actionURL("rebase"), nil)
}

func (c *Client) Submit(ci *ChangeInfo) error {
	return c.query("POST", ci.actionURL("submit"), nil)
}

func (c *Client) query(method, endpoint string, into interface{}) error {
	req, err := http.NewRequest(method, c.base+"a/"+endpoint, nil)
	if err != nil {
		return errs.Wrap(err)
	}
	req.SetBasicAuth(c.user, c.pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errs.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errs.New("invalid status code %q: %d", endpoint, resp.StatusCode)
	}

	if into != nil {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errs.Wrap(err)
		}

		const skipPrefix = ")]}'\n"
		data = bytes.TrimPrefix(data, []byte(skipPrefix))

		if err := json.Unmarshal(data, into); err != nil {
			return errs.Wrap(err)
		}
	}

	return nil
}
