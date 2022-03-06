package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"moneropot/monerorpc"
	"moneropot/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	WinnerInfo struct {
		SignKey  string           `json:"sign_key"`
		Block    string           `json:"block"`
		Entries  int64            `json:"entries"`
		Amount   int64            `json:"amount"`
		Accounts map[string][]int `json:"accounts"`
	}

	WinAccount struct {
		ID          int64   `db:"id"`
		UserName    *string `db:"user_name"`
		UserAddress string  `db:"user_address"`
		Wins        int64   `db:"wins"`
		EntryID     int64   `db:"entry_id"`
	}
)

var (
	errAlreadyProcessed = fmt.Errorf("already processed")
)

func runPickWinner() {
	if err := pickWinner(); err != nil {
		log.Println("runPickWinner error ", err)
		time.AfterFunc(time.Minute*1, runPickWinner)
		return
	}
	pickWinnerTimer = time.AfterFunc(StartOfMonth(), runPickWinner)
}

func FlushWinPayload(month string) error {
	db := MustDB()
	_, err := db.Exec(`UPDATE winners SET transfer_body = NULL WHERE date = $1`, month)
	return err
}

func pickWinner() error {
	// make sure there's no missed transfers from last check before running the pick winner
	if err := CheckMissedTransfers(); err != nil {
		return fmt.Errorf("pickWinner missed transfer error %v", err)
	}
	now := util.UtcNow()
	year, month, _ := now.Date()
	prevMonth := time.Date(year, month-1, 1, 0, 0, 0, 0, now.Location())
	winMonth := prevMonth.Format("2006-01")
	log.Println("Picking winner for", winMonth)
	db := MustDB()
	dbLock.Lock()
	defer dbLock.Unlock()

	// first make sure this month hasn't already been processed and if it's already been distributed
	checkAndTransfer := func() error {
		w := &Winner{}
		err := db.Get(w, `SELECT * FROM winners WHERE date = $1`, winMonth)
		if err != nil {
			if util.NoRows(err) {
				return nil
			}
			return fmt.Errorf("checkAndTransfer select error %v", err)
		}
		if w.TransferBody != nil {
			// store transfer request to filesystem and remove in db
			reqPath := filepath.Join(util.Config.DataPath, "transfers")
			if err := os.MkdirAll(reqPath, 0755); err != nil {
				return fmt.Errorf("checkAndTransfers mkdir error %v", err)
			}
			reqPath = filepath.Join(reqPath, winMonth+".json")
			b := []byte(*w.TransferBody)
			if err := ioutil.WriteFile(reqPath, b, 0644); err != nil {
				return fmt.Errorf("checkAndTransfers writefile error %v", err)
			}
			tsr := &monerorpc.TransferSplitRequest{}
			if err := json.Unmarshal(b, tsr); err != nil {
				return fmt.Errorf("checkAndTransfers unmarshal error %v", err)
			}
			// make sure we don't have a locked amount
			_, err := GetDistributedAmounts(true)
			if err != nil {
				return fmt.Errorf("checkAndTransfers distribute amount has error %v", err)
			}
			walletLock.Lock()
			_, err = Wallet.TransferSplit(tsr)
			walletLock.Unlock()
			var randomOutsErr bool
			if err != nil {
				randomOutsErr = strings.Contains(err.Error(), "failed to get random outs")
				if !randomOutsErr {
					util.SendEvent("checkAndTransfers transfer error " + err.Error() + "\nPayload: \n" + *w.TransferBody)
					return fmt.Errorf("checkAndTransfers transfer error %v", err)
				}
			}
			if _, err := db.Exec(`UPDATE winners SET transfer_body = NULL WHERE date = $1`, winMonth); err != nil {
				log.Println("checkAndTransfers failed to null transfer_body", err)
				util.SendEvent("checkAndTransfers failed to null transfer_body: " + err.Error())
			} else if randomOutsErr {
				// try to send this 1 at a time, and not retry anymore
				// todo if it still fails we can do a sweep to itself?
				var failedTransfers []string
				walletLock.Lock()
				for _, v := range tsr.Destinations {
					_, err = Wallet.Transfer(&monerorpc.TransferRequest{
						Destinations: []monerorpc.Destination{
							{Amount: v.Amount, Address: v.Address},
						},
					})
					if err != nil {
						failedTransfers = append(failedTransfers,
							fmt.Sprintf("Address: %s \nAmount: %s \nXMR: %d \nError %s",
								v.Address, monerorpc.XMRToDecimal(v.Amount), v.Amount, err.Error()))
					}
				}
				walletLock.Unlock()
				if len(failedTransfers) > 0 {
					util.SendEvent("checkAndTransfers transfer failed --\n" + strings.Join(failedTransfers, "\n-----\n"))
				}
			}
		} else {
			return errAlreadyProcessed
		}
		return nil
	}
	if err := checkAndTransfer(); err != nil {
		if err == errAlreadyProcessed {
			log.Println("pick winner ran already processed", winMonth)
			util.SendEvent("pick winner ran already processed " + winMonth)
			return nil
		}
		return fmt.Errorf("pickWinner error %v", err)
	}

	firstBlock, err := GetFirstBlockOfMonth(util.UtcNow())
	if err != nil {
		return fmt.Errorf("pickWinner first block error %v", err)
	}

	amt, err := GetDistributedAmounts(false)
	if err != nil {
		return fmt.Errorf("pickWinner get distrubuted amount error %v", err)
	}

	type refAmounts struct {
		AccountID int64 `db:"ref_id"`
		Total     int64 `db:"total"`
	}
	ra := []refAmounts{}
	err = db.Select(&ra, `SELECT ref_id, SUM(entries) as total
	FROM accounts
	WHERE active = 1 AND entries > 0
	GROUP BY ref_id`)
	if err != nil {
		return fmt.Errorf("pickWinner select group error %v", err)
	}
	accounts := []Account{}
	err = db.Select(&accounts, `SELECT * FROM accounts WHERE active = 1 AND entries > 0 AND user_address IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("pickWinner select error %v", err)
	}

	type mData struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}
	var (
		md    []mData
		mdMap = make(map[string]string)
	)
	err = db.Select(&md, `SELECT * FROM metadata WHERE key IN ('entry_id','sign_key')`)
	if err != nil {
		return fmt.Errorf("pickWinner metadata select error %v", err)
	}
	for _, v := range md {
		mdMap[v.Key] = v.Value
	}
	entries := mdMap["entry_id"]
	signKey := mdMap["sign_key"]

	if entries == "0" {
		log.Println("pickWinner skipped, no entries for month", winMonth)
		util.SendEvent("pickWinner skipped, no entries for month " + winMonth)
		return nil
	}
	totalEntries, _ := strconv.Atoi(entries)
	var (
		winners []string
		highest int
	)
	log.Println("Processing", totalEntries, "entries")
	for i := 1; i <= totalEntries; i++ {
		h := util.HashMatchAlign(firstBlock, util.SignEntry(int64(i), signKey))
		if h > highest {
			highest = h
			winners = make([]string, 0)
		}
		if h >= highest {
			winners = append(winners, strconv.Itoa(i))
		}
	}
	var winAccounts []WinAccount
	if err := db.Select(&winAccounts, fmt.Sprintf(`
	SELECT a.id, a.user_address, a.user_name, COUNT(e.id) as wins
	FROM entries AS e
	LEFT JOIN accounts as a ON a.id = e.account_id
	WHERE e.id IN (%s)
	GROUP BY a.id`, strings.Join(winners, ","))); err != nil {
		return fmt.Errorf("pickWinner win accounts error %v", err)
	}
	totalWinners := float64(len(winners))
	var accountEntries []WinAccount
	if err := db.Select(&accountEntries, fmt.Sprintf(`
	SELECT a.user_address, a.user_name, e.id as entry_id
	FROM entries AS e
	LEFT JOIN accounts as a ON a.id = e.account_id
	WHERE e.id IN (%s)`, strings.Join(winners, ","))); err != nil {
		return fmt.Errorf("pickWinner win account map error %v", err)
	}
	winMap := make(map[string][]int)
	for _, entry := range accountEntries {
		key := entry.UserAddress
		if _, ok := winMap[key]; !ok {
			winMap[key] = make([]int, 0)
		}
		winMap[key] = append(winMap[key], int(entry.EntryID))
	}

	// calculate refs for distribution
	tr := &monerorpc.TransferSplitRequest{
		Priority:   0,
		Mixin:      8,
		UnlockTime: 10,
	}
	tr.Destinations = append(tr.Destinations, monerorpc.Destination{
		Amount:  amt.Maintenance,
		Address: util.Config.MaintAddress,
	})
	winAmount := float64(amt.Winner)
	destinations := make(map[string]uint64)
	for _, winAccount := range winAccounts {
		val, ok := destinations[winAccount.UserAddress]
		if !ok {
			destinations[winAccount.UserAddress] = 0
		}
		destinations[winAccount.UserAddress] = val + uint64(winAmount*(float64(winAccount.Wins)/totalWinners))
	}
	refAmt := float64(amt.Referrals)
	var refIDs []string
	refMap := make(map[int64]uint64)
	refCredit := make(map[int64]uint64)
	for _, val := range ra {
		refIDs = append(refIDs, strconv.FormatInt(val.AccountID, 10))
		award := uint64(refAmt * (float64(val.Total) / float64(totalEntries)))
		if award < CurrentPrice {
			refCredit[val.AccountID] = award
		} else {
			refMap[val.AccountID] = award
		}
	}
	refAccounts := []Account{}
	refs := strings.Join(refIDs, ",")
	err = db.Select(&refAccounts, fmt.Sprintf(`SELECT * FROM accounts WHERE id IN(%s) AND user_address IS NOT NULL`, refs))
	if err != nil {
		return fmt.Errorf("pickWinner select refs acounts error %v", err)
	}

	for _, account := range refAccounts {
		if amount, ok := refMap[account.ID]; ok {
			val, ok := destinations[*account.UserAddress]
			if !ok {
				destinations[*account.UserAddress] = 0
			}
			destinations[*account.UserAddress] = val + amount
		}
	}

	for addr, amount := range destinations {
		tr.Destinations = append(tr.Destinations, monerorpc.Destination{
			Amount:  amount,
			Address: addr,
		})
	}

	b, err := json.Marshal(tr)
	if err != nil {
		return fmt.Errorf("pickWinner marshall error %v", err)
	}
	// transaction here, must complete or fail all and restart the process
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("pickWinner tx error %v", err)
	}

	// insert the winner
	winInfo := WinnerInfo{
		SignKey:  signKey,
		Block:    firstBlock,
		Entries:  int64(totalEntries),
		Amount:   int64(amt.Winner),
		Accounts: winMap,
	}
	bw, err := json.Marshal(winInfo)
	if err != nil {
		return fmt.Errorf("pickWinner marshal info error %v", err)
	}
	_, err = tx.Exec(`INSERT INTO winners (date, info, transfer_body)
		VALUES ($1, $2, $3)`,
		winMonth,
		string(bw),
		string(b),
	)

	if err != nil {
		return fmt.Errorf("pickWinner tx insert error %v -> Rollback: %v", err, tx.Rollback())
	}

	// just credit referrers if they made less then entry amount
	for acctID, amount := range refCredit {
		_, err = tx.Exec(`UPDATE accounts SET amount = $1 WHERE id = $2`, amount, acctID)
		if err != nil {
			return fmt.Errorf("pickWinner tx ref credit error %v -> Rollback: %v", err, tx.Rollback())
		}
	}

	// reset accounts and leave active ones
	_, err = tx.Exec(fmt.Sprintf(`UPDATE accounts SET
		active = 0,
		user_name = NULL,
		user_address = NULL,
		amount = 0,
		entries = 0,
		ref_id = 0
		WHERE user_name IS NULL OR
			(entries = 0 AND id NOT IN (%s));
		UPDATE accounts SET entries = 0 WHERE entries > 0;
		UPDATE metadata SET value = '0' WHERE key = 'entry_id';
		UPDATE metadata SET value = '%s' WHERE key = 'sign_key';
		DELETE FROM entries;`, refs, firstBlock))
	if err != nil {
		return fmt.Errorf("pickWinner tx update error %v -> Rollback: %v", err, tx.Rollback())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pickWinner tx commit error %v -> Rollback: %v", err, tx.Rollback())
	}

	if err := checkAndTransfer(); err != nil {
		return fmt.Errorf("pickWinner checkAndTransfer error %v", err)
	}

	log.Println("pickWinner completed")
	return nil
}

func RunPickWinnerManually() {
	log.Println("Running pick winner manually")
	if pickWinnerTimer != nil {
		pickWinnerTimer.Stop()
		pickWinnerTimer = nil
	}
	runPickWinner()
}
