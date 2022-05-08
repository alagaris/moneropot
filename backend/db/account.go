package db

import (
	"encoding/json"
	"fmt"
	"moneropot/util"
	"strconv"

	"moneropot/monerorpc"
)

var (
	ErrDuplicateUser = fmt.Errorf("duplicate user")
)

type (
	Account struct {
		ID           int64   `db:"id"`
		AddressIndex uint64  `db:"address_index"`
		Address      string  `db:"address"`
		UserName     *string `db:"user_name"`
		UserAddress  *string `db:"user_address"`
		Amount       uint64  `db:"amount"`
		Entries      int64   `db:"entries"`
		RefID        int64   `db:"ref_id"`
		Active       bool    `db:"active"`
	}

	Winner struct {
		Date         string  `db:"date"`
		Info         string  `db:"info"`
		TransferBody *string `db:"transfer_body"`
	}

	Amount struct {
		Winner      uint64 `json:"winner"`
		Fund        uint64 `json:"fund"`
		Referrals   uint64 `json:"referrals"`
		Maintenance uint64 `json:"maintenance"`
	}

	Entry struct {
		ID        int64  `json:"id" db:"id"`
		AccountID int64  `json:"-" db:"account_id"`
		Hash      string `json:"hash" db:"hash"`
	}
)

func (a *Account) AddressUri(amount uint64) (string, error) {
	walletLock.Lock()
	r, err := Wallet.MakeUri(&monerorpc.MakeUriRequest{
		Address: a.Address,
		Amount:  amount,
	})
	walletLock.Unlock()
	if err != nil {
		return "", fmt.Errorf("AddressUri error %v", err)
	}
	return r.Uri, nil
}

func (a *Account) GetReferals() (int64, error) {
	db := MustDB()
	var total *int64
	err := db.Get(&total, `SELECT SUM(entries) FROM accounts WHERE ref_id = $1 AND active = 1 AND entries > 0`, a.ID)
	if total != nil {
		return *total, nil
	}
	return 0, err
}

func (a *Account) Save() error {
	db := MustDB()
	if a.ID == 0 {
		r, err := db.NamedExec(`INSERT INTO accounts (
			address_index,
			address,
			user_name,
			user_address,
			amount,
			entries,
			active,
			ref_id
			)
			VALUES (
			:address_index,
			:address,
			:user_name,
			:user_address,
			:amount,
			:entries,
			:active,
			:ref_id
			)`, a)
		if err != nil {
			return err
		}
		a.ID, err = r.LastInsertId()
		return err
	}
	_, err := db.NamedExec(`UPDATE accounts SET
		address_index = :address_index,
		address = :address,
		user_name = :user_name,
		user_address = :user_address,
		amount = :amount,
		entries = :entries,
		active = :active,
		ref_id = :ref_id
		WHERE id = :id`, a)
	return err
}

func GetAccount(userAddress string, userName *string, referrer *string) (*Account, error) {
	db := MustDB()
	account := &Account{}
	// first find your active account then inactive account
	err := db.Get(account, `SELECT * FROM accounts WHERE active = true AND user_address = $1`, userAddress)
	if err != nil {
		if !util.NoRows(err) {
			return nil, fmt.Errorf("GetAccount: select error %v", err)
		}
		err = db.Get(account, `SELECT * FROM accounts WHERE active = false ORDER BY address_index`)
		if err != nil {
			if !util.NoRows(err) {
				return nil, fmt.Errorf("GetAccount: select error %v", err)
			}
		}
	}

	if userName != nil && account.UserName == nil {
		acct := &Account{}
		err = db.Get(acct, `SELECT * FROM accounts WHERE user_name = $1`, *userName)
		if !util.NoRows(err) {
			return nil, ErrDuplicateUser
		}
	}
	updateAccount := false
	if referrer != nil && account.RefID == 0 {
		refAcct := &Account{}
		err = db.Get(refAcct, `SELECT * FROM accounts WHERE user_name = $1`, *referrer)
		if !util.NoRows(err) && refAcct.ID != account.ID {
			account.RefID = refAcct.ID
			updateAccount = true
		}
	}
	if account.Active && account.UserAddress != nil && *account.UserAddress == userAddress {
		if userName != nil && account.UserName == nil {
			account.UserName = userName
			updateAccount = true
		}
		if updateAccount {
			if err := account.Save(); err != nil {
				return nil, err
			}
		}
		// existing account just return after updatables
		return account, nil
	}
	if account.ID == 0 {
		walletLock.Lock()
		resp, err := Wallet.CreateAddress(&monerorpc.CreateAddressRequest{})
		walletLock.Unlock()
		if err != nil {
			return nil, fmt.Errorf("GetAccount: error wallet.create_address %v", err)
		}
		account.AddressIndex = resp.AddressIndex
		account.Address = resp.Address
	}
	account.UserName = userName
	account.UserAddress = &userAddress
	account.Amount = 0
	account.Active = true
	if err := account.Save(); err != nil {
		return nil, err
	}
	return account, nil
}

func GetIdByUsername(userName string) (int64, error) {
	db := MustDB()
	account := &Account{}
	err := db.Get(account, `SELECT * FROM accounts
	WHERE active = 1 AND user_name = $1`, userName)
	if err != nil {
		if !util.NoRows(err) {
			return 0, fmt.Errorf("CreateAccount: select error %v", err)
		}
		return 0, nil
	}
	return account.ID, nil
}

func GetWinner(dt string) (*WinnerInfo, error) {
	db := MustDB()
	winner := Winner{}
	var err error
	if dt == "" {
		err = db.Get(&winner, `SELECT * FROM winners ORDER BY date DESC`)
	} else {
		err = db.Get(&winner, `SELECT * FROM winners WHERE date = ?`, dt)
	}
	if err != nil {
		if util.NoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	winInfo := &WinnerInfo{Date: winner.Date}
	if err := json.Unmarshal([]byte(winner.Info), winInfo); err != nil {
		return nil, fmt.Errorf("GetWinner error %v", err)
	}
	accounts := winInfo.Accounts
	winInfo.Accounts = make(map[string][]int)
	for k, v := range accounts {
		winInfo.Accounts[k[0:5]+"..."+k[len(k)-5:]] = v
	}
	return winInfo, nil
}

func TotalEntries() (int64, error) {
	total, err := GetMetadata("entry_id", "0")
	if err != nil {
		return 0, err
	}
	t, _ := strconv.ParseInt(total, 10, 64)
	return t, nil
}

func GetDistributedAmounts(all bool) (*Amount, error) {
	walletLock.Lock()
	balance, err := Wallet.GetBalance(&monerorpc.GetBalanceRequest{})
	walletLock.Unlock()
	if err != nil {
		return nil, fmt.Errorf("GetDistributedAmounts error %v", err)
	}
	if all && balance.UnlockedBalance != balance.Balance {
		return nil, fmt.Errorf("GetDistributedAmounts has locked balance")
	}
	// keep 1XMR reserve for transfer fees
	var bal float64
	if 1e12 > balance.Balance {
		bal = 0
	} else {
		bal = float64(balance.Balance - 1e12)
	}
	amt := &Amount{
		Winner:      uint64(bal * .7),
		Fund:        uint64(bal * .05),
		Referrals:   uint64(bal * .15),
		Maintenance: uint64(bal * .1),
	}
	return amt, nil
}

func GetEntries(accountID int64, page int) ([]Entry, error) {
	db := MustDB()
	var (
		entries []Entry
		args    []interface{}
	)
	sql := `SELECT * FROM entries`
	limit := 100
	if accountID > 0 {
		sql += ` WHERE account_id = ?`
		args = append(args, accountID)
	}

	sql += fmt.Sprintf(` ORDER BY id LIMIT %d OFFSET %d`, limit, (page-1)*limit)

	if err := db.Select(&entries, sql, args...); err != nil {
		return nil, fmt.Errorf("GetEntries error %v", err)
	}
	return entries, nil
}
