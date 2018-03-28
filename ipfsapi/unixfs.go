package ipfsapi

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
)

type UnixLsObject struct {
	Hash  string
	Size  uint64
	Type  string
	Links []*UnixLsLink
}

type UnixLsLink struct {
	Hash string
	Name string
	Size uint64
	Type string
}

type lsOutput struct {
	Objects map[string]*UnixLsObject
}

// FileList entries at the given path using the UnixFS commands
func (c *Client) FileList(path string) (*UnixLsObject, error) {
	resp, err := c.newRequest(context.Background(), "file/ls", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out lsOutput
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}
	for _, v := range out.Objects {
		return v, nil
	}
	return nil, errors.New("no object in results")
}
