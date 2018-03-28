package ipfsapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multiaddr-net"
)

type Client struct {
	url     string
	httpCli *http.Client
}

func NewClient(url string) *Client {
	c := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	if a, err := ma.NewMultiaddr(url); err == nil {
		_, host, err := manet.DialArgs(a)
		if err == nil {
			url = host
		}
	}

	return &Client{
		url:     url,
		httpCli: c,
	}
}

func (c *Client) newRequest(ctx context.Context, command string, args ...string) *Request {
	return NewRequest(ctx, c.url, command, args...)
}

type ID struct {
	AgentVersion    string
	ProtocolVersion string
	ID              string
	PublicKey       string
	Addresses       []string
}

func (c *Client) ID(peers ...string) (*ID, error) {
	if len(peers) > 1 {
		return nil, errors.New("too many peer arguments")
	}

	resp, err := c.newRequest(context.Background(), "id", peers...).Send(c.httpCli)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var output ID
	err = json.NewDecoder(resp.Output).Decode(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (c *Client) Refs(hash string, recursive bool) (<-chan string, error) {
	req := c.newRequest(context.Background(), "refs", hash)
	if recursive {
		req.Opts["r"] = "true"
	}

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	out := make(chan string)
	go func() {
		var ref struct {
			Ref string
		}
		defer resp.Close()
		defer close(out)
		dec := json.NewDecoder(resp.Output)
		for {
			err := dec.Decode(&ref)
			if err != nil {
				return
			}
			if len(ref.Ref) > 0 {
				out <- ref.Ref
			}
		}
	}()

	return out, nil
}

type Version struct {
	Version string
	Commit  string
	Repo    string
	System  string
	Golang  string
}

// returns ipfs 的版本信息
func (c *Client) Version() (*Version, error) {
	resp, err := c.newRequest(context.Background(), "version").Send(c.httpCli)
	if err != nil {
		return nil, err
	}

	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}

	ver := Version{}

	err = json.NewDecoder(resp.Output).Decode(&ver)
	if err != nil {
		return nil, err
	}

	return &ver, nil
}
