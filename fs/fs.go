package fs

import (
	"fmt"
	"path/filepath"
	"tab/ipfs_test/fuse/cgofuse"
	"tab/ipfs_test/ipfs"

	"io"
	"io/ioutil"
	"strings"

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
	if path == "/" || path == "" {
		stat.Mode = cgofuse.S_IFDIR | 0555
		return 0
	}
	path = getPath(path)
	i, ok := sf.Objects[path]
	var itype unixfs_pb.Data_DataType
	if !ok {
		fmt.Println("do not exit", path)
		itype, _ = sf.IPFSc.GetType(path)
	} else {
		fmt.Println("--find", path)
		itype = i.Type
		stat.Size = int64(i.Size)
		fmt.Println("Size: ", i.Size)
		stat.Hash = i.Hash
	}
	switch itype {
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

	return 0
}

func (sf *SuFS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	fmt.Println("Read................")
	path = getPath(path)
	fmt.Println(path)
	r, err := sf.IPFSc.Cat(path)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	fmt.Println(r.Size())
	r.Seek(ofst, io.SeekStart)
	r2 := io.LimitReader(r, int64(len(buff)))
	b, err := ioutil.ReadAll(r2)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
	return copy(buff, b)
}

func getPath(p string) string {
	if p == "/" {
		return examplesHash
	}
	p = filepath.Join(examplesHash, p)
	return strings.Replace(p, "\\", "/", -1)
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
		tt := filepath.Join(path, v.Name)
		tt = strings.Replace(tt, "\\", "/", -1)
		sf.Objects[tt] = v
		fill(v.Name, &cgofuse.Stat_t{
			Size: int64(v.Size),
			Hash: v.Hash,
		}, 0)
	}

	fill(filename, nil, 0)
	return 0
}

var _ cgofuse.FileSystemInterface = (*SuFS)(nil)
