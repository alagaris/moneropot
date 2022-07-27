package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"moneropot/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"moneropot/monerorpc"
)

var (
	pickWinnerTimer  *time.Timer
	checkTransferCtr int64
)

func checkTransfers() {
	checkMins := int64(120)
	if !util.Config.Production {
		log.Println("checkTransfers...", checkTransferCtr)
		checkMins = 2
	}
	if checkTransferCtr%checkMins == 0 {
		// check missed transfers every hour
		if !util.Config.Production {
			log.Println("checking missed transfers...")
		}
		if err := CheckMissedTransfers(); err != nil {
			log.Println("checkTransfers missed transfer error: ", err)
			return
		}
	}
	checkTransferCtr++

	h, err := LastHeight()
	if err != nil {
		log.Println("checkTransfers last height error: ", err)
		return
	}
	db, err := GetDB()
	if err != nil {
		log.Println("checkTransfers get db error: ", err)
		return
	}
	walletLock.Lock()
	resp, err := Wallet.GetTransfers(&monerorpc.GetTransfersRequest{
		In:             true,
		FilterByHeight: true,
		MinHeight:      h,
	})
	walletLock.Unlock()
	if err != nil {
		log.Println("checkTransfers: ", err)
		return
	}
	if resp.In == nil {
		return
	}
	if CurrentPrice == 0 {
		if err := SetCurrentPrice(); err != nil {
			log.Println("checkTransfers: failed to set current price", err)
			return
		}
	}
	// Create a map of rows to inbound transfers
	var indexes []string
	for _, val := range resp.In {
		indexes = append(indexes, strconv.FormatUint(val.SubaddrIndex.Minor, 10))
	}

	accounts := []Account{}
	err = db.Select(&accounts, fmt.Sprintf(`
		SELECT * FROM accounts WHERE address_index IN (%s) AND active = 1`, strings.Join(indexes, ",")))
	if err != nil {
		log.Println("checkTransfers select accounts error: ", err)
		return
	}
	log.Println("checkTransfers found ", len(accounts), " accounts")
	// Map subaddress_index to callbackDest{url, description}.
	m := make(map[uint64]*Account)
	for _, account := range accounts {
		m[account.AddressIndex] = &account
	}
	newAmounts := make(map[int64]uint64)
	tx, err := db.Begin()
	for _, t := range resp.In {
		if t.Height > h {
			h = t.Height
		}

		account, ok := m[t.SubaddrIndex.Minor]
		if !ok {
			log.Printf("No account associated with transfer: %d -> %d", t.SubaddrIndex.Minor, t.Amount)
			continue
		}

		if _, ok := newAmounts[account.ID]; !ok {
			newAmounts[account.ID] = account.Amount
		}
		newAmounts[account.ID] += t.Amount
		if _, err := tx.Exec(`INSERT INTO transactions (id) VALUES ($1)`, t.Txid); err != nil {
			log.Println("checkTransfers insert tx error", err, "Rollback:", tx.Rollback())
			return
		}
	}
	if err != nil {
		log.Println("checkTransfers begin tx error", err)
		return
	}
	if err := createNewEntries(tx, newAmounts, h); err != nil {
		log.Println("checkTransfers create new entries error", err, "Rollback:", tx.Rollback())
		return
	}

	for acctID, _ := range newAmounts {
		event := strconv.FormatInt(acctID, 10)
		util.PublishTopic(event, event)
	}
	util.Cache.Delete("info")
	util.PublishTopic("", "info")
	log.Println("Updated height to ", h)
}

func createNewEntries(tx *sql.Tx, newAmounts map[int64]uint64, newHeight uint64) error {
	md := make(map[string]string)
	rows, err := tx.Query(`SELECT key, value FROM metadata WHERE key IN ('entry_id', 'sign_key')`)
	if err != nil {
		return fmt.Errorf("createNewEntries select metadata error %v", err)
	}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("createNewEntries scan key value error %v", err)
		}
		md[key] = value
	}
	entryID, err := strconv.ParseInt(md["entry_id"], 10, 64)
	if err != nil {
		return fmt.Errorf("createNewEntries parse entry id error %v", err)
	}
	signKey := md["sign_key"]
	log.Println("Current entry id", entryID)
	for accountId, amount := range newAmounts {
		entries := 0
		for {
			if amount >= CurrentPrice {
				amount -= CurrentPrice
				entries++
				entryID++
				_, err := tx.Exec(`INSERT INTO entries (id, account_id, hash) VALUES ($1, $2, $3)`,
					entryID, accountId, util.SignEntry(entryID, signKey))
				if err != nil {
					return fmt.Errorf("createNewEntries insert entry error %v", err)
				}
			} else {
				break
			}
		}
		_, err := tx.Exec(`UPDATE accounts SET amount = $1, entries = entries + $2 WHERE id = $3`, amount, entries, accountId)
		if err != nil {
			return fmt.Errorf("createNewEntries update amount error %v", err)
		}
	}
	_, err = tx.Exec(`
		UPDATE metadata SET value = $1 WHERE key = 'last_height';
		UPDATE metadata SET value = $2 WHERE key = 'entry_id';`,
		strconv.FormatUint(newHeight, 10), strconv.FormatInt(entryID, 10))
	if err != nil {
		return fmt.Errorf("createNewEntries update metadata error %v", err)
	}
	return tx.Commit()
}

func entriesFromAmount(amount uint64) (int64, uint64) {
	var entries int64
	for {
		if amount >= CurrentPrice {
			amount -= CurrentPrice
			entries++
		} else {
			break
		}
	}
	return entries, amount
}

func RunBackground() {
	// Check wallet first
	if err := syncWallet(); err != nil {
		panic(err)
	}

	// do price updates every 3AM
	time.AfterFunc(AtHourMinute(3, 0), priceUpdate)
	// pick winner every first of the month
	pickWinnerTimer = time.AfterFunc(StartOfMonth(), runPickWinner)

	log.Println("Started background task")
	for {
		checkTransfers()
		if util.Config.Production {
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

func CheckMissedTransfers() error {
	tx, err := MustDB().Begin()
	if err != nil {
		return fmt.Errorf("CheckMissedTransfers: error tx %v", err)
	}
	var (
		h      uint64
		maxH   uint64
		newVar bool
	)

	row := tx.QueryRow(`SELECT value FROM metadata WHERE key = 'missed_height_check'`)
	if err := row.Scan(&h); err != nil {
		if err == sql.ErrNoRows {
			newVar = true
		} else {
			return fmt.Errorf("CheckMissedTransfers: error missed_height_check %v", err)
		}
	}
	row = tx.QueryRow(`SELECT value FROM metadata WHERE key = 'last_height'`)
	if err := row.Scan(&maxH); err != nil {
		return fmt.Errorf("CheckMissedTransfers: error last_height %v", err)
	}

	walletLock.Lock()
	resp, err := Wallet.GetTransfers(&monerorpc.GetTransfersRequest{
		In:             true,
		FilterByHeight: true,
		MinHeight:      h,
		MaxHeight:      maxH,
	})
	walletLock.Unlock()
	if err != nil {
		return fmt.Errorf("CheckMissedTransfers: error %v", err)
	}

	var indexes []string
	dupIndex := make(map[uint64]bool)
	for _, val := range resp.In {
		if _, ok := dupIndex[val.SubaddrIndex.Minor]; !ok {
			indexes = append(indexes, strconv.FormatUint(val.SubaddrIndex.Minor, 10))
			dupIndex[val.SubaddrIndex.Minor] = true
		}
	}

	var accounts []Account
	if err != nil {
		return fmt.Errorf("CheckMissedTransfers: error begin transaction %v", err)
	}
	rows, err := tx.Query(fmt.Sprintf(`SELECT id, address_index, address, user_name, user_address, amount, entries, active, ref_id
		FROM accounts WHERE active = 1 AND address_index IN (%s)`, strings.Join(indexes, ",")))
	if err != nil {
		return fmt.Errorf("CheckMissedTransfers: query error %v", err)
	}
	accountTotal := make(map[uint64]uint64)
	for rows.Next() {
		account := Account{}
		if err := rows.Scan(&account.ID, &account.AddressIndex, &account.Address, &account.UserName, &account.UserAddress, &account.Amount, &account.Entries, &account.Active, &account.RefID); err != nil {
			return fmt.Errorf("CheckMissedTransfers: scan error %v", err)
		}
		accounts = append(accounts, account)
		if !newVar && account.Amount > 0 {
			// not first run so tally all, after will be tallied from current db amount
			accountTotal[account.AddressIndex] = account.Amount
		}
	}
	newAmounts := make(map[int64]uint64)
	for _, t := range resp.In {
		if t.Height > h {
			h = t.Height
		}

		var txID string
		row := tx.QueryRow(`SELECT id FROM transactions WHERE id = $1`, t.Txid)
		err := row.Scan(&txID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("CheckMissedTransfers: error select tx %v", err)
		}
		if err == nil && txID != "" {
			continue // already processed
		}

		total, ok := accountTotal[t.SubaddrIndex.Minor]
		if !ok {
			accountTotal[t.SubaddrIndex.Minor] = 0
		}
		accountTotal[t.SubaddrIndex.Minor] = total + t.Amount

		if _, err := tx.Exec(`INSERT INTO transactions (id) VALUES ($1)`, t.Txid); err != nil {
			return fmt.Errorf("CheckMissedTransfers: insert tx error %v", err)
		}
	}
	var missingEntries int64
	for _, account := range accounts {
		total, ok := accountTotal[account.AddressIndex]
		if !ok {
			continue
		}
		// initial fix would check against total entries but continous fix would only check against current balance
		if newVar {
			entries, amountLeft := entriesFromAmount(total)
			if account.Entries < entries {
				missingEntries += entries - account.Entries
				newAmounts[account.ID] = uint64((entries-account.Entries)*int64(CurrentPrice)) + amountLeft
			}
		} else if total > 0 && account.Amount < total {
			// unaccounted totals that got missed
			newAmounts[account.ID] = total
		}
	}
	sql := `UPDATE metadata SET value = $1 WHERE key = 'missed_height_check'`
	if newVar {
		sql = `INSERT INTO metadata (key, value) VALUES ('missed_height_check', $1)`
	}
	if _, err := tx.Exec(sql, h); err != nil {
		return fmt.Errorf("CheckMissedTransfers: error metadata set %v", err)
	}
	if len(newAmounts) > 0 {
		if err := createNewEntries(tx, newAmounts, h); err != nil {
			return fmt.Errorf("CheckMissedTransfers: create entries error %v", err)
		}
		log.Println("Created missing entries: ", missingEntries)
	} else {
		return tx.Commit()
	}
	return nil
}

func priceUpdate() {
	log.Printf("Updating price...")
	// make sure we get all missed transaction from last check (until we figure out how we are missing transactions)
	if err := CheckMissedTransfers(); err != nil {
		log.Println("Error check missed transfer", err)
		time.AfterFunc(time.Minute*1, priceUpdate)
		return
	}
	err := SetCurrentPrice()
	if err != nil {
		log.Println("Error updating price", err)
		time.AfterFunc(time.Minute*1, priceUpdate)
		return
	}
	time.AfterFunc(AtHourMinute(0, 40), priceUpdate)
}

func AtHourMinute(hour int, minute int) time.Duration {
	t := util.UtcNow()
	n := time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
	if t.After(n) {
		n = n.Add(24 * time.Hour)
	}
	d := n.Sub(t)
	return d
}

func StartOfMonth() time.Duration {
	now := util.UtcNow()
	year, month, _ := now.Date()
	addMonth := time.Month(1)
	return time.Date(year, month+addMonth, 1, 0, 5, 0, 0, now.Location()).Sub(now)
}

// database backups every 11:30PM
func doBackup() {
	defer time.AfterFunc(AtHourMinute(23, 30), doBackup)
	if !util.FileExists(dbPath) {
		return
	}
	backupDir := filepath.Join(util.Config.DataPath, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Failed to create backup dir: %v", err)
		return
	}
	backupPath := filepath.Join(backupDir, util.UtcNow().Format(SDateTimeFormat)+".db")
	if !util.FileExists(backupPath) {
		dbLock.Lock()
		if dbx != nil {
			dbx.Close()
			dbx = nil
		}
		log.Printf("Backing up db %s to %s", dbPath, backupPath)
		if err := util.CopyFile(dbPath, backupPath); err != nil {
			log.Printf("Failed to copy backup db: %v", err)
		}
		dbLock.Unlock()

		bkFiles, err := ioutil.ReadDir(backupDir)
		if err != nil {
			log.Println("Failed to read dir", err)
			return
		}

		for _, file := range bkFiles {
			if !file.IsDir() {
				if time.Since(file.ModTime()) > time.Hour*24*60 {
					p := filepath.Join(backupDir, file.Name())
					if err := os.Remove(p); err != nil {
						log.Println("Failed delete file", file.Name(), err)
					}
				}
			}
		}
	}
}
