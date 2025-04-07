package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func main() {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	bot, err := messaging_api.NewMessagingApiAPI(
		os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Health Check
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "✅ LineBot webhook is alive")
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[Webhook] /callback triggered")

		cb, err := webhook.ParseRequest(channelSecret, req)
		if err != nil {
			log.Printf("Failed to parse webhook: %+v", err)
			if errors.Is(err, webhook.ErrInvalidSignature) {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range cb.Events {
			switch e := event.(type) {
			case webhook.MessageEvent:
				switch message := e.Message.(type) {
				case webhook.TextMessageContent:
					userId := e.Source.UserID
					if source, ok := e.Source.(*webhook.UserSource); ok {
	                                    log.Printf("UserID: %s", source.UserId)
                                        }
                                        
					replyText := message.Text
					if message.Text == "/test" {
						replyText = "✅ 推播測試成功！"
					}

					_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							messaging_api.TextMessage{Text: replyText},
						},
					})
					if err != nil {
						log.Println("Reply error:", err)
					}
				}
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Println("🚀 Listening on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
