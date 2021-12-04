package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elideveloper/trades-puller/currency"
	"github.com/elideveloper/trades-puller/poloniex"
	"github.com/elideveloper/trades-puller/trade"
	"github.com/elideveloper/trades-puller/wsclient"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("run application")

	polonClient, err := poloniex.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// request information about given currency pairs
	for i := range currency.PairsList {
		cmd, err := poloniex.NewSubscribeCmd(poloniex.FormatPair(currency.PairsList[i]))
		if err != nil {
			log.Fatal(err)
		}
		err = polonClient.Send(cmd)
		if err != nil {
			log.Fatal(err)
		}
	}

	terminate := make(chan struct{})
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	// listen to termination signal
	go func() {
		<-termChan
		cancel()
	}()

	go listenToTrades(ctxzap.ToContext(ctx, logger), polonClient, terminate)

	<-terminate
	logger.Info("terminate application")
}

// listenToTrades receives websocket messages and prints only trades
func listenToTrades(ctx context.Context, polonClient wsclient.WsClient, terminate chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			{
				err := polonClient.Shutdown()
				if err != nil {
					ctxzap.Extract(ctx).Error(fmt.Sprintf("failed shutting down client: %s", err.Error()))
				} else {
					ctxzap.Extract(ctx).Info("success client shutdown")
				}
				terminate <- struct{}{}
				return
			}
		default:
			{
				msg, err := polonClient.Receive()
				if err != nil {
					ctxzap.Extract(ctx).Error(fmt.Sprintf("failed receiving a message: %s", err.Error()))
					continue
				}

				tradeFields, err := poloniex.FetchTradeFields(msg)
				if err != nil {
					ctxzap.Extract(ctx).Error(fmt.Sprintf("failed fetching trade fields: %s", err.Error()))
					continue
				}

				// if find trades then print all of them
				for i := range tradeFields {
					t := trade.NewRecentTrade(tradeFields[i])
					//fmt.Println(t) // raw output for quick printing
					t.PrettyPrint(ctx)
				}
			}
		}
	}
}
