package fs

import (
	"fmt"
	"path/filepath"
	"tab/ipfs_test/fuse/cgofuse"
	"tab/ipfs_test/ipfs"

	"io"
	"strings"

	"github.com/ipfs/go-ipfs/unixfs/pb"
)

var host = "http://localhost:5001"
var examplesHash = "QmRGNz9E1bz5toAddQmHpbE6hSmEgu8GkpG3ziMU3SsbPi"

type SuFS struct {
	cgofuse.FileSystemBase
	IPFSc   ipfs.ContextIPFS
	Objects map[string]*ipfs.LinkInfo
}

func (sf *SuFS) Open(path string, flags int) (errc int, fh uint64) {
	path, name := filepath.Split(path)
	if name == "" {
		return 0, 0
	}
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
		itype, _ = sf.IPFSc.GetType(path)
	} else {
		fmt.Println("--find", path)
		itype = i.Type
		stat.Size = int64(i.Size)
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
	path = getPath(path)
	r, err := sf.IPFSc.Cat(path)
	if err != nil {
		return cgofuse.EACCES
	}
	r.Seek(ofst, io.SeekStart)
	r2 := io.LimitReader(r, int64(len(buff)))
	if err != nil {
		return cgofuse.EACCES
	}
	n,_ = r2.Read(buff)
	return
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
	go func() {
		r, err := sf.IPFSc.GetDirLinksInfo(path)
		if err != nil {
			return
		}
		sf.Objects[path] = r.Self
		for _, v := range r.Links {
			tt := filepath.Join(path, v.Name)
			tt = strings.Replace(tt, "\\", "/", -1)
			sf.Objects[tt] = v
			fill(v.Name, &cgofuse.Stat_t{
			}, 0)
		}
	}()


	return 0
}

var _ cgofuse.FileSystemInterface = (*SuFS)(nil)
