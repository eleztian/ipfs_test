package ipfs

import (
	//"os"
	//
	//"tab/ipfs_test/fs"
	//"tab/ipfs_test/ipfsapi"
	//"tab/ipfs_test/fuse/cgofuse"
	"context"
	mprome "gx/ipfs/QmSTf3wJXBQk2fxdmXtodvyczrCPgJaK1B1maY78qeebNX/go-metrics-prometheus"

	"fmt"

	"github.com/ipfs/go-ipfs/core"

	logging "gx/ipfs/QmRb5jh8z2E8hMGN2tkvs1yHynUanqnZ3UeKwgN1i9P1F8/go-log"

	"github.com/ipfs/go-ipfs-cmds"
	oldcmds "github.com/ipfs/go-ipfs/commands"
	"github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	"github.com/ipfs/go-ipfs/core/corehttp"

	"gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	"gx/ipfs/QmX3QZ5jHEPidwUrymXV1iSCSUhdGxj15sm2gP4jKMef7B/client_golang/prometheus"
	"os"
	"path/filepath"

	"github.com/ipfs/go-ipfs-cmds/cli"
)

func getRepoPath() (string, error) {

	repoPath, err := fsrepo.BestKnownPath()
	if err != nil {
		return "", err
	}
	return repoPath, nil
}

func loadConfig(path string) (*config.Config, error) {
	return fsrepo.ConfigAt(path)
}

var log = logging.Logger("cmd/ipfs")

// serveHTTPApi collects options, creates listener, prints status message and starts serving requests
func serveHTTPApi(cctx *oldcmds.Context) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	apiAddr := "/ip4/127.0.0.1/tcp/5001"
	apiMaddr, err := ma.NewMultiaddr(apiAddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: invalid API address: %q (err: %s)", apiAddr, err)
	}

	apiLis, err := manet.Listen(apiMaddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: manet.Listen(%s) failed: %s", apiMaddr, err)
	}
	// we might have listened to /tcp/0 - lets see what we are listing on
	apiMaddr = apiLis.Multiaddr()
	fmt.Printf("API server listening on %s\n", apiMaddr)

	gatewayOpt := corehttp.GatewayOption(false, corehttp.WebUIPaths...)

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsOption(*cctx),
		corehttp.WebUIOption,
		gatewayOpt,
		corehttp.VersionOption(),
		corehttp.MetricsScrapingOption("/debug/metrics/prometheus"),
		corehttp.LogOption(),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.GetNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: ConstructNode() failed: %s", err)
	}

	if err := node.Repo.SetAPIAddr(apiMaddr); err != nil {
		return nil, fmt.Errorf("serveHTTPApi: SetAPIAddr() failed: %s", err)
	}

	errc := make(chan error)
	go func() {
		errc <- corehttp.Serve(node, apiLis.NetListener(), opts...)
		close(errc)
	}()
	return errc, nil
}

// serveHTTPGateway collects options, creates listener, prints status message and starts serving requests
func serveHTTPGateway(cctx *oldcmds.Context) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: GetConfig() failed: %s", err)
	}

	gatewayMaddr, err := ma.NewMultiaddr(cfg.Addresses.Gateway)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: invalid gateway address: %q (err: %s)", cfg.Addresses.Gateway, err)
	}

	writable := cfg.Gateway.Writable

	gwLis, err := manet.Listen(gatewayMaddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: manet.Listen(%s) failed: %s", gatewayMaddr, err)
	}
	// we might have listened to /tcp/0 - lets see what we are listing on
	gatewayMaddr = gwLis.Multiaddr()

	if writable {
		fmt.Printf("Gateway (writable) server listening on %s\n", gatewayMaddr)
	} else {
		fmt.Printf("Gateway (readonly) server listening on %s\n", gatewayMaddr)
	}

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("gateway"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsROOption(*cctx),
		corehttp.VersionOption(),
		corehttp.IPNSHostnameOption(),
		corehttp.GatewayOption(writable, "/ipfs", "/ipns"),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.GetNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: ConstructNode() failed: %s", err)
	}

	errc := make(chan error)
	go func() {
		errc <- corehttp.Serve(node, gwLis.NetListener(), opts...)
		close(errc)
	}()
	return errc, nil
}

func StartDaemon(ctx context.Context, ch chan interface{}) {
	err := mprome.Inject()
	if err != nil {
		log.Errorf("Injecting prometheus handler for metrics failed with message: %s\n", err.Error())
	}

	env, err := buildEnv(ctx)
	if err != nil {
		panic(err)
	}

	if c, ok := env.(cli.Closer); ok {
		defer c.Close()
	}

	cctx := env.(*oldcmds.Context)
	repo, err := fsrepo.Open(cctx.ConfigRoot)
	cfg, err := cctx.GetConfig()
	if err != nil {
		panic(err)
	}

	ncfg := &core.BuildCfg{
		Repo:      repo,
		Permanent: true, // It is temporary way to signify that node is permanent
		Online:    true,
		ExtraOpts: map[string]bool{
			"pubsub": true,
			"ipnsps": true,
			"mplex":  true,
		},
	}

	ncfg.Routing = core.DHTOption

	node, err := core.NewNode(ctx, ncfg)
	if err != nil {
		panic(err)
	}

	node.SetLocal(false)

	cctx.ConstructNode = func() (*core.IpfsNode, error) {
		return node, nil
	}

ReTry:
	apiErrc, err := serveHTTPApi(cctx)
	if err != nil {
		err2 := os.Remove(filepath.Join(cctx.ConfigRoot, "repo.lock"))
		if err2 != nil {
			panic(err)
		}
		goto ReTry
	}

	var gwErrc <-chan error
	if len(cfg.Addresses.Gateway) > 0 {
		gwErrc, err = serveHTTPGateway(cctx)
		if err != nil {
			panic(err)
		}
	}

	prometheus.MustRegister(&corehttp.IpfsNodeCollector{Node: node})

	ch <- node

	select {
	case es := <-apiErrc:
		fmt.Println(es)
	case es := <-gwErrc:
		fmt.Println(es)
	case <-ctx.Done():
		return
	}
	fmt.Println("over")
	return
}

func buildEnv(ctx context.Context) (cmds.Environment, error) {
	repoPath, err := getRepoPath() // 获取配置文件位置
	if err != nil {
		return nil, err
	}
	log.Debugf("config path is %s", repoPath)

	// this sets up the function that will initialize the node
	// this is so that we can construct the node lazily.
	return &oldcmds.Context{
		ConfigRoot: repoPath,
		LoadConfig: loadConfig,
		ReqLog:     &oldcmds.ReqLog{},
		ConstructNode: func() (n *core.IpfsNode, err error) {
			r, err := fsrepo.Open(repoPath)
			if err != nil { // repo is owned by the node
				return nil, err
			}
			fmt.Println("test")
			// ok everything is good. set it on the invocation (for ownership)
			// and return it.
			n, err = core.NewNode(ctx, &core.BuildCfg{
				Repo: r,
			})
			if err != nil {
				return nil, err
			}

			n.SetLocal(true)
			return n, nil
		},
	}, nil
}
