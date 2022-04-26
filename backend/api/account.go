package api

import (
	"fmt"
	"log"
	"moneropot/db"
	"moneropot/util"
	"net/http"
	"strconv"
	"time"

	"moneropot/monerorpc"
)

func (s *Server) handlePostAccount() http.HandlerFunc {
	type (
		request struct {
			UserName *string `json:"username" validate:"omitempty,username"`
			Address  string  `json:"address" validate:"required,invalid"`
			Referrer *string `json:"ref"`
		}
		response struct {
			ID           int64   `json:"id"`
			UserName     *string `json:"username"`
			Address      string  `json:"address"`
			AddressUri   string  `json:"address_uri"`
			UserAddress  string  `json:"user_address"`
			Entries      int64   `json:"entries"`
			RemainingXMR string  `json:"xmr"`
			Referrals    int64   `json:"referrals"`
		}
	)
	return s.handler(func(r *http.Request) interface{} {
		var req request
		if err := s.bind(r, &req); err != nil {
			return err
		}
		if db.IsValidAddress(req.Address) != nil {
			return newValidationErr("address", "invalid")
		}
		acct, err := db.GetAccount(req.Address, req.UserName, req.Referrer)
		if err != nil {
			if err == db.ErrDuplicateUser {
				return newValidationErr("username", "exists")
			}
			return err
		}
		resp := &response{
			ID:           acct.ID,
			UserName:     acct.UserName,
			Address:      acct.Address,
			Entries:      acct.Entries,
			RemainingXMR: monerorpc.XMRToDecimal(acct.Amount),
		}
		if acct.UserAddress != nil {
			userAddress := *acct.UserAddress
			resp.UserAddress = userAddress[0:5] + "..." + userAddress[len(userAddress)-5:]
		}
		refs, err := acct.GetReferals()
		if err != nil {
			return err
		}
		uri, err := acct.AddressUri(db.CurrentPrice)
		if err != nil {
			return err
		}
		resp.AddressUri = uri
		resp.Referrals = refs
		return resp
	})
}

func (s *Server) handleGetInfo() http.HandlerFunc {
	type response struct {
		WinAmount         string         `json:"win_amount"`
		AffiliateAmount   string         `json:"ref_amount"`
		MaintenanceAmount string         `json:"maint_amount"`
		EntryPrice        string         `json:"entry_price"`
		XmrRate           string         `json:"xmr_rate"`
		TotalEntries      int64          `json:"entries"`
		UntilDraw         int64          `json:"until_draw"`
		UntilPrice        int64          `json:"until_price"`
		WalletAddress     string         `json:"address"`
		WalletOffline     bool           `json:"wallet_offline"`
		SignKey           string         `json:"sign_key"`
		LastWinner        *db.WinnerInfo `json:"last_winner"`
	}
	return s.handler(func(r *http.Request) interface{} {
		var resp response
		cKey := "info"
		item, ok := util.Cache.Get(cKey)
		if !ok {
			// load sync first time then refresh on background on demand
			if util.XmrPrice == 0 {
				util.XmrToUSD(false)
			} else {
				go func() {
					util.XmrToUSD(false)
				}()
			}
			resp = response{}
			rate := util.XmrPrice
			resp.EntryPrice = monerorpc.XMRToDecimal(db.CurrentPrice)
			resp.XmrRate = util.USDToDecimal(rate)
			resp.UntilDraw = int64(db.StartOfMonth().Seconds())
			resp.UntilPrice = int64(db.AtHourMinute(3, 0).Seconds())
			entries, err := db.TotalEntries()
			if err != nil {
				return err
			}
			resp.TotalEntries = entries
			signKey, err := db.GetMetadata("sign_key", "")
			if err != nil {
				return err
			}
			resp.SignKey = signKey
			lastWinner, err := db.GetWinner("")
			if err != nil {
				return err
			}
			resp.LastWinner = lastWinner
			// wallet calls
			amt, err := db.GetDistributedAmounts(false)
			if err != nil {
				resp.WalletOffline = true
				log.Println("WalletOffline:", err)
			} else {
				resp.WinAmount = monerorpc.XMRToDecimal(amt.Winner)
				resp.AffiliateAmount = monerorpc.XMRToDecimal(amt.Referrals)
				resp.MaintenanceAmount = monerorpc.XMRToDecimal(amt.Maintenance)
			}
			address := s.GetWalletAddress(r)
			addr, ok := address.(string)
			if ok {
				resp.WalletAddress = addr
			} else {
				resp.WalletOffline = true
				log.Println("WalletOffline:", address)
			}
			// end wallet calls
			d := time.Minute * 5
			if !util.Config.Production {
				d = time.Second * 1
			}
			util.Cache.Set(cKey, resp, d)
		} else {
			resp = item.(response)
		}
		return resp
	})
}

func (s *Server) handleGetEntries() http.HandlerFunc {
	return s.handler(func(r *http.Request) interface{} {
		var (
			aID     int64
			entries []db.Entry
			err     error
		)
		page := 1
		acctId := r.URL.Query()["a"]
		pg := r.URL.Query()["p"]
		if len(acctId) > 0 {
			aID, _ = strconv.ParseInt(acctId[0], 10, 64)
		}
		if len(pg) > 0 {
			page, _ = strconv.Atoi(pg[0])
		}
		if page < 1 {
			page = 1
		}
		cKey := fmt.Sprintf("entries:%d:%d", aID, page)
		item, ok := util.Cache.Get(cKey)
		if !ok {
			entries, err = db.GetEntries(aID, page)
			if err != nil {
				return err
			}
			d := time.Minute * 5
			if !util.Config.Production {
				d = time.Second * 1
			}
			util.Cache.Set(cKey, entries, d)
		} else {
			entries = item.([]db.Entry)
		}
		return entries
	})
}
