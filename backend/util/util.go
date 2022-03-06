package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	lastPriceUpdate time.Time
	XmrPrice        uint64
	alphaNums       = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

var (
	Now   = time.Now
	Cache *cache.Cache
)

func init() {
	Cache = cache.New(time.Hour*24, time.Hour*1)
}

func UtcNow() time.Time {
	return Now().UTC()
}

func RandomString(n int) string {
	b := make([]rune, n)
	at := len(alphaNums)
	for i := range b {
		b[i] = alphaNums[rand.Intn(at)]
	}
	return string(b)
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func CopyFile(src string, dst string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)
	return err
}

func NoRows(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no rows in result set")
}

func GetURL(url string, dst interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GetURL: error %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("GetURL: read body error %v", err)
	}

	err = json.Unmarshal(body, dst)
	if err != nil {
		return fmt.Errorf("GetURL: json unmarshal error %v", err)
	}
	return nil
}

func PriceMarket(market string) (float64, error) {
	type response struct {
		Success bool   `json:"success"`
		Price   string `json:"price"`
	}
	resp := &response{}
	if err := GetURL("https://tradeogre.com/api/v1/ticker/"+market, resp); err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf("PriceMarket: failed to get %s", market)
	}
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("PriceMarket: parseFloat error %v", err)
	}
	return price, nil
}

func XmrToUSD(force bool) (uint64, error) {
	if !force && XmrPrice > 0 && time.Since(lastPriceUpdate).Minutes() <= 5 {
		return XmrPrice, nil
	}
	var (
		wg             sync.WaitGroup
		btcUSD, xmrBTC float64
		errBTC, errXMR error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		btcUSD, errBTC = PriceMarket("USDT-BTC")
	}()
	go func() {
		defer wg.Done()
		xmrBTC, errXMR = PriceMarket("BTC-XMR")
	}()
	wg.Wait()
	if errBTC != nil {
		return 0, fmt.Errorf("XmrToUSD btc price error %v", errBTC)
	}
	if errXMR != nil {
		return 0, fmt.Errorf("XmrToUSD xmr price error %v", errXMR)
	}
	XmrPrice = uint64(btcUSD * xmrBTC * 100000000)
	return XmrPrice, nil
}

func CalcEntryFromUSD(xmrUSDPrice uint64) uint64 {
	usd := (float64(xmrUSDPrice) / 100000000)
	return uint64((2.5 / usd) * 1e12)
}

func USDToDecimal(usd uint64) string {
	str0 := fmt.Sprintf("%09d", usd)
	l := len(str0)
	return str0[:l-8] + "." + str0[l-8:]
}

func HashMatchAlign(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	m := 0
	t := len(a)
	for i := 0; i < t; i++ {
		if a[i] == b[i] {
			m++
		}
	}
	return m
}

func HashMatchChar(a, b string) int {
	m := 0
	aMap := make(map[byte]int)
	t := len(a)
	for i := 0; i < t; i++ {
		v, ok := aMap[a[i]]
		if !ok {
			aMap[a[i]] = 0
		}
		aMap[a[i]] = v + 1
	}
	t = len(b)
	for i := 0; i < t; i++ {
		v, ok := aMap[b[i]]
		if ok && v > 0 {
			m++
			aMap[b[i]] = v - 1
		}
	}
	return m
}

func SignEntry(id int64, key string) string {
	s256 := sha256.New()
	s256.Write([]byte(key + strconv.FormatInt(id, 10)))
	return fmt.Sprintf("%x", s256.Sum(nil))
}
