package ipfsapi

import (
	"context"
	"encoding/json"
	"errors"
)

type PeerInfo struct {
	ID    string
	Addrs []string
}

func (c *Client) FindPeer(peer string) (*PeerInfo, error) {
	resp, err := c.newRequest(context.Background(), "dht/findpeer", peer).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out struct{ Responses []*PeerInfo }
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	if len(out.Responses) == 0 {
		return nil, errors.New("peer not found ")
	}

	return out.Responses[0], nil
}
