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
	"encoding/json"
	"fmt"
	"log"
	"time"

	bisq "github.com/hypoballad/bitprice/bisq"
	market "github.com/hypoballad/bitprice/marketprice"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func btcprice(conn *grpc.ClientConn, price bisq.Price, ts string) (err error) {
	c := market.NewMarketPriceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	uts, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return
	}
	in := &market.TimeParam{Uts: uts.UTC().Unix()}
	var item *market.PriceResp
	switch price {
	case bisq.USD:
		item, err = c.BtcUsd(ctx, in)
	case bisq.JPY:
		item, err = c.BtcJpy(ctx, in)
	}
	if err != nil {
		return
	}
	var out []byte
	if viper.GetBool("cli.indent") {
		out, err = json.MarshalIndent(item, "", "	")
	} else {
		out, err = json.Marshal(item)
	}

	if err != nil {
		return
	}
	fmt.Println(string(out))
	return
}

func btcJpy(conn *grpc.ClientConn, ts string) error {
	return btcprice(conn, bisq.JPY, ts)
}

func btcUsd(conn *grpc.ClientConn, ts string) error {
	return btcprice(conn, bisq.USD, ts)
}

func btcPriceArray(conn *grpc.ClientConn, price bisq.Price, start, stop string) (err error) {
	c := market.NewMarketPriceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	startuts, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return
	}
	stoputs, err := time.Parse(time.RFC3339, stop)
	if err != nil {
		return
	}
	in := &market.TimeRange{
		Start: startuts.UTC().Unix(),
		End:   stoputs.UTC().Unix(),
	}
	var itemArray *market.PriceArray
	switch price {
	case bisq.USD:
		itemArray, err = c.BtcUsdArray(ctx, in)
	case bisq.JPY:
		itemArray, err = c.BtcJpyArray(ctx, in)
	}
	if err != nil {
		return
	}

	//for _, item := range itemArray.Items {
	var b []byte
	if viper.GetBool("cli.indent") {
		b, err = json.MarshalIndent(itemArray.Items, "", "	")
	} else {
		b, err = json.Marshal(itemArray.Items)
	}

	if err != nil {
		return
	}
	fmt.Println(string(b))
	//}

	return
}

func btcUsdArray(conn *grpc.ClientConn, start, stop string) (err error) {
	return btcPriceArray(conn, bisq.USD, start, stop)
}

func btcJpyArray(conn *grpc.ClientConn, start, stop string) (err error) {
	return btcPriceArray(conn, bisq.JPY, start, stop)
}

// cliCmd represents the cli command
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Sample client for retrieving price information from bitcoin price server",
	Long: `Sample client for retrieving price information from bitcoin price server. For example:

./bitprice cli usd 2021-02-21T16:40:00+09:00

Obtained by specifying the start time and end time.
./bitprice cli jpy 2021-02-21T16:40:00+09:00 2021-02-21T17:14:00+09:00`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("cli called")
		currency := args[0]
		time1 := args[1]
		var time2 string
		if len(args) == 3 {
			time2 = args[2]
		}
		addr := viper.GetString("root.addr")
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v\n", err)
		}
		defer conn.Close()
		switch currency {
		case "usd":
			if time2 == "" {
				if err := btcUsd(conn, time1); err != nil {
					log.Fatalf("btc_usd error: %v\n", err)
				}
			} else {
				if err := btcUsdArray(conn, time1, time2); err != nil {
					log.Fatalf("btc_usd array error: %v\n", err)
				}
			}
		case "jpy":
			if time2 == "" {
				if err := btcJpy(conn, time1); err != nil {
					log.Fatalf("btc_jpy error: %v\n", err)
				}
			} else {
				if err := btcJpyArray(conn, time1, time2); err != nil {
					log.Fatalf("btc_jpy array error: %v\n", err)
				}
			}
		default:
			fmt.Errorf("The first argument is 'usd' or 'jpy'.")
		}
	},
}

func init() {
	rootCmd.AddCommand(cliCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cliCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cliCmd.Flags().Bool("indent", false, "json marshal indent")
	viper.BindPFlag("cli.indent", cliCmd.Flags().Lookup("indent"))
}
