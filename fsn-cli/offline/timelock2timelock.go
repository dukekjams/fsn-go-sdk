// Copyright 2019 The fsn-go-sdk Authors
// This file is part of the fsn-go-sdk library.
//
// The fsn-go-sdk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The fsn-go-sdk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the fsn-go-sdk library. If not, see <http://www.gnu.org/licenses/>.

package offline

import (
	"github.com/FusionFoundation/fsn-go-sdk/efsn/cmd/utils"
	"github.com/FusionFoundation/fsn-go-sdk/efsn/common"
	clicommon "github.com/FusionFoundation/fsn-go-sdk/fsn-cli/common"
	"github.com/FusionFoundation/fsn-go-sdk/fsnapi"
	"gopkg.in/urfave/cli.v1"
)

var CommandTimeLock2Timelock = cli.Command{
	Name:      "timelock2timelock",
	Aliases:   []string{"t2t"},
	Usage:     "(offline) build timelock to timelock raw transaction",
	ArgsUsage: "<assetID> <to> <value>",
	Description: `
build timelock to timelock raw transaction`,
	Flags: append([]cli.Flag{
		timeLockStartFlag,
		timeLockEndFlag,
	}, commonFlags...),
	Action: timelock2timelock,
}

func timelock2timelock(ctx *cli.Context) error {
	setLogger(ctx)
	if len(ctx.Args()) != 3 {
		cli.ShowCommandHelpAndExit(ctx, "timelock2timelock", 1)
	}

	assetID_ := ctx.Args().Get(0)
	to_ := ctx.Args().Get(1)
	value_ := ctx.Args().Get(2)

	assetID := clicommon.GetHashFromText("assetID", assetID_)
	to := clicommon.GetAddressFromText("to", to_)
	value := clicommon.GetHexBigIntFromText("asset", value_)

	start := getHexUint64Time(ctx, timeLockStartFlag.Name)
	end := getHexUint64Time(ctx, timeLockEndFlag.Name)

	// 1. construct corresponding arguments and options
	baseArgs, signOptions := getBaseArgsAndSignOptions(ctx)
	args := &common.TimeLockArgs{
		SendAssetArgs: common.SendAssetArgs{
			FusionBaseArgs: baseArgs,
			AssetID:        assetID,
			To:             to,
			Value:          value,
		},
		StartTime: start,
		EndTime:   end,
	}

	// 2. check parameters
	now := getNowTime()
	args.Init(common.TimeLockToTimeLock)
	if err := args.ToParam().Check(common.BigMaxUint64, now); err != nil {
		utils.Fatalf("check parameter failed: %v", err)
	}

	// 3. build and/or sign transaction through fsnapi
	tx, err := fsnapi.BuildFSNTx(common.TimeLockFunc, args, signOptions)
	if err != nil {
		utils.Fatalf("create tx error: %v", err)
	}

	return printTx(tx, false)
}
