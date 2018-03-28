package ipfsapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ipfs/go-ipfs-cmdkit/files"
	"github.com/whyrusleeping/tar-utils"
)

type Object struct {
	Name string
	Hash string
	Size string
}

func (c *Client) Add(reader io.Reader, name string) (*Object, error) {
	return c.AddWithOpt(reader, name, true, false)
}

func (c *Client) AddNoPin(reader io.Reader, name string) (*Object, error) {
	return c.AddWithOpt(reader, name, false, false)
}

func (c *Client) AddWithOpt(r io.Reader, name string, pin bool, rawLeaves bool) (*Object, error) {
	var rc io.ReadCloser
	if rClose, ok := r.(io.ReadCloser); ok {
		rc = rClose
	} else {
		rc = ioutil.NopCloser(r)
	}
	path, fileName := filepath.Split(name)
	// handler expects an array of files
	fr := files.NewReaderFile(fileName, path, rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)

	req := NewRequest(context.Background(), c.url, "add")
	req.Body = fileReader
	req.Opts["progress"] = "false"
	if !pin {
		req.Opts["pin"] = "false"
	}

	if rawLeaves {
		req.Opts["raw-leaves"] = "true"
	}

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}

	var out Object
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil

}

// Cat the content at the given path. Callers need to drain and close the returned reader after usage.
func (c *Client) Cat(path string) (io.ReadCloser, error) {
	resp, err := NewRequest(context.Background(), c.url, "cat", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

func (c *Client) AddDir(dir string) ([]Object, error) {
	stat, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	dirPath, dirNmae := filepath.Split(dir)
	sf, err := files.NewSerialFile(dirPath, dir, false, stat)
	if err != nil {
		return nil, err
	}
	slf := files.NewSliceFile(dirNmae, dir, []files.File{sf})
	reader := files.NewMultiFileReader(slf, true)

	req := NewRequest(context.Background(), c.url, "add")
	req.Opts["r"] = "true"
	req.Body = reader

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}

	dec := json.NewDecoder(resp.Output)
	var output = make([]Object, 1)
	for {
		var out Object
		err = dec.Decode(&out)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		output = append(output, out)
	}

	if len(output) == 0 {
		return nil, errors.New("no results received")
	}

	return output, nil

}

// List 列出文件夹的内容 path = /ipfs/<dir-Hash>
func (c *Client) List(path string) ([]*Link, error) {
	resp, err := NewRequest(context.Background(), c.url, "ls", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out struct{ Objects []LsObject }
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return out.Objects[0].Links, nil
}

// Get 获取hash下的所有文件保存到outDir中
func (c *Client) Get(hash, outDir string) error {
	resp, err := c.newRequest(context.Background(), "get", hash).Send(c.httpCli)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	defer resp.Close()

	extractor := &tar.Extractor{Path: outDir}
	return extractor.Extract(resp.Output)
}

func (c *Client) AddLink(target string) (*Object, error) {
	path, name := filepath.Split(target)
	link := files.NewLinkFile(name, path, target, nil)
	slf := files.NewSliceFile("", "", []files.File{link})
	reader := files.NewMultiFileReader(slf, true)

	req := c.newRequest(context.Background(), "add")
	req.Body = reader

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}

	var out Object
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
