package ipfsapi

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"

	"bytes"
	"fmt"
	"strings"

	"github.com/ipfs/go-ipfs-cmdkit/files"
)

// LsObject 文件夹列表
type LsObject struct {
	Hash  string
	Links []*Link
}

// Link 文件链接信息
type Link struct {
	Name string
	Hash string
	Size uint64
	Type FType
}

// FType 文件类型
type FType int

const (
	TRaw FType = iota
	TDirectory
	TFile
	TMetadata
	TSymlink
)

type PinType string

const (
	DirectPin    PinType = "direct"
	RecursivePin         = "recursive"
	IndirectPin          = "indirect"
)

type PinInfo struct {
	Type PinType
}

// PatchRmLink 在文件夹中删除一个子对象,返回新的对象hash
func (c *Client) PatchRmLink(root string, args ...string) (*LsObject, error) {
	cmdArgs := append([]string{root}, args...)
	resp, err := c.newRequest(context.Background(), "object/patch/rm-link", cmdArgs...).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	dec := json.NewDecoder(resp.Output)
	var out LsObject
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// PatchAddLink 在文件夹中添加一个子对象,返回新的对象hash
func (c *Client) PatchAddLink(root, path, childHash string, create bool) (*LsObject, error) {
	cmdargs := []string{root, path, childHash}

	req := c.newRequest(context.Background(), "object/patch/add-link", cmdargs...)
	if create {
		req.Opts["create"] = "true"
	}

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	var out LsObject
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// PatchData 设置或者添加数据到root对象
func (c *Client) PatchData(root string, set bool, name string, data interface{}) (*LsObject, error) {
	var read io.Reader
	switch d := data.(type) {
	case io.Reader:
		read = d
	case []byte:
		read = bytes.NewReader(d)
	case string:
		read = strings.NewReader(d)
	default:
		return nil, fmt.Errorf("unrecognized type: %#v", data)
	}

	cmd := "append-data"
	if set {
		cmd = "set-data"
	}
	path, fName := filepath.Split(name)
	fr := files.NewReaderFile(fName, path, ioutil.NopCloser(read), nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)

	req := c.newRequest(context.Background(), "object/patch/"+cmd, root)
	req.Body = fileReader

	resp, err := req.Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out LsObject
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// NewObject 创建一个新的对象, 返回这个对象的信息
func (c *Client) NewObject(template string) (*LsObject, error) {
	args := []string{}
	if template != "" {
		args = []string{template}
	}

	resp, err := c.newRequest(context.Background(), "object/new", args...).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out LsObject
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

type ObjectStat struct {
	NumLinks       int // 文件数
	BlockSize      int
	LinksSize      int
	DataSize       int
	CumulativeSize int // 累积大小
	Hash           string
}

// ResolvePath 获取以 . 命名的DAG节点的统计数据。
func (c *Client) ResolvePath(path string) (*ObjectStat, error) {
	resp, err := c.newRequest(context.Background(), "object/stat", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var out ObjectStat
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}


type IPFSObject struct {
	Links []Object
	Data  string
}

func (c *Client) ObjectGet(path string) (*IPFSObject, error) {
	resp, err := c.newRequest(context.Background(), "object/get", path).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	var obj IPFSObject
	err = json.NewDecoder(resp.Output).Decode(&obj)
	if err != nil {
		return nil, err
	}

	return &obj, nil
}

func (c *Client) ObjectPut(obj *IPFSObject) (*Object, error) {
	data := new(bytes.Buffer)
	err := json.NewEncoder(data).Encode(obj)
	if err != nil {
		return nil, err
	}

	rc := ioutil.NopCloser(data)

	fr := files.NewReaderFile("", "", rc, nil)
	slf := files.NewSliceFile("", "", []files.File{fr})
	fileReader := files.NewMultiFileReader(slf, true)

	req := c.newRequest(context.Background(), "object/put")
	req.Body = fileReader
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

// ObjectStat gets stats for the DAG object named by key. It returns
// the stats of the requested Object or an error.
func (c *Client) ObjectStat(key string) (*ObjectStat, error) {
	resp, err := c.newRequest(context.Background(), "object/stat", key).Send(c.httpCli)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	stat := &ObjectStat{}

	err = json.NewDecoder(resp.Output).Decode(stat)
	if err != nil {
		return nil, err
	}

	return stat, nil
}