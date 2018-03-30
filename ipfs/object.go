package ipfs

import (
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/merkledag"
	"context"
	"github.com/ipfs/go-ipfs/unixfs"
	"github.com/ipfs/go-ipfs/unixfs/pb"
	"github.com/ipfs/go-ipfs/path"

	"github.com/pkg/errors"
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