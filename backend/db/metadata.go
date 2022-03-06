package db

import (
	"fmt"
	"log"
	"moneropot/util"
	"strconv"
	"strings"
)

type (
	Metadata struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}
)

func GetMetadata(key string, def string) (string, error) {
	db := MustDB()
	md := &Metadata{}
	if err := db.Get(md, `SELECT * FROM metadata WHERE key = $1`, key); err != nil {
		if !util.NoRows(err) {
			return def, err
		}
	}
	if md.Value == "" {
		md.Value = def
	}
	return md.Value, nil
}

func SetMetadata(key string, value string) error {
	db := MustDB()
	md := &Metadata{}
	sql := "UPDATE metadata SET value = $2 WHERE key = $1"
	if err := db.Get(md, `SELECT * FROM metadata WHERE key = $1`, key); err != nil {
		if !util.NoRows(err) {
			return err
		}
		sql = "INSERT INTO metadata (key, value) VALUES ($1, $2)"
	}
	_, err := db.Exec(sql, key, value)
	return err
}

func LastHeight() (uint64, error) {
	lastHeight, err := GetMetadata("last_height", "0")
	if err != nil {
		return 0, err
	}
	var h uint64
	h, err = strconv.ParseUint(lastHeight, 10, 64)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func SetCurrentPrice() error {
	price, err := GetMetadata("current_price", "")
	if err != nil {
		return fmt.Errorf("SetCurrentPrice error %v", err)
	}
	dt := util.UtcNow().Format(DateFormat)
	if !strings.Contains(price, dt) {
		// price is outdated need to update
		xmrPrice, err := util.XmrToUSD(true)
		if err != nil {
			return fmt.Errorf("SetCurrentPrice error %v", err)
		}
		newPrice := util.CalcEntryFromUSD(xmrPrice)
		if err := SetMetadata("current_price", dt+":"+strconv.FormatUint(newPrice, 10)); err != nil {
			return fmt.Errorf("SetCurrentPrice error %v", err)
		}
		CurrentPrice = newPrice
		util.Cache.Delete("info")
		util.PublishTopic("", "info")
		log.Printf("Updated Price: %d", CurrentPrice)
		return nil
	}
	p := strings.Split(price, ":")
	CurrentPrice, _ = strconv.ParseUint(p[1], 10, 64)
	log.Printf("Loaded Price: %d", CurrentPrice)
	return nil
}
