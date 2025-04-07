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
	// 正確讀取環境變數
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	bot, err := messaging_api.NewMessagingApiAPI(
		os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 健康檢查路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "✅ LineBot webhook is alive")
	})

	// LINE webhook callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		log.Println("📩 /callback called...")

		cb, err := webhook.ParseRequest(channelSecret, req)
		if err != nil {
			log.Printf("❌ Cannot parse request: %+v\n", err)
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
					// 嘗試從來源斷言為 UserSource，才能取得 UserId
					if source, ok := e.Source.(*webhook.UserSource); ok {
						log.Printf("🪪 使用者 UserID：%s", source.UserId)
					
						_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								messaging_api.TextMessage{
									Text: fmt.Sprintf("✅ 你的 User ID 是：%s\n你說了：%s", source.UserId, message.Text),
								},
							},
						})
						if err != nil {
							log.Println("❌ 回覆失敗:", err)
						}
					} else {
						log.Println("⚠️ 無法轉換為 UserSource（可能是群聊或聊天室）")
					}
					
				}
			}
		}
	})

	// 啟動伺服器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("🚀 Starting server at port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
