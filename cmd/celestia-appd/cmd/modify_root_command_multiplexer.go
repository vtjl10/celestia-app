//go:build multiplexer

package cmd

import (
	"github.com/celestiaorg/celestia-app/v5/app"
	embedding "github.com/celestiaorg/celestia-app/v5/internal/embedding"
	"github.com/celestiaorg/celestia-app/v5/multiplexer/abci"
	"github.com/celestiaorg/celestia-app/v5/multiplexer/appd"
	multiplexer "github.com/celestiaorg/celestia-app/v5/multiplexer/cmd"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
)

// v2UpgradeHeight is the block height at which the v2 upgrade occurred.
// this can be overridden at build time using ldflags:
// -ldflags="-X 'github.com/celestiaorg/celestia-app/v5/cmd/celestia-appd/cmd.v2UpgradeHeight=1751707'" for arabica
// -ldflags="-X 'github.com/celestiaorg/celestia-app/v5/cmd/celestia-appd/cmd.v2UpgradeHeight=2585031'" for mocha
// -ldflags="-X 'github.com/celestiaorg/celestia-app/v5/cmd/celestia-appd/cmd.v2UpgradeHeight=2371495'" for mainnet
var v2UpgradeHeight = ""

// modifyRootCommand enhances the root command with the pass through and multiplexer.
func modifyRootCommand(rootCommand *cobra.Command) {
	version, compressedBinary, err := embedding.CelestiaAppV3()
	if err != nil {
		panic(err)
	}

	appdV3, err := appd.New(version, compressedBinary)
	if err != nil {
		panic(err)
	}

	var extraArgs []string
	if v2UpgradeHeight != "" {
		extraArgs = append(extraArgs, "--v2-upgrade-height="+v2UpgradeHeight)
	}

	versions, err := abci.NewVersions(abci.Version{
		Appd:        appdV3,
		ABCIVersion: abci.ABCIClientVersion1,
		AppVersion:  3,
		StartArgs: append([]string{
			"--grpc.enable",
			"--api.enable",
			"--api.swagger=false",
			"--with-tendermint=false",
			"--transport=grpc",
		}, extraArgs...),
	})
	if err != nil {
		panic(err)
	}

	rootCommand.AddCommand(
		multiplexer.NewPassthroughCmd(versions),
	)

	// Add the following commands to the rootCommand: start, tendermint, export, version, and rollback and wire multiplexer.
	server.AddCommandsWithStartCmdOptions(
		rootCommand,
		app.NodeHome,
		NewAppServer,
		appExporter,
		server.StartCmdOptions{
			AddFlags:            addStartFlags,
			StartCommandHandler: multiplexer.New(versions),
		},
	)
}
