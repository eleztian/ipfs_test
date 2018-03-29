package fs

import (
	"fmt"
	"tab/ipfs_test/fuse/cgofuse"
	"tab/ipfs_test/ipfsapi"
	"path/filepath"
)

const (
	filename = "hello"
	contents = "hello, world\n"
)

var host = "http://localhost:5001"
var examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"

type SuFS struct {
	cgofuse.FileSystemBase
	IpfsClient *ipfsapi.Client
}

func (self *SuFS) Open(path string, flags int) (errc int, fh uint64) {
	path, name := filepath.Split(path)
	if name == "" {
		return 0,0
	}

	//switch path {
	//case "/" + filename:
	//	return 0, 0
	//default:
	//	return -cgofuse.ENOENT, ^uint64(0)
	//}
}

func (self *SuFS) Getattr(path string, stat *cgofuse.Stat_t, fh uint64) (errc int) {
	path, name := filepath.Split(path)
	if name == "" {
		stat.Mode = cgofuse.S_IFDIR | 0555
		return 0
	}
	r, err := self.IpfsClient.List(examplesHash+path)
	if err != nil {
		return 0
	}
	stati := &ipfsapi.Link{}
	for _, v := range r {
		if v.Name == name {
			fmt.Println(v)
			stati = v
			break
		}
	}

	if stati == nil {
		return -cgofuse.EACCES
	}
	switch stati.Type {
	case ipfsapi.TDirectory:
		stat.Mode = cgofuse.S_IFDIR | 0555
	case ipfsapi.TFile:
		stat.Mode = cgofuse.S_IFREG | 0444
	case ipfsapi.TMetadata:
		stat.Mode = cgofuse.S_IFMT | 0444
	case ipfsapi.TSymlink:
		stat.Mode = cgofuse.S_IFLNK | 0444
	default:
		stat.Mode = cgofuse.S_IFREG | 0444
	}
	stat.Size = int64(stati.Size)
	stat.Hash = stati.Hash
	return 0
}

func (self *SuFS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	endofst := ofst + int64(len(buff))

	r, err := self.IpfsClient.BlockGet()

	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst])
	return
}

func (self *SuFS) Readdir(path string,
	fill func(name string, stat *cgofuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)
	if path == "/" {
		path = examplesHash
	}
	r, err := self.IpfsClient.List(path)
	if err != nil {
		return -cgofuse.ENOENT
	}
	for _, v := range r {
		fmt.Println(v)
		fill(v.Name, nil, 0)
	}

	fill(filename, nil, 0)
	return 0
}

var _ cgofuse.FileSystemInterface = (*SuFS)(nil)
