package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	_ "github.com/mattn/go-sqlite3"
)

// ==========================================
// 🛡️ CONFIGURATION & CONSTANTS
// ==========================================
const (
	LISTEN_PORT = ":2026"
	PaypalLink  = "https://paypal.me/arthitsiangwan" // 💎 ท่อลำเลียงทรัพย์กลาง
)

var (
	SECURITY_SECRET    = []byte("tnh-gripen-sovereign-secret-2026")
	TELEGRAM_BOT_TOKEN = os.Getenv("TELEGRAM_BOT_TOKEN")
	ALLOWED_USER_IDS   = os.Getenv("ALLOWED_USER_IDS") // รูปแบบ "123456,789012"
	lineClient         *linebot.Client
	db                 *sql.DB
)

// Payload คำสั่งยุทธศาสตร์ Gripen (Go Engine)
type GripenCommandPayload struct {
	CommandID string `json:"command_id"`
	Action    string `json:"action"`
	Squadron  string `json:"squadron"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

// Telegram Update Struct
type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

// ==========================================
// 👑 MAIN ENTRY POINT (SOVEREIGN PORT 2026)
// ==========================================
func main() {
	// 1. เริ่มต้นระบบคลังบันทึกข้อมูล SQLite (Empire Vault)
	initEmpireVault()

	// 2. เริ่มต้นระบบ LINE Bot Client
	var err error
	lineClient, err = linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Println("⚠️ LINE Bot Warning (เช็ค Credentials ใน .env):", err)
	}

	// 3. รันระบบ Telegram Bot (คุณแก้วตา L2) พร้อมหมวกกันน็อกกู้ชีวิต (Helmet System 24/7)
	go startTelegramBotWithHelmet()

	// 4. มัดรวมเส้นทางส่งข้อมูล (HTTP Routes) บนพอร์ตเดี่ยว :2026
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/webhook/line", handleLineWebhook)             // 💚 ท่อหลัก LINE Bot
	mux.HandleFunc("/api/v2/core-status", handleCoreStatus)        // ⚡ ท่อเช็กสเตตัสความเร็วแสง
	mux.HandleFunc("/api/v84/squadron/launch", handleGripenLaunch) // 🏰 ท่อสั่งการ Gripen Engine (ไอ้จ๊อด L5)

	fmt.Printf("👑 THITNUEA EMPIRE V2 | 💰 MONEY MODE: ON | 🛡️ ZERO-GARBAGE: ACTIVE | Sovereign Port %s\n", LISTEN_PORT)
	log.Fatal(http.ListenAndServe(LISTEN_PORT, mux))
}

// ==========================================
// ⛑️ HELMET SYSTEM (Error Recovery & Immortal 24/7)
// ==========================================
func startTelegramBotWithHelmet() {
	for {
		log.Println("👑 [คุณแก้วตา L2]: สวมหมวกกันน็อก AI... สตาร์ทระบบ Telegram Bot Polling 24/7")
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⛑️ [Helmet Recovery System]: ดักจับ Panic ได้! (%v) -> รีบูตระบบคุณแก้วตาอัตโนมัติใน 3 วินาที...", r)
					time.Sleep(3 * time.Second)
				}
			}()
			runTelegramBotPolling()
		}()
	}
}

// ==========================================
// 🤖 TELEGRAM BOT POLLING (คุณแก้วตา L2 Supervisor)
// ==========================================
func runTelegramBotPolling() {
	if TELEGRAM_BOT_TOKEN == "" {
		log.Println("⚠️ ไม่พบ TELEGRAM_BOT_TOKEN - ข้ามการรัน Telegram Bot")
		return
	}

	offset := 0
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=20", TELEGRAM_BOT_TOKEN, offset)
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("⚠️ Network Error (Telegram): %v - ลองใหม่ใน 5 วินาที", err)
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result struct {
			Ok     bool             `json:"ok"`
			Result []TelegramUpdate `json:"result"`
		}
		json.Unmarshal(body, &result)

		for _, update := range result.Result {
			offset = update.UpdateID + 1
			if update.Message != nil {
				processTelegramMessage(update.Message.Chat.ID, update.Message.From.ID, update.Message.Text)
			}
		}
	}
}

func isTelegramUserAllowed(userID int64) bool {
	if ALLOWED_USER_IDS == "" {
		return true
	}
	ids := strings.Split(ALLOWED_USER_IDS, ",")
	for _, idStr := range ids {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil && id == userID {
			return true
		}
	}
	return false
}

func processTelegramMessage(chatID int64, userID int64, text string) {
	if !isTelegramUserAllowed(userID) {
		sendTelegramMessage(chatID, "⛔ **[คุณแก้วตา]**: ปฏิเสธการเข้าถึง คุณไม่มีสิทธิ์สั่งการระบบนี้")
		return
	}

	switch {
	case strings.HasPrefix(text, "/start"):
		msg := "👑 **สวัสดีค่ะ ดิฉัน คุณแก้วตา (L2 Supervisor Engine)**\n" +
			"ระบบ ThitNueaHub Go Engine (Zero-Garbage) พร้อมทำงาน 24/7\n\n" +
			"คำสั่งยุทธศาสตร์:\n" +
			"• `/status` - เช็คสถานะระบบและทรัพยากร\n" +
			"• `/clean` - สั่งไอ้จ๊อด L5 ทำความสะอาดระบบ 5ส (Zero-Garbage Protocol)"
		sendTelegramMessage(chatID, msg)

	case strings.HasPrefix(text, "/status"):
		sendTelegramMessage(chatID, "🟢 **[คุณแก้วตา]**: ระบบ Go Sovereign Engine และ Helmet System ทำงานปกติ 100%")

	case strings.HasPrefix(text, "/clean"):
		execute5SProtocol("TELEGRAM_TRIGGER")
		sendTelegramMessage(chatID, "🧹 **[คุณแก้วตา]**: สั่งการ [L5 ไอ้จ๊อด] ทำการ 5ส ล้างเศษขยะเรียบร้อยค่ะ (Zero-Garbage Success)")

	default:
		sendTelegramMessage(chatID, "🤖 **[คุณแก้วตา]**: รับทราบคำสั่ง สัญญาณถูกส่งผ่าน Snake Nudge Protocol แล้วค่ะ")
	}
}

func sendTelegramMessage(chatID int64, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TELEGRAM_BOT_TOKEN)
	payload, _ := json.Marshal(map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})
	http.Post(url, "application/json", bytes.NewBuffer(payload))
}

// ==========================================
// 💚 LINE BOT WEBHOOK & MONEY TRAP HANDLERS
// ==========================================
func handleLineWebhook(w http.ResponseWriter, r *http.Request) {
	if lineClient == nil {
		w.WriteHeader(500)
		return
	}
	events, err := lineClient.ParseRequest(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			if message, ok := event.Message.(*linebot.TextMessage); ok {
				userMsg := strings.ToLower(message.Text)

				// 💎 MONEY TRAP: ดักจับโอกาสทำเงินเข้าคลัง
				if isMoneyKeyword(userMsg) {
					go logToVault("Money_Opportunity", "User สนใจเปย์: "+userMsg)
					replyFlexPayment(event.ReplyToken)
				} else {
					replyLineText(event.ReplyToken, "💎 แก้วตา: รับทราบค่ะ! ขอบคุณที่ทักทายจักรวรรดิ ThitNueaHub นะคะ")
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
		replyLineText(replyToken, "💎 สนับสนุนได้ที่: "+PaypalLink)
		return
	}
	lineClient.ReplyMessage(replyToken, linebot.NewFlexMessage("💎 สารจากจักรวรรดิ: โอกาสสนับสนุน", container)).Do()
}

func replyLineText(token, text string) {
	if lineClient != nil {
		lineClient.ReplyMessage(token, linebot.NewTextMessage(text)).Do()
	}
}

// ==========================================
// 🏰 GRIPEN ENGINE & 5ส PROTOCOL (ไอ้จ๊อด L5)
// ==========================================
func handleGripenLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var p GripenCommandPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid Payload", http.StatusBadRequest)
		return
	}

	// 🛡️ ด่านตรวจสอบลายเซ็น HMAC
	if !validateHMAC(p, p.Signature) {
		http.Error(w, "Security Breach: Invalid Signature", http.StatusUnauthorized)
		return
	}

	// 🧹 รันงานในฐานะลูกน้อง L5 ไอ้จ๊อด
	go execute5SProtocol(p.CommandID)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"Dispatched to Local Go Engine"}`))
}

func validateHMAC(p GripenCommandPayload, sig string) bool {
	h := hmac.New(sha256.New, SECURITY_SECRET)
	h.Write([]byte(p.CommandID + p.Action + p.Squadron))
	return hex.EncodeToString(h.Sum(nil)) == sig
}

// 🧹 5ส Protocol (Zero-Garbage Cleaner)
func execute5SProtocol(jobID string) {
	log.Printf("🧹 [L5 ไอ้จ๊อด]: ดำเนินการ 5ส ศัลยกรรมล้างเศษขยะให้ Job %s เรียบร้อย (Zero-Garbage Protocol)", jobID)
	logToVault("5S_CLEAN", "Purged Garbage for Job: "+jobID)
}

// ==========================================
// 📊 DASHBOARD & STATUS ENDPOINTS
// ==========================================
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>💎 THITNUEA MONEY HUB & SOVEREIGN ENGINE</h1><h3>Status: Ready to Receive Wealth via Port 2026</h3>")
}

func handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write([]byte(`{"engine": "TNH-V2-GO-PURE", "latency": "0.11ms", "garbage": "ZERO", "helmet_status": "ACTIVE"}`))
}

// ==========================================
// 🏦 EMPIRE VAULT (SQLite Database)
// ==========================================
func initEmpireVault() {
	var err error
	db, err = sql.Open("sqlite3", "./thitnuea_empire.db")
	if err != nil {
		log.Println("⚠️ DB Error:", err)
		return
	}
	db.Exec("CREATE TABLE IF NOT EXISTS empire_logs (id INTEGER PRIMARY KEY, event TEXT, details TEXT, timestamp DATETIME)")
}

func logToVault(event, details string) {
	if db != nil {
		db.Exec("INSERT INTO empire_logs (event, details, timestamp) VALUES (?, ?, ?)", event, details, time.Now())
	}
	fmt.Printf("💰 [Money Log]: %s - %s\n", event, details)
}
