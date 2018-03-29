/*
 * hellofs.go
 *
 * Copyright 2017 Bill Zissimopoulos
 */
/*
 * This file is part of Cgofuse.
 *
 * It is licensed under the MIT license. The full license text can be found
 * in the License.txt file at the root of this project.
 */

package main

import (
	"os"

	"tab/ipfs_test/fs"
	"tab/ipfs_test/ipfsapi"
	"tab/ipfs_test/fuse/cgofuse"
)

func main() {
	hellofs := &fs.SuFS{IpfsClient:ipfsapi.NewClient(ipfsapi.IPFS_API_URL)}
	host := cgofuse.NewFileSystemHost(hellofs)
	host.Mount("", os.Args[1:])
}