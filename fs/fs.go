package fs

import (
	"fmt"
	"path/filepath"
	"tab/ipfs_test/fuse/cgofuse"
	"tab/ipfs_test/ipfs"

	"github.com/ipfs/go-ipfs/unixfs/pb"
)

const (
	filename = "hello"
	contents = "hello, world\n"
)

var host = "http://localhost:5001"
var examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"

type SuFS struct {
	cgofuse.FileSystemBase
	IPFSc   ipfs.ContextIPFS
	Objects map[string]*ipfs.LinkInfo
}

func (self *SuFS) Open(path string, flags int) (errc int, fh uint64) {
	path, name := filepath.Split(path)
	if name == "" {
		return 0, 0
	}

	//switch path {
	//case "/" + filename:
	//	return 0, 0
	//default:
	//	return -cgofuse.ENOENT, ^uint64(0)
	//}
	return 0, 0
}

func (sf *SuFS) Getattr(path string, stat *cgofuse.Stat_t, fh uint64) (errc int) {
	if path == "/" || path == ""{
		stat.Mode = cgofuse.S_IFDIR | 0555
		return 0
	}
	path = getPath(path)
	i, ok := sf.Objects[path]
	if !ok {
		fmt.Println("----", path)
		stat.Mode = cgofuse.S_IFDIR | 0555
		return 0
	}

	switch i.Type {
	case unixfs_pb.Data_Directory:
		stat.Mode = cgofuse.S_IFDIR | 0555
	case unixfs_pb.Data_File:
		stat.Mode = cgofuse.S_IFREG | 0444
	case unixfs_pb.Data_Metadata:
		stat.Mode = cgofuse.S_IFMT | 0444
	case unixfs_pb.Data_Symlink:
		stat.Mode = cgofuse.S_IFLNK | 0444
	default:
		stat.Mode = cgofuse.S_IFREG | 0444
	}
	stat.Size = int64(i.Size)
	stat.Hash = i.Hash
	return 0
}

func (self *SuFS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	endofst := ofst + int64(len(buff))

	//r, err := self.IpfsClient.BlockGet()

	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst])
	return
}

func getPath(p string) string {
	return examplesHash + p
}

func (sf *SuFS) Readdir(path string,
	fill func(name string, stat *cgofuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)

	path = getPath(path)

	r, err := sf.IPFSc.GetDirLinksInfo(path)
	if err != nil {
		fmt.Println(err)
		return -cgofuse.ENOENT
	}
	sf.Objects[path] = r.Self
	for _, v := range r.Links {
		fmt.Println(v)
		fmt.Println(filepath.Join(path, v.Name))
		sf.Objects[filepath.Join(path, v.Name)] = v
		fill(v.Name, nil, 0)
	}

	fill(filename, nil, 0)
	return 0
}

var _ cgofuse.FileSystemInterface = (*SuFS)(nil)
