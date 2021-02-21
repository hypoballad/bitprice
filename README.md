# bitprice

Store bitcoin prices in a database and retrieve them with gRPC

## usage

```
Usage:
  bitprice [command]

Available Commands:
  cli         Sample client for retrieving price information from bitcoin price server
  help        Help about any command
  server      Start the bitcoin price server.

Flags:
      --addr string     server address (default ":9911")
      --config string   config file (default is $HOME/.bitprice.yaml)
      --debug           debug mode
  -h, --help            help for bitprice
  -t, --toggle          Help message for toggle

Use "bitprice [command] --help" for more information about a command.
```

## server

``` 
Store the price of bitcoin. gRPC can be used to get the price of BTC_USD and BTC_JPY. For example:

./bitprice server

Usage:
  bitprice server [flags]

Flags:
      --db string         db directory (default "bitprice_db")
  -h, --help              help for server
      --spec string       cron spec (default "@every 10s")
      --truncate string   truncate duration (default "10s")

Global Flags:
      --addr string     server address (default ":9911")
      --config string   config file (default is $HOME/.bitprice.yaml)
      --debug           debug mode
```

## cli

```
Sample client for retrieving price information from bitcoin price server. For example:

./bitprice cli usd 2021-02-21T16:40:00+09:00

Obtained by specifying the start time and end time.
./bitprice cli jpy 2021-02-21T16:40:00+09:00 2021-02-21T17:14:00+09:00

Usage:
  bitprice cli [flags]

Flags:
  -h, --help     help for cli
      --indent   json marshal indent

Global Flags:
      --addr string     server address (default ":9911")
      --config string   config file (default is $HOME/.bitprice.yaml)
      --debug           debug mode
```

## update grpc proto command

```
protoc  --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  marketprice/marketprice.proto
```