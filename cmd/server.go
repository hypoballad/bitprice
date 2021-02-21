/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	bisq "github.com/hypoballad/bitprice/bisq"
	market "github.com/hypoballad/bitprice/marketprice"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
)

type server struct {
	market.UnimplementedMarketPriceServer
}

func (s server) BtcJpyArray(ctx context.Context, in *market.TimeRange) (*market.PriceArray, error) {
	var items market.PriceArray
	out, err := bisq.BtcJpyArray(db, in.Start, in.End, viper.GetString("server.truncate"), viper.GetBool("root.debug"))
	if err != nil {
		return &items, err
	}
	item := []*market.PriceResp{}
	for _, o := range out {
		resp := market.PriceResp{
			Code:     o.CurrencyCode,
			Uts:      o.TimestampSec,
			Price:    float32(o.Price),
			Provider: o.Provider,
		}
		item = append(item, &resp)
	}
	items.Items = item
	return &items, nil
}

func (s server) BtcUsdArray(ctx context.Context, in *market.TimeRange) (*market.PriceArray, error) {
	var items market.PriceArray
	out, err := bisq.BtcUsdArray(db, in.Start, in.End, viper.GetString("server.truncate"), viper.GetBool("root.debug"))
	if err != nil {
		return &items, err
	}
	item := []*market.PriceResp{}
	for _, o := range out {
		resp := market.PriceResp{
			Code:     o.CurrencyCode,
			Uts:      o.TimestampSec,
			Price:    float32(o.Price),
			Provider: o.Provider,
		}
		item = append(item, &resp)
	}
	items.Items = item
	return &items, nil
}

func (s server) BtcUsd(ctx context.Context, in *market.TimeParam) (*market.PriceResp, error) {
	var item market.PriceResp
	fmt.Printf("server btcusd: %+v\n", in)
	out, err := bisq.BtcUsd(db, in.GetUts(), viper.GetString("server.truncate"), viper.GetBool("root.debug"))
	if err != nil {
		return &item, err
	}
	item.Code = out.CurrencyCode
	item.Uts = out.TimestampSec
	item.Price = float32(out.Price)
	item.Provider = out.Provider
	return &item, nil
}

func (s server) BtcJpy(ctx context.Context, in *market.TimeParam) (*market.PriceResp, error) {
	var item market.PriceResp
	out, err := bisq.BtcJpy(db, in.GetUts(), viper.GetString("server.truncate"), viper.GetBool("root.debug"))
	if err != nil {
		return &item, err
	}
	item.Code = out.CurrencyCode
	item.Uts = out.TimestampSec
	item.Price = float32(out.Price)
	item.Provider = out.Provider
	return &item, nil
}

var db *leveldb.DB
var err error

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the bitcoin price server.",
	Long: `Store the price of bitcoin. gRPC can be used to get the price of BTC_USD and BTC_JPY. For example:

./bitprice server`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err = leveldb.OpenFile(viper.GetString("server.db"), nil)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		// var err error
		c := cron.New()
		spec := viper.GetString("server.spec")
		c.AddFunc(spec, func() {
			market, err := bisq.GetAllMarketPrices()
			if err != nil {
				log.Printf("get all market prices error: %v\n", err)
				return
			}
			if err := bisq.SaveMarket(db, market, viper.GetString("server.truncate"), viper.GetBool("root.debug")); err != nil {
				log.Printf("save market error: %v\n", err)
				return
			}
		})
		c.Start()

		//fmt.Println("server called")
		addr := viper.GetString("root.addr")
		var lis net.Listener
		lis, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalln("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		log.Printf("listen to %s\n", addr)
		market.RegisterMarketPriceServer(s, &server{})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// serverCmd.Flags().String("addr", ":9991", "server addr")
	serverCmd.Flags().String("spec", "@every 10s", "cron spec")
	serverCmd.Flags().String("truncate", "10s", "truncate duration")
	serverCmd.Flags().String("db", "bitprice_db", "db directory")

	// viper.BindPFlag("server.addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.spec", serverCmd.Flags().Lookup("spec"))
	viper.BindPFlag("server.truncate", serverCmd.Flags().Lookup("truncate"))
	viper.BindPFlag("server.db", serverCmd.Flags().Lookup("db"))
}
