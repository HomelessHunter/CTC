package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/HomelessHunter/CTC/db"
	dbModels "github.com/HomelessHunter/CTC/db/models"
	"github.com/HomelessHunter/CTC/wrapper"
	cryptoMarkets "github.com/HomelessHunter/CTC/wrapper/models/cryptoMarkets"
	other "github.com/HomelessHunter/CTC/wrapper/models/other"
	models "github.com/HomelessHunter/CTC/wrapper/models/telegram"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	uc           *other.UC
	userChannels map[int64]*other.UserChannels
	regexps      map[string]*regexp.Regexp
	session      *other.Session
	wg           sync.WaitGroup
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		// log.Fatal("$PORT must be set")
		port = "8080"
	}

	shutdownCtx, cancel := context.WithCancel(context.Background())
	dialer := &websocket.Dialer{
		NetDialContext:   (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   256,
		WriteBufferSize:  256,
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ReadBufferSize:        256,
			WriteBufferSize:       256,
		},
	}
	mongoClient, err := mongo.Connect(shutdownCtx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		panic(err)
	}

	coll := db.GetUserCollection(mongoClient)

	regexps = compileRegexp()
	uc = other.NewUC()
	session = other.NewSession()

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("TG")), updateFromTG(client, dialer, coll, shutdownCtx))
	// For testing only
	// mux.HandleFunc("/update", createUpdate(dialer, client, shutdownCtx, coll))

	server := &http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTSTP)
		<-sigs

		if len(session.Alerts()) > 0 {
			err := db.ShutdownSequence(coll, session.Alerts(), session.AlertsCount(), shutdownCtx)
			if err != nil {
				fmt.Printf("cannot initiate shutdowns sequence: %s\n", err)
			}
		}

		mongoClient.Disconnect(shutdownCtx)
		client.CloseIdleConnections()
		cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	err = startSqc(coll, dialer, client, shutdownCtx)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Connected")

	wrapper.SetWebhook(client)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}

	completed := make(chan int)
	go func() {
		wg.Wait()
		completed <- 1
	}()

	select {
	case <-completed:
		fmt.Println("Completed")
	case <-time.After(30 * time.Second):
		fmt.Println("Timed out")
	}
}

func startSqc(
	coll *mongo.Collection, dialer *websocket.Dialer,
	client *http.Client, ctx context.Context,
) error {
	users, err := db.GetUsersWithPairs(coll, false, ctx)
	if err != nil {
		fmt.Printf("starSqc: %s", err)
	}
	if len(users) == 0 {
		return nil
	}
	for _, v := range users {
		dbModels.SortByMarket(v.Alerts)

		userChan, err := other.NewUserChannels(
			other.WithUCCancel(make(map[string]context.CancelFunc)),
			other.WithUCShutdown(make(map[string]chan int)),
			other.WithUCSubscribeCh(make(map[string]chan other.PairSignal)),
			other.WithUCUnsubscribeCh(make(map[string]chan other.PairSignal)),
		)
		if err != nil {
			fmt.Printf("userChan: %s", err)
		}
		session.InitMarketsByID(v.UsedID)

		if dbModels.AlertExist(v.Alerts, wrapper.Binance) {
			wsQuery := &other.WSQuery{UserId: v.UsedID, Market: wrapper.Binance}
			userChan.SetShutdownCh(wsQuery.Market, make(chan int, 1))
			userChan.SetSubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
			userChan.SetUnsubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))

			uc.SetUC(v.UsedID, userChan)
			alerts, err := connectToWS(coll, dialer, client, wsQuery, ctx)
			if err != nil {
				if err == db.NoPairsErr {
				} else {
					userChan.DeleteShutdownCh(wsQuery.Market)
					userChan.DeleteSubscribeCh(wsQuery.Market)
					userChan.DeleteUnsubscribeCh(wsQuery.Market)
					return err
				}
			}
			if len(alerts) > 0 {
				session.AddAlerts(wsQuery.UserId, alerts...)
			}
		}

		if dbModels.AlertExist(v.Alerts, wrapper.Huobi) {
			wsQuery := &other.WSQuery{UserId: v.UsedID, Market: wrapper.Huobi}
			userChan.SetShutdownCh(wsQuery.Market, make(chan int, 1))
			userChan.SetSubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
			userChan.SetUnsubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))

			if _, ok := uc.GetUserChannels(v.UsedID); !ok {
				uc.SetUC(v.UsedID, userChan)
			}
			alerts, err := connectToWS(coll, dialer, client, &other.WSQuery{UserId: v.UsedID, Market: wrapper.Huobi}, ctx)
			if err != nil {
				if err == db.NoPairsErr {
				} else {
					userChan.DeleteShutdownCh(wsQuery.Market)
					userChan.DeleteSubscribeCh(wsQuery.Market)
					userChan.DeleteUnsubscribeCh(wsQuery.Market)
					return err
				}
			}
			if len(alerts) > 0 {
				session.AddAlerts(wsQuery.UserId, alerts...)
			}
		}
	}
	return nil
}

func alertHandler(
	dialer *websocket.Dialer, client *http.Client,
	wsQuery *other.WSQuery, ctx context.Context,
	coll *mongo.Collection,
) error {
	var alert *dbModels.Alert
	var err error

	alerts, err := db.GetAlerts(coll, wsQuery.UserId, ctx)
	if err != nil {
		return err
	}

	// Check if Pair already exists
	if pairExist(wsQuery.Market, wsQuery.Pair, alerts) {
		// Send user response
		return errors.New("pair already exists")
	}

	// Check if user have connected previously
	// If false cancel that connection and add a new token pair
	userChannel, ok := uc.GetUserChannels(wsQuery.UserId)
	if ok {
		alert, err = dbModels.NewAlert(dbModels.WithPair(strings.ToLower(wsQuery.Pair)), dbModels.WithTargetPrice(wsQuery.Price), dbModels.WithMarket(wsQuery.Market))
		if err != nil {
			return fmt.Errorf("cannot create new alert: %s", err)
		}

		if session.MarketExist(wsQuery.UserId, wsQuery.Market) {
			alert.Connected = true
			err = db.AddAlert(coll, wsQuery.UserId, alert, ctx)
			if err != nil {
				return fmt.Errorf("cannot add alert: %s", err)
			}
			session.AddAlerts(wsQuery.UserId, *alert)
			userChannel.SubscribeSignal(wsQuery.Market, other.PairSignal{Pair: wsQuery.Pair, Size: len(session.AlertsByID(wsQuery.UserId)) + 1})
			return nil
		}
		userChannel.SetShutdownCh(wsQuery.Market, make(chan int, 1))
		userChannel.SetSubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
		userChannel.SetUnsubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
		// userChannels[wsQuery.UserId] = userChannel
		err = db.AddAlert(coll, wsQuery.UserId, alert, ctx)
		if err != nil {
			return fmt.Errorf("cannot add alert: %s", err)
		}
	} else {
		alert, err = dbModels.NewAlert(dbModels.WithPair(strings.ToLower(wsQuery.Pair)), dbModels.WithTargetPrice(wsQuery.Price), dbModels.WithMarket(wsQuery.Market))
		if err != nil {
			return fmt.Errorf("cannot create new alert: %s", err)
		}
		err = db.AddAlert(coll, wsQuery.UserId, alert, ctx)
		if err != nil {
			return fmt.Errorf("cannot add alert: %s", err)
		}

		userChan, err := other.NewUserChannels(
			other.WithUCCancel(make(map[string]context.CancelFunc)),
			other.WithUCShutdown(make(map[string]chan int)),
			other.WithUCSubscribeCh(make(map[string]chan other.PairSignal)),
			other.WithUCUnsubscribeCh(make(map[string]chan other.PairSignal)),
		)
		if err != nil {
			fmt.Printf("userChan: %s", err)
		}

		userChan.SetShutdownCh(wsQuery.Market, make(chan int, 1))
		userChan.SetSubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
		userChan.SetUnsubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))

		session.InitMarketsByID(wsQuery.UserId)

		// userChannels[wsQuery.UserId] = userChan
		// !!!!!!!!!!!!!!!!!!!!!!!!!!
		uc.SetUC(wsQuery.UserId, userChan)

	}
	// fmt.Println(userChannels[wsQuery.UserId])

	_, err = connectToWS(coll, dialer, client, wsQuery, ctx)
	if err != nil {
		return err
	}

	return nil
}

func connectToWS(
	coll *mongo.Collection,
	dialer *websocket.Dialer, client *http.Client,
	wsQuery *other.WSQuery, shutdownSrv context.Context,
) ([]dbModels.Alert, error) {
	// userChannel := userChannels[wsQuery.UserId]
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	userChannel, _ := uc.GetUserChannels(wsQuery.UserId)
	pairs, alerts, err := db.GetPairsByMarket(coll, wsQuery.UserId, wsQuery.Market, false, shutdownSrv)
	if err != nil {
		fmt.Printf("cannot get pairs by market %s for user with %d id: %s\n", wsQuery.Market, wsQuery.UserID(), err)
		return nil, err
	}
	// Connect to Market
	conn, err := wrapper.TickerConnect(wsQuery.Market, pairs, dialer, client)
	if err != nil {
		return nil, err
	}

	err = db.UpdateAlertsSqc(coll, wsQuery.UserId, alerts, true, shutdownSrv)
	if err != nil {
		conn.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(shutdownSrv)
	userChannel.SetCancel(wsQuery.Market, cancel)

	// userChannels[wsQuery.UserId] = userChannel

	// Close websocket connection
	go func(ctx context.Context, shutdownSrv context.Context, conn *websocket.Conn, userChannel *other.UserChannels, wsQuery *other.WSQuery) {
		defer fmt.Println("Websocket CLOSED")
		<-ctx.Done()
		session.SetMarketByID(wsQuery.UserId, wsQuery.Market, false)
		err := conn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection closed: %s\n", err)
			return
		}
		// select {
		// case <-shutdownSrv.Done():
		// 	userChannel.Shutdown(wsQuery.Market)
		// 	conn.Close()
		// case <-ctx.Done():
		// 	session.SetMarketByID(wsQuery.UserId, wsQuery.Market, false)
		// 	err := conn.Close()
		// 	if err != nil {
		// 		fmt.Fprintf(os.Stderr, "Connection closed: %s\n", err)
		// 		return
		// 	}
		// }
	}(ctx, shutdownSrv, conn, userChannel, wsQuery)

	session.AddAlerts(wsQuery.UserId, alerts...)
	session.SetMarketByID(wsQuery.UserId, wsQuery.Market, true)

	go checkPrice(conn, dialer, client, wsQuery, coll, shutdownSrv)

	fmt.Println("Connect Finished")
	return alerts, nil
}

func checkPrice(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	wg.Add(1)
	fmt.Println("CheckPrice started")
	defer func() {
		fmt.Println("CheckPrice is DONE")
		wg.Done()
	}()
	switch wsQuery.Market {
	case wrapper.Huobi:
		go checkPriceHu(conn, dialer, client, wsQuery, coll, shutdownSrv)
	case wrapper.Binance:
		go checkPriceBi(conn, dialer, client, wsQuery, coll, shutdownSrv)
	}
}

func checkPriceBi(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	wg.Add(1)
	defer func() {
		fmt.Println("checkPriceBi is DONE")
		wg.Done()
	}()
	fmt.Println("CheckPriceBi started")
	close, cancel := context.WithCancel(shutdownSrv)
	// userChannel := userChannels[wsQuery.UserId]
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	userChannel, _ := uc.GetUserChannels(wsQuery.UserId)
	alerts := session.AlertsByID(wsQuery.UserId)
	if len(alerts) > 0 {
		dbModels.SortByHEX(alerts)
	}

	go func() {
		wg.Add(1)
		defer func() {
			wg.Done()
			fmt.Println("Binance SUB/UNSUB")
		}()
		for {
			select {
			case pair := <-userChannel.SubscribeCh(wsQuery.Market):
				fmt.Println("SUBBIN")
				alerts = session.AlertsByID(wsQuery.UserId)
				dbModels.SortByHEX(alerts)
				err := wrapper.SubscribeBi(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = wrapper.SendAlertConfirmed(client, wsQuery.ChatId)
				if err != nil {
					fmt.Printf("checkPriceBi: %v", err)
				}
			case pair := <-userChannel.UnsubscribeCh(wsQuery.Market):
				fmt.Println("UNSUBIN")
				alerts = session.AlertsByID(wsQuery.UserId)
				if len(alerts) > 0 {
					dbModels.SortByHEX(alerts)
				}
				err := wrapper.UnsubBi(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				if pair.Size <= 0 {
					fmt.Println("Bin pairs empty")
					userChannel.Shutdown(wsQuery.Market)
					userChannel.Cancel(wsQuery.Market)
					return
				}
			case <-shutdownSrv.Done():
				return
			case <-close.Done():
				return
			}
		}
	}()

	ticker := cryptoMarkets.NewTickerBi()
	for {
		err := conn.ReadJSON(ticker)
		if err != nil {
			cancel()
			fmt.Println("ReadJson: ", err)
			handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
			return
		}

		symbol := ticker.GetSymbol()
		if symbol != "" {
			alert, _ := dbModels.NewAlert(dbModels.WithMarket(wrapper.Binance), dbModels.WithPair(symbol))
			i, err := alert.Find(alerts)
			if err != nil {
				alerts = session.AlertsByID(wsQuery.UserId)
				if len(alerts) > 0 {
					dbModels.SortByHEX(alerts)
				}
				fmt.Println(alerts)
				fmt.Println("Find", err)
			}
			if len(alerts) == 0 {
				cancel()
				return
			}
			alert = &alerts[i]
			lastPrice, err := ticker.GetLastPrice()
			if err != nil {
				fmt.Println("lastPrice", err)
			}

			if lastPrice <= alert.TargetPrice+(alert.TargetPrice*0.01) && lastPrice >= alert.TargetPrice-(alert.TargetPrice*0.01) {
				// send notification
				if alert.LastSignal.IsZero() || time.Now().UTC().UnixMilli() >= (alert.LastSignal.Add(15*time.Minute)).UnixMilli() {
					err = wrapper.SendAlert(client, wsQuery.ChatId, symbol, lastPrice)
					if err != nil {
						fmt.Println("sendAlert", err)
					}
					alert.SetLastSignal(time.Now().In(time.UTC))
				}
			}
			// fmt.Printf("Channel: %s, Price: %f\n", ticker.Stream, lastPrice)
		}
	}
}

func checkPriceHu(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	wg.Add(1)
	defer func() {
		fmt.Println("checkPriceHu is DONE")
		wg.Done()
	}()
	// var close bool
	close, cancel := context.WithCancel(shutdownSrv)
	// userChannel := userChannels[wsQuery.UserId]
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!
	userChannel, _ := uc.GetUserChannels(wsQuery.UserId)
	alerts := session.AlertsByID(wsQuery.UserId)
	if len(alerts) > 0 {
		dbModels.SortByHEX(alerts)
	}

	go func() {
		wg.Add(1)
		defer func() {
			wg.Done()
			fmt.Println("Huobi SUB/UNSUB")
		}()
		for {
			select {
			case pair := <-userChannel.SubscribeCh(wsQuery.Market):
				fmt.Println("HUOBISUB")
				alerts = session.AlertsByID(wsQuery.UserId)
				dbModels.SortByHEX(alerts)
				err := wrapper.SubscribeHu(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = wrapper.SendAlertConfirmed(client, wsQuery.ChatId)
				if err != nil {
					fmt.Printf("checkPriceHu: %v", err)
				}
			case pair := <-userChannel.UnsubscribeCh(wsQuery.Market):
				fmt.Println("HUOBIUNSUB")
				alerts = session.AlertsByID(wsQuery.UserId)
				if len(alerts) > 0 {
					dbModels.SortByHEX(alerts)
				}
				err := wrapper.UnsubHu(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Pair.Size: ", pair.Size)
				if pair.Size <= 0 {
					fmt.Println("Huo pairs empty")
					userChannel.Shutdown(wsQuery.Market)
					userChannel.Cancel(wsQuery.Market)
					return
				}
			case <-close.Done():
				return
			}
		}
	}()

	var buf bytes.Buffer
	_, data, err := conn.ReadMessage()
	if err != nil {
		cancel()
		handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
		return
	}
	buf.Write(data)
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		cancel()
		handleCheckPriceErr(wsQuery)
		return
	}
	defer zr.Close()

	ticker := cryptoMarkets.NewTickerHuobi()
	for {
		zr.Multistream(false)
		msgType, dataRaw, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("ReadMSG %d, %v\nData: %s", msgType, err, string(dataRaw))
			cancel()
			handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
			return
		}

		buf.Write(dataRaw)
		data, err = io.ReadAll(zr)
		if err != nil {
			cancel()
			handleCheckPriceErr(wsQuery)
			return
		}

		err = wrapper.CheckPingHuobi(conn, data)
		if err != nil {
			err = json.Unmarshal(data, ticker)
			if err != nil {
				cancel()
				fmt.Println("unmarshal", err)
				handleCheckPriceErr(wsQuery)
				return
			}

			symbol := ticker.GetSymbol()
			if symbol != "" {
				alert, _ := dbModels.NewAlert(dbModels.WithMarket(wrapper.Huobi), dbModels.WithPair(symbol))
				i, err := alert.Find(alerts)
				if err != nil {
					alerts = session.AlertsByID(wsQuery.UserId)
					if len(alerts) > 0 {
						dbModels.SortByHEX(alerts)
					}
					fmt.Println("huobi", err)
				}
				if len(alerts) == 0 {
					cancel()
					return
				}
				alert = &alerts[i]
				lastPrice := ticker.GetLastPrice()

				if lastPrice <= alert.TargetPrice+(alert.TargetPrice*0.01) && lastPrice >= alert.TargetPrice-(alert.TargetPrice*0.01) {
					// send notification
					if alert.LastSignal.IsZero() || time.Now().UTC().UnixMilli() >= (alert.LastSignal.Add(15*time.Minute)).UnixMilli() {
						err = wrapper.SendAlert(client, wsQuery.ChatId, symbol, lastPrice)
						if err != nil {
							fmt.Println(err)
						}
						alert.SetLastSignal(time.Now().In(time.UTC))
					}
				}
				// fmt.Printf("Channel: %s, Price: %f\n", ticker.Channel, lastPrice)
			}
		}

		err = zr.Reset(&buf)
		if err != nil {
			cancel()
			fmt.Println(err)
			handleCheckPriceErr(wsQuery)
			return
		}
	}
}

func handleCheckPriceErr(wsQuery *other.WSQuery) {
	// userChannel := userChannels[wsQuery.UserId]
	// !!!!!!!!!!!!!!!!!!!!!!!!
	userChannel, _ := uc.GetUserChannels(wsQuery.UserId)
	userChannel.Cancel(wsQuery.Market)
}

func handleConnReadErr(
	dialer *websocket.Dialer, client *http.Client,
	wsQuery *other.WSQuery, coll *mongo.Collection,
	shutdownSrv context.Context,
) {
	// userChannel := userChannels[wsQuery.UserId]
	// !!!!!!!!!!!!!!!!!!!!!!
	userChannel, _ := uc.GetUserChannels(wsQuery.UserId)
	select {
	case <-userChannel.ShutdownCh(wsQuery.Market):
		// delete(userChannels, wsQuery.UserId)
		// session.DeleteAlerts(wsQuery.UserId, len(session.AlertsByID(wsQuery.ChatId)))
		session.SetMarketByID(wsQuery.UserId, wsQuery.Market, false)
	default:
		// Close go-routine to prevent leakage
		userChannel.Cancel(wsQuery.Market)
		// reconnect in case of 24h limit or other error
		_, alerts := session.AlertsByMarket(wsQuery.UserId, wsQuery.Market)
		db.UpdateAlertsSqc(coll, wsQuery.UserId, alerts, false, shutdownSrv)
		session.SetMarketByID(wsQuery.UserId, wsQuery.Market, false)
		_, err := connectToWS(coll, dialer, client, wsQuery, shutdownSrv)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			// delete(userChannels, wsQuery.UserId)
		}
	}
}

func disconnectAlert(coll *mongo.Collection, ctx context.Context, callback *models.CallbackQuery) error {
	wg.Add(1)
	defer func() {
		wg.Done()
		fmt.Println("DISCONNECTED")
	}()

	userID := callback.From.Id

	userChannel, ok := uc.GetUserChannels(userID)
	var index int
	// split callback data to get pair and market
	pair, market := wrapper.SplitCallbackData(callback.Data)
	fmt.Println(pair, market)

	// alerts := session.AlertsByID(userID)
	alerts, alertsMarket := session.AlertsByMarket(userID, market)
	if len(alerts) == 0 {
		return errors.New("disconnectAlert: alerts are empty")
	}
	lenghtAM := len(alertsMarket)
	lenghtA := len(alerts)

	// using Callback from TG to get pair for deletion
	// remove pair
	index, err := findDisconnectAlert(market, pair, alerts)
	if err != nil {
		return err
	}

	switch {
	case index >= 0:
		err := db.RemoveAlert(coll, userID, pair, ctx)
		if err != nil {
			return err
		}

		if ok && session.MarketExist(userID, market) {
			// fmt.Println("START_UNSUB")
			fmt.Println("disconnectAlert", alerts)
			// fmt.Println("disconnectAlert", alerts[index])
			err = session.DeleteAlert(userID, index)
			if err != nil {
				return err
			}
			userChannel.UnsubscribeSignal(market, other.PairSignal{Pair: pair, Size: lenghtAM - 1})
			// fmt.Println("FINISH_UNSUB")
		}
		// fmt.Println("FINISH", session.AlertsByID(userID))
		return nil
	case index == -1:
		err = db.DeleteAlerts(coll, userID, alerts, ctx)
		if err != nil {
			return err
		}
		if ok {
			session.DeleteAlerts(userID, len(alerts))
			session.SetMarketsByID(userID, false)
			for _, v := range alerts {
				userChannel.UnsubscribeSignal(v.Market, other.PairSignal{Pair: v.Pair, Size: lenghtA - 1})
			}
		}
		fmt.Println("AFTER UNSUB")
		return nil
	default:
		return fmt.Errorf("there's no pair with the market %s for this userID %d", market, userID)
	}
}

func updateFromTG(client *http.Client, dialer *websocket.Dialer, coll *mongo.Collection, shutdownSrv context.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(string(data))
		result, err := models.NewUpdate()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(data, result)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(result)
		msg := result.Msg
		switch {
		case len(msg.Entities) > 0:
			// handle commands
			command := msg.Text

			switch wrapper.CommandRouter(command, regexps) {
			case "start":
				err := wrapper.StartRouter(result, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				// alerts := make([]dbModels.Alert, 0)
				user, err := dbModels.NewMongoUser(
					dbModels.WithUserID(result.FromUser().Id),
					dbModels.WithChatID(result.FromChat().Id),
				)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				user.Alerts = make([]dbModels.Alert, 0)
				err = db.InsertNewUser(coll, user, shutdownSrv)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			case "help":
				err = wrapper.HelpRouter(result, client)
				if err != nil {
					fmt.Println(err)
					return
				}
			case "alert":
				wsQuery, err := wrapper.AlertRouter(command, regexps, result, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				err = alertHandler(dialer, client, wsQuery, shutdownSrv, coll)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			case "disconnect":
				pairs, err := db.GetAlerts(coll, result.FromUser().Id, shutdownSrv)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				err = wrapper.DisconnectRouter(result, pairs, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			case "price":
				err = wrapper.PriceRouter(client, result, command, regexps)
				if err != nil {
					fmt.Println(err)
					return
				}
			default:
				return
			}

		case result.GetCallbackData() != "":
			// handle callbacks
			fmt.Println(result.GetCallbackData())
			switch wrapper.CallbackHandler(client, result.GetCallbackData(), regexps) {
			case "disconnect":
				// fmt.Println("Callback")
				callback, err := wrapper.DisconnectCallback(result, regexps, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				// fmt.Println(callback)
				err = disconnectAlert(coll, shutdownSrv, callback)
				if err != nil {
					fmt.Println("disconnectAlert")
					fmt.Fprintln(os.Stderr, err)
					return
				}
				fmt.Println("AFTER DISCONNECT")

				answer, err := models.NewCallbackAnswer(
					models.WithAnswerID(callback.Id),
					models.WithAnswerText("Alert(s) disabled"),
					models.WithAnswerCacheTime(1),
				)
				if err != nil {
					fmt.Println("CreateNewAnswer")
					fmt.Fprintln(os.Stderr, err)
					return
				}
				err = wrapper.SendCallbackAnswer(client, answer)
				if err != nil {
					fmt.Println("SendCallbackAnswer", err)
				}
				fmt.Println(result.FromUser().Id)
				pairs, err := db.GetAlerts(coll, callback.From.Id, shutdownSrv)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				// fmt.Println(pairs)
				// us := userChannels[callback.From.Id]
				err = wrapper.EditMarkup(client, callback, pairs)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			default:
				return
			}
		}
	}
}

// For testing only
/*
func createUpdate(dialer *websocket.Dialer, client *http.Client, shutdownSrv context.Context, coll *mongo.Collection) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		offset := 1
		id := -1
		for {
			postBody, err := json.Marshal(map[string]int{
				"offset":  -offset,
				"timeout": 0,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			bodyReq := bytes.NewBuffer(postBody)
			resp, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", os.Getenv("TG")), "application/json", bodyReq)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			update := &models.TGUpdate{}
			err = json.Unmarshal(body, update)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			go func(update *models.TGUpdate) {
				wg.Add(1)
				defer wg.Done()
				if len(update.Result) > 0 {
					result := update.Result[len(update.Result)-1]

					if result.Id == id {
						return
					}

					id = result.Id

					msg := result.Msg

					switch {
					case len(msg.Entities) > 0:
						// handle commands
						// fmt.Println(msg.Text)

						// us := userChannels[msg.From.Id]
						command := msg.Text

						switch wrapper.CommandRouter(command, regexps) {
						case "start":
							err := wrapper.StartRouter(&result, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							user, err := dbModels.NewMongoUser(
								dbModels.WithUserID(result.FromUser().Id),
								dbModels.WithChatID(result.FromChat().Id),
							)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							user.Alerts = make([]dbModels.Alert, 0)
							err = db.InsertNewUser(coll, user, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						case "help":
							err = wrapper.HelpRouter(&result, client)
							if err != nil {
								fmt.Println(err)
								return
							}
						case "alert":
							wsQuery, err := wrapper.AlertRouter(command, regexps, &result, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							err = alertHandler(dialer, client, wsQuery, shutdownSrv, coll)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						case "disconnect":
							pairs, err := db.GetAlerts(coll, result.FromUser().Id, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							err = wrapper.DisconnectRouter(&result, pairs, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						case "price":
							err = wrapper.PriceRouter(client, &result, command, regexps)
							if err != nil {
								fmt.Println(err)
								return
							}
						default:
							return
						}

					case result.GetCallbackData() != "":
						// handle callbacks
						fmt.Println(result.GetCallbackData())
						switch wrapper.CallbackHandler(client, result.GetCallbackData(), regexps) {
						case "disconnect":
							// fmt.Println("Callback")
							callback, err := wrapper.DisconnectCallback(&result, regexps, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							// fmt.Println(callback)
							err = disconnectAlert(coll, shutdownSrv, callback)
							if err != nil {
								fmt.Println("disconnectAlert")
								fmt.Fprintln(os.Stderr, err)
								return
							}
							fmt.Println("AFTER DISCONNECT")

							answer, err := models.NewCallbackAnswer(
								models.WithAnswerID(callback.Id),
								models.WithAnswerText("Alert(s) disabled"),
								models.WithAnswerCacheTime(1),
							)
							if err != nil {
								fmt.Println("CreateNewAnswer")
								fmt.Fprintln(os.Stderr, err)
								return
							}
							err = wrapper.SendCallbackAnswer(client, answer)
							if err != nil {
								fmt.Println("SendCallbackAnswer", err)
							}
							fmt.Println(result.FromUser().Id)
							pairs, err := db.GetAlerts(coll, callback.From.Id, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							// fmt.Println(pairs)
							// us := userChannels[callback.From.Id]
							err = wrapper.EditMarkup(client, callback, pairs)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						default:
							return
						}
					}
				}
			}(update)
		}
	}
}
*/
