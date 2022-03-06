package db

import (
	"fmt"
	"log"
	"sync"
	"time"

	"moneropot/monerorpc"
	"moneropot/util"
)

var (
	Wallet     *monerorpc.Client
	Daemon     *monerorpc.Client
	walletLock sync.Mutex
	daemonLock sync.Mutex
)

func IsValidAddress(address string) error {
	walletLock.Lock()
	r, err := Wallet.ValidateAddress(&monerorpc.ValidateAddressRequest{Address: address})
	walletLock.Unlock()
	if err != nil {
		return err
	}
	if r.Valid {
		return nil
	}
	return fmt.Errorf("invalid address")
}

func GetWalletAddress() (string, error) {
	walletLock.Lock()
	r, err := Wallet.GetAddress(&monerorpc.GetAddressRequest{})
	walletLock.Unlock()
	if err != nil {
		return "", err
	}
	return r.Address, nil
}

func GetFirstBlockOfMonth(tm time.Time) (string, error) {
	daemonLock.Lock()
	defer daemonLock.Unlock()

	bh, err := Daemon.GetLastBlockHeader()
	if err != nil {
		return "", fmt.Errorf("first block error %v", err)
	}
	latestBlockTime := time.Unix(int64(bh.BlockHeader.Timestamp), 0)
	month := time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, time.UTC)
	beforeMonth := month.Add(-1 * time.Microsecond)
	if latestBlockTime.Before(month) {
		return "", fmt.Errorf("first block for month not yet created")
	}
	foundBeforeMonth := false
	var foundBlock string
	blockDiff := uint64(latestBlockTime.Sub(month).Minutes() / 2)
	var sb int64
	if bh.BlockHeader.Height > blockDiff {
		sb = int64(bh.BlockHeader.Height - blockDiff)
	}
	st := int64(0)
	et := int64(70)
	adjustCount := 0
	for {
		var s, e uint64
		if sb+st > 0 {
			s = uint64(sb + st)
		}
		if sb+et > 0 {
			e = uint64(sb + et)
		}
		if s == bh.BlockHeader.Height {
			s -= 20
		}
		if e > bh.BlockHeader.Height {
			e = bh.BlockHeader.Height
		}
		br, err := Daemon.GetBlockHeadersRange(&monerorpc.GetBlockHeadersRangeRequest{
			StartHeight: s,
			EndHeight:   e,
		})
		if err != nil {
			return "", fmt.Errorf("first block error range %v", err)
		}
		for _, h := range br.BlockHeaders {
			t := time.Unix(int64(h.Timestamp), 0)
			if t.Before(month) {
				foundBeforeMonth = true
			} else if foundBeforeMonth && t.After(beforeMonth) {
				if !util.Config.Production {
					log.Println("Block", h)
				}
				foundBlock = h.Hash
				break
			}
		}
		if foundBlock != "" {
			break
		}
		if foundBeforeMonth {
			st += 50
			et += 50
		} else {
			st -= 50
		}
		adjustCount++
		if adjustCount > 3 {
			return "", fmt.Errorf("first block of month not found")
		}
	}
	return foundBlock, nil
}
