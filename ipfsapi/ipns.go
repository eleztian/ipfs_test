package ipfsapi

import (
	"context"
	"encoding/json"
)

// Publish updates a mutable name to point to a given value
func (c *Client) Publish(node string, value string) error {
	args := []string{value}
	if node != "" {
		args = []string{node, value}
	}

	resp, err := c.newRequest(context.Background(), "name/publish", args...).Send(c.httpCli)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// Resolve 将 Resolves 提供给/ipfs/<Hash>。如果hash为空，
// 则解析改为解析节点c own/ipns值。
func (c *Client) Resolve(id string) (string, error) {
	var resp *Response
	var err error
	if id != "" {
		resp, err = c.newRequest(context.Background(), "name/resolve", id).Send(c.httpCli)
	} else {
		resp, err = c.newRequest(context.Background(), "name/resolve").Send(c.httpCli)
	}
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out struct{ Path string }
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Path, nil
}
