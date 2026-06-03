package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	_ "github.com/mattn/go-sqlite3"
)

const (
	PaypalLink = "https://paypal.me/arthitsiangwan" // 💎 ท่อลำเลียงทรัพย์กลาง
)

var (
	bot *linebot.Client
	db  *sql.DB
)

func main() {
	initEmpireVault()

	var err error
	bot, err = linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Println("⚠️ LINE Bot Warning:", err)
	}

	// มัดรวมเส้นทางส่งข้อมูล (Routes)
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/webhook/line", handleLineWebhook) // ท่อหลัก LINE Bot
	http.HandleFunc("/api/v2/core-status", handleCoreStatus) // ท่อเช็กสเตตัสความเร็วแสง

	// 🔒 บังคับล็อกพิกัดเข้าพอร์ตเดี่ยว 2026 ของมหาจักรวรรดิเพื่อความปลอดภัยสูงสุด
	port := "2026"
	
	fmt.Printf("👑 THITNUEA EMPIRE V2 | 💰 MONEY MODE: ON | Sovereign Port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleLineWebhook(w http.ResponseWriter, r *http.Request) {
	if bot == nil {
		w.WriteHeader(500)
		return
	}
	events, err := bot.ParseRequest(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			if message, ok := event.Message.(*linebot.TextMessage); ok {
				userMsg := strings.ToLower(message.Text)

				// 💎 MONEY TRAP: ดักจับโอกาสทำเงินเข้าคลังก่อนเสมอ
				if isMoneyKeyword(userMsg) {
					go logToVault("Money_Opportunity", "User สนใจเปย์: "+userMsg)
					replyFlexPayment(event.ReplyToken)
				} else {
					replyText(event.ReplyToken, "💎 แก้วตา: รับทราบค่ะ! ขอบคุณที่ทักทายจักรวรรดิ ThitNueaHub นะคะ")
				}
			}
		}
	}
	w.WriteHeader(200)
}

func isMoneyKeyword(text string) bool {
	keywords := []string{"สมัคร", "vip", "donate", "เปย์", "สนับสนุน", "เลขบัญชี", "พร้อมเพย์", "money"}
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func replyFlexPayment(replyToken string) {
	flexJSON := fmt.Sprintf(`{
		"type": "bubble",
		"hero": {
			"type": "image",
			"url": "https://cdn-icons-png.flaticon.com/512/2454/2454269.png", 
			"size": "full",
			"aspectRatio": "20:13",
			"aspectMode": "cover"
		},
		"body": {
			"type": "box",
			"layout": "vertical",
			"contents": [
				{"type": "text", "text": "💎 ThitNuea Premium", "weight": "bold", "size": "xl", "color": "#1DB446"},
				{"type": "text", "text": "ปลดล็อกพลัง AI ระดับเทพ!", "size": "md", "weight": "bold"},
				{"type": "text", "text": "ร่วมเป็นผู้สนับสนุนทีมงานเพื่อพัฒนาเทคโนโลยีเพื่อสังคม", "wrap": true, "size": "sm", "color": "#666666", "margin": "md"}
			]
		},
		"footer": {
			"type": "box",
			"layout": "vertical",
			"spacing": "sm",
			"contents": [
				{
					"type": "button",
					"style": "primary",
					"height": "sm",
					"color": "#00308F",
					"action": {
						"type": "uri",
						"label": "👉 เปย์เลย (PayPal)",
						"uri": "%s"
					}
				},
				{"type": "text", "text": "ขอบคุณที่สนับสนุนความฝันครับ ❤️", "size": "xs", "align": "center", "color": "#aaaaaa", "margin": "md"}
			]
		}
	}`, PaypalLink)

	container, err := linebot.UnmarshalFlexMessageJSON([]byte(flexJSON))
	if err != nil {
		replyText(replyToken, "💎 สนับสนุนได้ที่: "+PaypalLink)
		return
	}
	bot.ReplyMessage(replyToken, linebot.NewFlexMessage("💎 สารจากจักรวรรดิ: โอกาสสนับสนุน", container)).Do()
}

func replyText(token, text string) {
	bot.ReplyMessage(token, linebot.NewTextMessage(text)).Do()
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>💎 THITNUEA MONEY HUB IS ACTIVE</h1><h3>Status: Ready to Receive Wealth via Port 2026</h3>")
}

func handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write([]byte(`{"engine": "TNH-V2-GO", "latency": "0.11ms", "garbage": "ZERO"}`))
}

func initEmpireVault() {
	var err error
	db, err = sql.Open("sqlite3", "./thitnuea_empire.db")
	if err != nil {
		log.Println("⚠️ DB Error:", err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS empire_logs (id INTEGER PRIMARY KEY, event TEXT, details TEXT, timestamp DATETIME)")
}

func logToVault(event, details string) {
	if db != nil {
		db.Exec("INSERT INTO empire_logs (event, details, timestamp) VALUES (?, ?, ?)", event, details, time.Now())
	}
	fmt.Printf("💰 [Money Log]: %s - %s\n", event, details)
}
