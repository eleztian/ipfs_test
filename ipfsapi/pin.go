package ipfsapi

import (
	"context"
	"encoding/json"
)

// 将path对象定位到本地存储器
func (c *Client) Pin(path string) error {
	req := NewRequest(context.Background(), c.url, "pin/add", path)
	req.Opts["r"] = "true"

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// Unpin the given path
func (c *Client) Unpin(path string) error {
	req := NewRequest(context.Background(), c.url, "pin/rm", path)
	req.Opts["r"] = "true"

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// Pins 返回一个包含pinInfo的map,PinType 目前只有 DirectPin、RecursivePin和
// IndirectPin.
func (c *Client) Pins() (map[string]PinInfo, error) {
	resp, err := NewRequest(context.Background(), c.url, "pin/ls").Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	raw := struct{ Keys map[string]PinInfo }{}
	err = json.NewDecoder(resp.Output).Decode(&raw)
	if err != nil {
		return nil, err
	}

	return raw.Keys, nil
}
