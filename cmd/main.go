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
	userChannels  map[int64]other.UserChannels
	regexps       map[string]*regexp.Regexp
	markets       map[string]bool
	sessionAlerts map[int64][]dbModels.Alert
	alertsCount   int = 0
)

func addSessionAlerts(id int64, alerts ...dbModels.Alert) {
	sessionAlerts[id] = append(sessionAlerts[id], alerts...)
	alertsCount += len(alerts)
}

func deleteSessionAlert(id int64, index int) {
	sessionAlerts[id] = append(sessionAlerts[id][:index], sessionAlerts[id][index+1:]...)
	alertsCount -= 1
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		// log.Fatal("$PORT must be set")
		port = "8080"
	}

	ctx, cancel := context.WithCancel(context.Background())
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
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}

	coll := db.GetUserCollection(mongoClient)

	regexps = compileRegexp()
	userChannels = make(map[int64]other.UserChannels)
	markets = make(map[string]bool)

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("TG")), updateFromTG)
	mux.HandleFunc("/update", createUpdate(dialer, client, ctx, coll))

	server := &http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		<-sigs
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		// mongoModels := make([]mongo.WriteModel, 0, alertsCount)
		// keys := make([]int64, len(sessionAlerts))
		// for k, v := range sessionAlerts {

		// }

		mongoClient.Disconnect(ctx)
		client.CloseIdleConnections()
		cancel()
	}()

	fmt.Println("Connected")

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
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
	userChannel, ok := userChannels[wsQuery.UserId]
	if ok {
		alert, err = dbModels.NewAlert(dbModels.WithPair(strings.ToLower(wsQuery.Pair)), dbModels.WithTargetPrice(wsQuery.Price), dbModels.WithMarket(wsQuery.Market))
		if err != nil {
			return fmt.Errorf("cannot create new alert: %s", err)
		}

		if markets[wsQuery.Market] {
			alert.Connected = true
			err = db.AddAlert(coll, wsQuery.UserId, alert, ctx)
			if err != nil {
				return fmt.Errorf("cannot add alert: %s", err)
			}
			addSessionAlerts(wsQuery.UserId, *alert)
			userChannel.SubscribeSignal(wsQuery.Market, other.PairSignal{Pair: wsQuery.Pair, Size: len(alerts)})
			return nil
		}
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

		userChannel.AssignCancel(make(map[string]context.CancelFunc))
		userChannel.AssignShutdown(make(map[string]chan int))
		userChannel.AssignSubscribe(make(map[string]chan other.PairSignal))
		userChannel.AssignUnsub(make(map[string]chan other.PairSignal))

		userChannel.SetShutdownCh(wsQuery.Market, make(chan int, 1))
		userChannel.SetSubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))
		userChannel.SetUnsubscriberCh(wsQuery.Market, make(chan other.PairSignal, 1))

		sessionAlerts[wsQuery.UserId] = make([]dbModels.Alert, 0, 5)
	}

	err = connectToWS(coll, &userChannel, dialer, client, wsQuery, ctx)
	if err != nil {
		return err
	}

	return nil
}

func connectToWS(
	coll *mongo.Collection, userChannel *other.UserChannels,
	dialer *websocket.Dialer, client *http.Client,
	wsQuery *other.WSQuery, shutdownSrv context.Context,
) error {
	pairs, oldAlerts, err := db.GetPairsByMarket(coll, wsQuery.UserId, wsQuery.Market, false, shutdownSrv)
	if err != nil {
		return fmt.Errorf("cannot get pairs by market: %s", err)
	}
	// Connect to Market
	conn, err := wrapper.TickerConnect(wsQuery.Market, pairs, dialer, client)
	if err != nil {
		return err
	}

	newAlerts := make([]dbModels.Alert, len(oldAlerts))

	for i, v := range oldAlerts {
		v.Connected = true
		newAlerts[i] = v
	}

	err = db.UpdateAlerts(coll, wsQuery.UserId, oldAlerts, newAlerts, shutdownSrv)
	if err != nil {
		conn.Close()
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	userChannel.SetCancel(wsQuery.Market, cancel)

	userChannels[wsQuery.UserId] = *userChannel

	// Close websocket connection
	go func(ctx context.Context, shutdownSrv context.Context, conn *websocket.Conn, userChannel *other.UserChannels, wsQuery *other.WSQuery) {
		defer fmt.Println("Websocket CLOSED")
		select {
		case <-shutdownSrv.Done():
			userChannel.Shutdown(wsQuery.Market)
			conn.Close()
		case <-ctx.Done():
			fmt.Println("Done")
			err := conn.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Connection closed: %s\n", err)
				return
			}
		}
	}(ctx, shutdownSrv, conn, userChannel, wsQuery)

	go checkPrice(conn, dialer, client, wsQuery, coll, shutdownSrv)

	addSessionAlerts(wsQuery.UserId, newAlerts...)
	markets[wsQuery.Market] = true
	return nil
}

func checkPrice(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	defer fmt.Println("CheckPrice is DONE")
	switch wsQuery.Market {
	case wrapper.Huobi:
		checkPriceHu(conn, dialer, client, wsQuery, coll, shutdownSrv)
	case wrapper.Binance:
		checkPriceBi(conn, dialer, client, wsQuery, coll, shutdownSrv)
	}
}

func checkPriceBi(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	var close bool
	userChannel := userChannels[wsQuery.UserId]

	go func() {
		for !close {
			select {
			case pair := <-userChannel.SubscribeCh(wsQuery.Market):
				err := wrapper.SubscribeBi(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
			case pair := <-userChannel.UnsubscribeCh(wsQuery.Market):
				err := wrapper.UnsubBi(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				if pair.Size == 0 {
					userChannel.Shutdown(wsQuery.Market)
					userChannel.Cancel(wsQuery.Market)
				}
			}
		}
	}()

	ticker := cryptoMarkets.NewTickerBi()
	defer fmt.Println("checkPriceBi is DONE")
	for {
		err := conn.ReadJSON(ticker)

		if err != nil {
			close = true
			fmt.Println("ReadJson: ", err)
			handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
			return
		}

		fmt.Printf("Channel: %s, Price: %s\n", ticker.Stream, ticker.GetLastPrice())
	}
}

func checkPriceHu(
	conn *websocket.Conn, dialer *websocket.Dialer,
	client *http.Client, wsQuery *other.WSQuery,
	coll *mongo.Collection, shutdownSrv context.Context,
) {
	var close bool
	userChannel := userChannels[wsQuery.UserId]

	go func() {
		for !close {
			select {
			case pair := <-userChannel.SubscribeCh(wsQuery.Market):
				fmt.Println("SubHuobi")
				err := wrapper.SubscribeHu(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
			case pair := <-userChannel.UnsubscribeCh(wsQuery.Market):
				fmt.Println("UnsubHuobi")
				err := wrapper.UnsubHu(conn, pair.Pair)
				if err != nil {
					fmt.Println(err)
					return
				}
				if pair.Size == 0 {
					userChannel.Shutdown(wsQuery.Market)
					userChannel.Cancel(wsQuery.Market)
				}
			}
		}
	}()

	defer fmt.Println("CheckPriceHu is DONE")
	var buf bytes.Buffer
	_, data, err := conn.ReadMessage()
	if err != nil {
		close = true
		handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
		return
	}
	buf.Write(data)
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		close = true
		handleCheckPriceErr(wsQuery)
		return
	}
	defer zr.Close()

	ticker := cryptoMarkets.NewTickerHuobi()
	for {
		zr.Multistream(false)
		msgType, data, err := conn.ReadMessage()
		fmt.Println(msgType)
		if err != nil {
			fmt.Printf("ReadMSG %d, %s", msgType, err)
			close = true
			handleConnReadErr(dialer, client, wsQuery, coll, shutdownSrv)
			return
		}
		buf.Write(data)
		data, err = io.ReadAll(zr)
		if err != nil {
			fmt.Println("ReadAll")
			close = true
			handleCheckPriceErr(wsQuery)
			return
		}
		wrapper.CheckPingHuobi(conn, data)
		err = json.Unmarshal(data, ticker)
		if err != nil {
			fmt.Println("Unmarshal")
			close = true
			handleCheckPriceErr(wsQuery)
			return
		}
		fmt.Printf("Channel: %s, Price: %f\n\n", ticker.Channel, ticker.GetLastPrice())
		err = zr.Reset(&buf)
		if err != nil {
			fmt.Println("Reset")
			close = true
			handleCheckPriceErr(wsQuery)
			return
		}
	}
}

func handleCheckPriceErr(wsQuery *other.WSQuery) {
	userChannel := userChannels[wsQuery.UserId]
	markets[wsQuery.Market] = false
	userChannel.Cancel(wsQuery.Market)
}

func handleConnReadErr(
	dialer *websocket.Dialer, client *http.Client,
	wsQuery *other.WSQuery, coll *mongo.Collection,
	shutdownSrv context.Context,
) {
	userChannel := userChannels[wsQuery.UserId]
	markets[wsQuery.Market] = false
	select {
	case <-userChannel.ShutdownCh(wsQuery.Market):
		delete(userChannels, wsQuery.UserId)
	default:
		// Close go-routine to prevent leakage
		userChannel.Cancel(wsQuery.Market)
		// reconnect in case of 24h limit or other error
		err := connectToWS(coll, &userChannel, dialer, client, wsQuery, shutdownSrv)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			// delete(userChannels, wsQuery.UserId)
		}
	}
}

func disconnectAlert(coll *mongo.Collection, ctx context.Context, callback *models.CallbackQuery) error {
	fmt.Println("Start")
	userID := callback.From.Id
	userChannel, ok := userChannels[userID]
	var index int
	// split callback data to get pair and market
	pair, market := wrapper.SplitCallbackData(callback.Data)

	alerts := sessionAlerts[userID]

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
		if ok && markets[market] {
			userChannel.UnsubscribeSignal(market, other.PairSignal{Pair: pair, Size: len(alerts) - 1})
		}
		deleteSessionAlert(userID, index)
		return nil
	case index == -1:
		err = db.DeleteAlerts(coll, userID, alerts, ctx)
		if err != nil {
			return err
		}
		if ok {
			for _, v := range alerts {
				userChannel.UnsubscribeSignal(market, other.PairSignal{Pair: v.Pair, Size: len(alerts) - 1})
			}
		}
		alertsCount -= len(alerts)
		delete(sessionAlerts, userID)
		return nil
	default:
		return fmt.Errorf("there's no pair with the market %s for this userID %d", market, userID)
	}
}

func updateFromTG(rw http.ResponseWriter, r *http.Request) {
	// data, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	// update := &wrapper.Update{}
	// err = json.Unmarshal(data, update)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	// fmt.Fprintln(os.Stdout, "Update:", update)

}

func createUpdate(dialer *websocket.Dialer, client *http.Client, shutdownSrv context.Context, coll *mongo.Collection) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		offset := 0
		id := -1
		for {
			postBody, err := json.Marshal(map[string]int{
				"offset":  -offset,
				"limit":   100,
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
						fmt.Println(msg.Text)

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
							pairs, err := db.GetPairs(coll, result.FromUser().Id, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							err = wrapper.DisconnectRouter(&result, pairs, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						default:
							return
						}

					case result.GetCallbackData() != "":
						// handle callbacks
						switch wrapper.CallbackHandler(client, result.GetCallbackData(), regexps) {
						case "disconnect":
							// fmt.Println("Callback")
							callback, err := wrapper.DisconnectCallback(&result, regexps, client)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							// fmt.Println("Callback_1")
							err = disconnectAlert(coll, shutdownSrv, callback)
							if err != nil {
								// fmt.Println("disconnectAlert")
								fmt.Fprintln(os.Stderr, err)
								return
							}

							answer, err := models.NewCallbackAnswer(
								models.WithAnswerID(callback.Id),
								models.WithAnswerText("Alert(s) disabled"),
								models.WithAnswerCacheTime(1),
							)
							if err != nil {
								// fmt.Println("CreateNewAnswer")
								fmt.Fprintln(os.Stderr, err)
								return
							}
							err = wrapper.SendCallbackAnswer(client, answer)
							if err != nil {
								fmt.Println("SendCallbackAnswer", err)
							}
							pairs, err := db.GetPairs(coll, result.FromUser().Id, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							fmt.Println(pairs)
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
