package bisq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Currency struct {
	CurrencyCode string  `json:"currencyCode"`
	Price        float64 `json:"price"`
	TimestampSec int64   `json:"timestampSec"`
	Provider     string  `json"provider"`
}

type Market struct {
	Data []Currency `json:"data"`
}

func GetAllMarketPrices() (market Market, err error) {
	resp, e := http.Get("https://price.bisq.wiz.biz/getAllMarketPrices")
	if e != nil {
		err = fmt.Errorf("http get error: %v", e)
		return
	}
	defer resp.Body.Close()
	bytes, e := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read response error: %v", e)
		return
	}

	if e := json.Unmarshal(bytes, &market); err != nil {
		err = fmt.Errorf("json unmarshal error: %v", e)
		return
	}
	// fmt.Printf("%+v\n", market)
	//parsePrices(prices)
	return
}

func saveCurrency(db *leveldb.DB, currency Currency, truncate string, debug bool) (err error) {
	tm := time.Unix(int64(currency.TimestampSec/1000), 0)
	duration, err := time.ParseDuration(truncate)
	if err != nil {
		return
	}
	keyts := tm.UTC().Truncate(duration)
	key := fmt.Sprintf("%s::%d", currency.CurrencyCode, keyts.Unix())
	b, err := json.Marshal(currency)
	if err != nil {
		return
	}
	if debug {
		fmt.Printf("%s (%s)\n%s\n", key, keyts.Format("2006-01-02 15:04:05"), string(b))
	}
	if err = db.Put([]byte(key), b, nil); err != nil {
		return
	}
	return
}

func SaveMarket(db *leveldb.DB, market Market, truncate string, debug bool) (err error) {
	for _, currency := range market.Data {
		switch currency.CurrencyCode {
		case "USD", "JPY":
			if err = saveCurrency(db, currency, truncate, debug); err != nil {
				return
			}
		default:
			continue
		}
	}
	return
}

type Price int

const (
	USD Price = iota
	JPY
)

func (p Price) String() string {
	return [...]string{"USD", "JPY"}[p]
}

func btcprice(db *leveldb.DB, price Price, in int64, truncate string, debug bool) (out Currency, err error) {
	// tm := time.Unix(in, 0)
	// var d time.Duration
	// d, err = time.ParseDuration(truncate)
	// if err != nil {
	// 	return
	// }
	// uts := tm.Truncate(d).Unix()
	// key := fmt.Sprintf("%s::%d", price.String(), uts)
	key, err := priceKey(price, in, truncate)
	if err != nil {
		return
	}
	if debug {
		log.Printf("[btcprice]key %s\n", key)
	}
	var data []byte
	data, err = db.Get([]byte(key), nil)
	if err != nil {
		if debug {
			log.Printf("db get error: %v", err)
		}
		return
	}
	if debug {
		log.Printf("[btcprice]data %s\n", string(data))
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return
	}
	return
}

func BtcUsd(db *leveldb.DB, in int64, truncate string, debug bool) (out Currency, err error) {
	return btcprice(db, USD, in, truncate, debug)
}

func BtcJpy(db *leveldb.DB, in int64, truncate string, debug bool) (out Currency, err error) {
	return btcprice(db, JPY, in, truncate, debug)
}

func priceKey(price Price, uts int64, truncate string) (key string, err error) {
	tm := time.Unix(uts, 0)
	var d time.Duration
	d, err = time.ParseDuration(truncate)
	if err != nil {
		return
	}
	ts := tm.UTC().Truncate(d)
	key = fmt.Sprintf("%s::%d", price.String(), ts.Unix())
	return
}

func btcPriceRange(db *leveldb.DB, price Price, start, end int64, truncate string, debug bool) (out []Currency, err error) {
	startkey, err := priceKey(price, start, truncate)
	if err != nil {
		return
	}
	endkey, err := priceKey(price, end, truncate)
	if err != nil {
		return
	}
	out = []Currency{}
	if debug {
		log.Printf("start: %s, end: %s\n", startkey, endkey)
	}
	iter := db.NewIterator(&util.Range{Start: []byte(startkey), Limit: []byte(endkey)}, nil)
	for iter.Next() {
		// Use key/value.
		//...

		value := iter.Value()
		if debug {
			key := iter.Key()
			log.Printf("key: %s, value: %s\n", string(key), string(value))
		}
		var resp Currency
		if err = json.Unmarshal(value, &resp); err != nil {
			return
		}
		out = append(out, resp)
	}
	iter.Release()
	err = iter.Error()
	return
}

func BtcUsdArray(db *leveldb.DB, start, end int64, truncate string, debug bool) (out []Currency, err error) {
	return btcPriceRange(db, USD, start, end, truncate, debug)
}

func BtcJpyArray(db *leveldb.DB, start, end int64, truncate string, debug bool) (out []Currency, err error) {
	return btcPriceRange(db, JPY, start, end, truncate, debug)
}
