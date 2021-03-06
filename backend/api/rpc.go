package api

import (
	"fmt"
	"moneropot/db"
	"moneropot/util"
	"net/http"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

var (
	walletAddress string
)

func (s *Server) GetWalletAddress(r *http.Request) interface{} {
	if walletAddress == "" {
		address, err := db.GetWalletAddress()
		if err != nil {
			return err
		}
		walletAddress = address
	}
	return walletAddress
}

func (s *Server) RunPickWinner(r *http.Request) interface{} {
	if !s.isAdmin(r) {
		return errAuth
	}
	db.RunPickWinnerManually()
	return "OK"
}

func (s *Server) Contact(r *http.Request) interface{} {
	type request struct {
		Contact string `json:"contact"`
		Message string `json:"message" validate:"required"`
	}
	var req request
	if err := s.bind(r, &req); err != nil {
		return err
	}
	cKey := "contact:" + s.RealIP(r)
	_, ok := util.Cache.Get(cKey)
	if !ok {
		util.SendEvent(fmt.Sprintf(`Contact: %s
Message: %s`, req.Contact, req.Message))
		util.Cache.Set(cKey, "1", time.Hour*1)
		return nil
	}
	return errRateLimit
}

func (s *Server) FlushWinPayload(r *http.Request) interface{} {
	if !s.isAdmin(r) {
		return errAuth
	}
	m := s.QueryParam(r, "month")
	if m == "" {
		return errNotFound
	}
	return db.FlushWinPayload(m)
}

func (s *Server) QrCode(r *http.Request) interface{} {
	addr := s.QueryParam(r, "addr")
	amt := s.QueryParam(r, "amt")
	if amt != "" {
		addr = "monero:" + addr + "?tx_amount=" + amt
	}
	png, err := qrcode.Encode(addr, qrcode.Medium, 256)
	if err != nil {
		return err
	}
	return rpcType{
		contentType: "image/png",
		body:        png,
	}
}

func (s *Server) Winner(r *http.Request) interface{} {
	dt := s.QueryParam(r, "dt")
	var (
		w   *db.WinnerInfo
		err error
	)
	item, ok := util.Cache.Get(dt)
	if !ok {
		w, err = db.GetWinner(dt)
		if err != nil {
			return err
		}
		util.Cache.Set(dt, w, time.Hour*24)
	} else {
		w = item.(*db.WinnerInfo)
	}
	return w
}
