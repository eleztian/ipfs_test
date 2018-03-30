package main

import (
	"context"

	"github.com/ipfs/go-ipfs/core"
	"os"
	"tab/ipfs_test/fs"
	"tab/ipfs_test/fuse/cgofuse"
	"tab/ipfs_test/ipfs"
)

func main() {
	ctx, cancled := context.WithCancel(context.Background())
	defer cancled()

	ch := make(chan interface{})
	go ipfs.StartDaemon(ctx, ch)
	node := (<-ch).(*core.IpfsNode)
	defer func() {
		node.Close()
	}()

	hellofs := &fs.SuFS{
		IPFSc:ipfs.ContextIPFS{
			Ctx:ctx,
			Node:node,
		},
		Objects:make(map[string]*ipfs.LinkInfo),
	}
	host := cgofuse.NewFileSystemHost(hellofs)
	host.Mount("", os.Args[1:])
	//coreapi.NewCoreAPI(node)
	//
	//merkleNode, _ := core.Resolve(ctx, node.Namesys, node.Resolver, path.Path("QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"))
	//ndpb, ok := merkleNode.(*merkledag.ProtoNode)
	//if !ok {
	//	//res.SetError(merkledag.ErrNotProtobuf, cmdkit.ErrNormal)
	//	return
	//}
	//
	//unixFSNode, err := unixfs.FromBytes(ndpb.Data())
	//if err != nil {
	//
	//}
	//t := unixFSNode.GetType()
	//fmt.Println(t)
	//for _, link := range merkleNode.Links() {
	//	linkNode, err := link.GetNode(ctx, node.DAG)
	//	if err != nil {
	//		break
	//	}
	//	lnpb, ok := linkNode.(*merkledag.ProtoNode)
	//	if !ok {
	//		return
	//	}
	//	d, err := unixfs.FromBytes(lnpb.Data())
	//	t := d.GetType()
	//	fmt.Println(link.Name, link.Cid.String(), t.String(), link.Size, d.GetFilesize())
	//
	//}
	//r, err := api.Object().Links(ctx, coreapi.ResolvedPath("QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv", nil, nil))
	//if err != nil {
	//	panic(err)
	//}
	//for
	//fmt.Println(r)

}
