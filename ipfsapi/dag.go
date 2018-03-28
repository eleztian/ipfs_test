package ipfsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ipfs/go-ipfs-cmdkit/files"
)

func (c *Client) DagGet(ref string, out interface{}) error {
	req := c.newRequest(context.Background(), "dag/get")
	req.Args = []string{ref}

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	defer resp.Close()

	return json.NewDecoder(resp.Output).Decode(out)
}

func (c *Client) DagPut(data interface{}, ienc, kind string) (string, error) {
	req := c.newRequest(context.Background(), "dag/put")
	req.Opts = map[string]string{
		"input-enc": ienc,
		"format":    kind,
	}

	var r io.Reader
	switch data := data.(type) {
	case string:
		r = strings.NewReader(data)
	case []byte:
		r = bytes.NewReader(data)
	case io.Reader:
		r = data
	default:
		return "", fmt.Errorf("cannot current handle putting values of type %T", data)
	}
	rc := ioutil.NopCloser(r)
	fr := files.NewReaderFile("", "", rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)
	req.Body = fileReader

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	defer resp.Close()

	var out struct {
		Cid struct {
			Target string `json:"/"`
		}
	}
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Cid.Target, nil
}
