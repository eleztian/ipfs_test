package ipfs

import (
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/merkledag"
	"context"
	"github.com/ipfs/go-ipfs/unixfs"
	"github.com/ipfs/go-ipfs/unixfs/pb"
	"github.com/ipfs/go-ipfs/path"

	"github.com/pkg/errors"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/ipfs/go-ipfs/unixfs/io"
)

type ContextIPFS struct {
	Ctx context.Context
	Node *core.IpfsNode
}

type LinkInfo struct {
	Name, Hash string
	Size       uint64
	Type       unixfs_pb.Data_DataType
}

type ObjectInfo struct {
	Self *LinkInfo
	Links []*LinkInfo
}

func (c *ContextIPFS) GetDirLinksInfo(pathH string) (res *ObjectInfo, err error){
	merkleNode, _ := core.Resolve(c.Ctx, c.Node.Namesys, c.Node.Resolver, path.Path(pathH))
	ndpb, ok := merkleNode.(*merkledag.ProtoNode)
	if !ok {
		return
	}

	unixFSNode, err := unixfs.FromBytes(ndpb.Data())
	if err != nil {
		return
	}

	fType := unixFSNode.GetType()

	res = &ObjectInfo{Self:&LinkInfo{}}
	res.Self.Type = fType
	res.Self.Size = unixFSNode.GetFilesize()
	res.Self.Hash = merkleNode.String()
	switch fType {
	case unixfs_pb.Data_Directory:
		links := make([]*LinkInfo,0)
		for _, link := range merkleNode.Links() {
			linkNode, err := link.GetNode(c.Ctx, c.Node.DAG)
			if err != nil {
				break
			}
			lnpb, ok := linkNode.(*merkledag.ProtoNode)
			if !ok {
				return nil, errors.New("can not change linkNode to ProtoNode")
			}
			d, err := unixfs.FromBytes(lnpb.Data())
			l := &LinkInfo{
				Name:link.Name,
				Hash:link.Cid.String(),
				Size:d.GetFilesize(),
				Type:d.GetType(),
			}
			links = append(links, l)
		}
		res.Links = links
	}
	return
}

func (c *ContextIPFS) GetType(pathH string) (res unixfs_pb.Data_DataType, err error) {
	merkleNode, _ := core.Resolve(c.Ctx, c.Node.Namesys, c.Node.Resolver, path.Path(pathH))
	ndpb, ok := merkleNode.(*merkledag.ProtoNode)
	if !ok {
		return -1, errors.New("can not change merkleNode to ProtoNode")
	}
	unixFSNode, err := unixfs.FromBytes(ndpb.Data())
	if err != nil {
		return -1, err
	}
	res = unixFSNode.GetType()
	return
}


func (c *ContextIPFS) Cat(path string) (io.DagReader, error) {
	return coreunix.Cat(c.Ctx, c.Node, path)
	//if max == 0 {
	//	return nil, 0, nil
	//}
	//read, err :=
	//if err != nil {
	//	return nil, 0, err
	//}
	//if offset > int64(read.Size()) {
	//	return nil, read.Size(), errors.New("no more data to read")
	//}
	//count, err := read.Seek(offset, io.SeekStart)
	//if err != nil {
	//	return nil, 0, err
	//}
	//offset = 0
	//
	//size := uint64(read.Size() - uint64(count))
	//length += size
	//if max > 0 && length >= uint64(max) {
	//	var r io.Reader = read
	//	if overshoot := int64(length - uint64(max)); overshoot != 0 {
	//		r = io.LimitReader(read, int64(size)-overshoot)
	//		length = uint64(max)
	//	}
	//	readers = append(readers, r)
	//	break
	//}
	//readers = append(readers, read)
	//
	//return readers, length, nil
}