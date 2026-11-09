// Package telegram عميل صغير لـ Telegram Bot API — بوت بلاغات الموظفين (دازاي 11-05).
//
// ⚠️ الرمز (TELEGRAM_BOT_TOKEN) ما ينطبع بأي سجل ولا يطلع بأي رسالة خطأ:
// رابط الـ API بيه الرمز نفسه، فالأخطاء ترجع بلا الرابط.
package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	token string
	base  string // https://api.telegram.org (يتبدّل بالاختبار)
	http  *http.Client
	// Username اسم البوت (بلا @) — من getMe أول مرة.
	Username string
}

// New بلا رمز = بوت مطفي (Enabled=false) والنظام يشتغل عادي.
func New(token, base string) *Client {
	if base == "" {
		base = "https://api.telegram.org"
	}
	return &Client{token: strings.TrimSpace(token), base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 20 * time.Second}}
}

func (c *Client) Enabled() bool { return c != nil && c.token != "" }

var errAPI = errors.New("telegram api")

// call ينادي دالة من الـ API. الخطأ ما بيه الرابط (بيه الرمز).
func (c *Client) call(method string, payload any, out any) error {
	if !c.Enabled() {
		return errors.New("telegram: البوت مطفي (ماكو رمز)")
	}
	body, _ := json.Marshal(payload)
	resp, err := c.http.Post(c.base+"/bot"+c.token+"/"+method, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: %s: اتصال فشل", errAPI, method)
	}
	defer resp.Body.Close()
	var env struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		Description string          `json:"description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&env); err != nil || !env.OK {
		return fmt.Errorf("%w: %s: %s", errAPI, method, env.Description)
	}
	if out != nil {
		return json.Unmarshal(env.Result, out)
	}
	return nil
}

// GetMe يجيب اسم البوت (للرابط t.me/<اسم>?start=...).
func (c *Client) GetMe() (string, error) {
	var me struct {
		Username string `json:"username"`
	}
	if err := c.call("getMe", map[string]any{}, &me); err != nil {
		return "", err
	}
	c.Username = me.Username
	return me.Username, nil
}

// SendMessage رسالة نص عادي (بلا تنسيق — حتى نص الموظف ما ينكسر بالرموز).
func (c *Client) SendMessage(chatID int64, text string) error {
	return c.call("sendMessage", map[string]any{"chat_id": chatID, "text": text, "disable_web_page_preview": true}, nil)
}

// SendPhoto يدز صورة (بايتات من مخزن الملفات) ويا تعليق قصير.
func (c *Client) SendPhoto(chatID int64, data []byte, caption string) error {
	if !c.Enabled() {
		return errors.New("telegram: البوت مطفي (ماكو رمز)")
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	if caption != "" {
		_ = mw.WriteField("caption", caption)
	}
	fw, err := mw.CreateFormFile("photo", "report.jpg")
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
		return err
	}
	_ = mw.Close()
	resp, err := c.http.Post(c.base+"/bot"+c.token+"/sendPhoto", mw.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("%w: sendPhoto: اتصال فشل", errAPI)
	}
	defer resp.Body.Close()
	var env struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&env); err != nil || !env.OK {
		return fmt.Errorf("%w: sendPhoto: %s", errAPI, env.Description)
	}
	return nil
}

// SetWebhook يسجّل رابط السيرفر ويا رمز سري يرجع بترويسة كل تحديث.
func (c *Client) SetWebhook(url, secret string) error {
	return c.call("setWebhook", map[string]any{"url": url, "secret_token": secret, "allowed_updates": []string{"message"}, "drop_pending_updates": false}, nil)
}

// DownloadFile ينزّل ملف (صورة) بحد أقصى للحجم.
func (c *Client) DownloadFile(fileID string, max int64) ([]byte, error) {
	var f struct {
		FilePath string `json:"file_path"`
		FileSize int64  `json:"file_size"`
	}
	if err := c.call("getFile", map[string]any{"file_id": fileID}, &f); err != nil {
		return nil, err
	}
	if f.FileSize > max {
		return nil, errors.New("الصورة أكبر من الحد")
	}
	resp, err := c.http.Get(c.base + "/file/bot" + c.token + "/" + f.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: تنزيل الملف فشل", errAPI)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: تنزيل الملف %d", errAPI, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, errors.New("الصورة أكبر من الحد")
	}
	return data, nil
}

// ═══ أنواع التحديث (بس الي نحتاجه) ═══

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID    int64       `json:"message_id"`
	From         *User       `json:"from"`
	Chat         Chat        `json:"chat"`
	Text         string      `json:"text"`
	Caption      string      `json:"caption"`
	Photo        []PhotoSize `json:"photo"`
	Document     *Document   `json:"document"`
	MediaGroupID string      `json:"media_group_id"`
}

type User struct {
	ID       int64  `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // private | group | supergroup | channel
}

type PhotoSize struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type Document struct {
	FileID   string `json:"file_id"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}
