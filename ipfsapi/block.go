package ipfsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/ipfs/go-ipfs-cmdkit/files"
)

type BlockStat struct {
	Key  string
	Size int
}

func (c *Client) BlockStat(path string) (*BlockStat, error) {
	resp, err := c.newRequest(context.Background(), "block/stat", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var inf BlockStat

	err = json.NewDecoder(resp.Output).Decode(&inf)
	if err != nil {
		return nil, err
	}

	return &inf, nil
}

func (c *Client) BlockGet(path string) ([]byte, error) {
	resp, err := c.newRequest(context.Background(), "block/get", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	return ioutil.ReadAll(resp.Output)
}

func (c *Client) BlockPut(block []byte) (string, error) {
	data := bytes.NewReader(block)
	rc := ioutil.NopCloser(data)
	fr := files.NewReaderFile("", "", rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)

	req := c.newRequest(context.Background(), "block/put")
	req.Body = fileReader
	resp, err := req.Send(c.httpCli)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out struct {
		Key string
	}
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Key, nil
}


