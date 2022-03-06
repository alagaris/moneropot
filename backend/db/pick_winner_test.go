package db

import (
	"encoding/json"
	"fmt"
	"log"
	"moneropot/monerorpc"
	"moneropot/util"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestPickWinner(t *testing.T) {
	util.Now = func() time.Time {
		return time.Date(2021, 11, 20, 0, 0, 0, 0, time.UTC)
	}
	os.Setenv("DB_NAME", ":memory:")
	util.ParseArgs()
	Init()
	newAmounts := make(map[int64]uint64)
	addrIndex := 0
	monerorpc.SetFakeResponse("create_address", func(in interface{}) string {
		addrIndex++
		addr := util.RandomString(95)
		return fmt.Sprintf(`{"address":"%s","address_index":%d}`, addr, addrIndex)
	})
	monerorpc.SetFakeResponse("get_transfers", func(i interface{}) string {
		return `{"in":[]}`
	})
	monerorpc.SetFakeResponse("get_last_block_header", func(i interface{}) string {
		return `{"block_header":{"hash":"b417bda53fb674146f18777c0d42bbc3bb5e110ee22acec108e7b40e6addc767","height":2496780,"timestamp":1637336695}}`
	})
	// firstBlock := "771fbcd656ec1464d3a02ead5e18644030007a0fc664c0a964d30922821a8148"
	firstBlock := "6666666666ec1464d3a02ead5e18644030007a0fc664c0a964d30408821a8bb0"
	monerorpc.SetFakeResponse("get_block_headers_range", func(i interface{}) string {
		return fmt.Sprintf(`{"headers":[
			{"hash":"771fbcd656ec1464d3a02ead5e18644030007a0fc664c0a964d30922821a8148","height":1,"timestamp":1397818193},
			{"hash":"%s","height":2496780,"timestamp":1637336695}
		]}`, firstBlock)
	})
	monerorpc.SetFakeResponse("get_balance", func(i interface{}) string {
		return `{
			"balance": 5000000000000,
			"unlocked_balance": 5000000000000
		}`
	})
	transferred := make(map[string]uint64)
	monerorpc.SetFakeResponse("transfer_split", func(i interface{}) string {
		tsr, ok := i.(**monerorpc.TransferSplitRequest)
		if !ok {
			t.Fatalf("Invalid type: %T", i)
		}
		ts := *tsr
		var total uint64
		for _, dest := range ts.Destinations {
			v, ok := transferred[dest.Address]
			if !ok {
				transferred[dest.Address] = 0
			}
			transferred[dest.Address] = v + dest.Amount
			total += dest.Amount
		}
		if total != 3885714285714 {
			t.Errorf("Wanted total transfer 3885714285714 got %d", total)
		}
		return `{}`
	})

	uname := "ABC"
	// create 10 accounts and 20 entries
	var (
		acct *Account
		err  error
	)
	for i := 0; i < 9; i++ {
		if i == 0 {
			acct, err = GetAccount(util.RandomString(94)+strconv.Itoa(i), &uname, nil)
		} else {
			acct, err = GetAccount(util.RandomString(94)+strconv.Itoa(i), nil, &uname)
		}
		if err != nil {
			t.Errorf("get account error %v", err)
		}
		newAmounts[acct.ID] = CurrentPrice
		if i%2 == 0 {
			newAmounts[acct.ID] += CurrentPrice
		}
		if acct.ID == 5 {
			newAmounts[acct.ID] += 50
		}
	}
	tx, err := MustDB().Begin()
	if err != nil {
		t.Errorf("test pick winner tx error %v", err)
	}
	if err := createNewEntries(tx, newAmounts, 50); err != nil {
		t.Errorf("test pick winner new entries error %v", err)
	}
	entries, err := TotalEntries()
	if err != nil {
		t.Errorf("total entries error %v", err)
	}
	if entries != 14 {
		t.Errorf("Wanted 14 entries got %d", entries)
	}
	if err := dbx.Get(acct, `SELECT * FROM accounts WHERE id = 5`); err != nil {
		t.Errorf("select account 5 error %v", err)
	}
	if acct.Amount != 50 {
		t.Errorf("Wanted 50 account 5 amount got %d", acct.Amount)
	}
	var allEntry []Entry
	if err := dbx.Select(&allEntry, `SELECT * FROM entries`); err != nil {
		t.Errorf("pick winner error select entries %v", err)
	}
	tableMap := map[int64]int{
		1: 7, 2: 5, 3: 3, 4: 2, 5: 8, 6: 5, 7: 1, 8: 7, 9: 3, 10: 3, 11: 4, 12: 2, 13: 4, 14: 5,
	}
	for _, entry := range allEntry {
		if v, ok := tableMap[entry.ID]; ok {
			m := util.HashMatchAlign(firstBlock, entry.Hash)
			if v != m {
				t.Errorf("Wanted match %d: %d got %d", entry.ID, v, m)
			}
		}
		log.Println("AllEntries", entry.ID, entry.AccountID, entry.Hash, util.HashMatchAlign(firstBlock, entry.Hash))
	}
	signKey, _ := GetMetadata("sign_key", "----------------")
	if err := pickWinner(); err != nil {
		t.Errorf("pick winner error %v", err)
	}
	winner := Winner{}
	if err := dbx.Get(&winner, `SELECT * FROM winners`); err != nil {
		t.Errorf("pick winner select winner error %v", err)
	}
	info := WinnerInfo{}
	json.Unmarshal([]byte(winner.Info), &info)
	if winner.Date != "2021-10" {
		t.Errorf("Wanted winner date 2021-10 got %s", winner.Date)
	}
	if info.Amount != 2800000000000 {
		t.Errorf("Wanted win amount 2800000000000 got %d", info.Amount)
	}
	if info.Block != firstBlock {
		t.Errorf("Wanted info.Block %s got %s", firstBlock, info.Block)
	}
	if info.SignKey != signKey {
		t.Errorf("Wanted info.SignKey %s got %s", signKey, info.SignKey)
	}
	winners := make(map[int]string)
	for addr, entries := range info.Accounts {
		for _, entry := range entries {
			winners[entry] = addr
		}
	}
	if len(winners) != 1 {
		t.Errorf("Wanted 1 winners got %d", len(winners))
	}
	_, win5 := winners[5]
	if !win5 {
		t.Errorf("Wanted winners 5 got %v", winners)
	}
	if transferred[util.Config.MaintAddress] != 400000000000 {
		t.Errorf("Wanted maintenance amount 40000000000 got %d", transferred[util.Config.MaintAddress])
	}
	entryID, _ := GetMetadata("entry_id", "1000")
	if entryID != "0" {
		t.Errorf("Wanted entry_id of '0' got '%s'", entryID)
	}
	signKey, _ = GetMetadata("sign_key", "")
	if signKey != firstBlock {
		t.Errorf("Wanted sign_key of '%s' got '%s'", firstBlock, signKey)
	}
	if entries != info.Entries {
		t.Errorf("Wanted entries %d got %d", entries, info.Entries)
	}
	var count int64
	err = dbx.Get(&count, `SELECT COUNT(*) FROM entries`)
	if err != nil {
		t.Errorf("select count entries error %v", err)
	}
	if count > 0 {
		t.Errorf("Wanted entry count 0 got %d", count)
	}
	err = dbx.Get(&count, `SELECT COUNT(*) FROM accounts WHERE active = 1`)
	if err != nil {
		t.Errorf("select active count error %v", err)
	}
	if count != 1 {
		t.Errorf("Wanted active count 1 got %d", count)
	}
	// todo maybe do more tests here

}
