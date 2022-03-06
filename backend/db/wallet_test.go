package db

import (
	"moneropot/monerorpc"
	"moneropot/util"
	"os"
	"testing"

	"github.com/gabstv/httpdigest"
)

func TestFirstBlockOfMonth(t *testing.T) {
	cfg := monerorpc.Config{
		Address:   "http://localhost:28081/json_rpc",
		Transport: httpdigest.New(util.Config.DaemonUser, util.Config.DaemonPass),
	}
	Daemon = monerorpc.New(cfg)

	// block, err := GetFirstBlockOfMonth(time.Date(2021, 12, 1, 0, 5, 0, 0, time.UTC))
	// t.Errorf("Block %s -> %v", block, err)
}

func TestGetTransfers(t *testing.T) {
	os.Setenv("DB_NAME", ":memory:")
	util.ParseArgs()
	Init()
	cfg := monerorpc.Config{
		Address:   "http://localhost:28081/json_rpc",
		Transport: httpdigest.New(util.Config.RpcUser, util.Config.RpcPass),
	}
	Wallet = monerorpc.New(cfg)
	// if err := CheckMissedTransfers(); err != nil {
	// 	t.Fatalf("Err checkmissedTransfers %v", err)
	// }
	// t.Errorf("TEST")
}
