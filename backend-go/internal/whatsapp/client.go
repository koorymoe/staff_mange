// Package whatsapp بوابة WhatsApp Business Cloud API (Meta) — دازاي 10-23.
// الإرسال، والتحقق من توقيع الـwebhook، وقراءة الرسائل الواردة. بلا مفاتيح
// Meta كل إرسال يرجع ErrOff، ووضع التجربة يشتغل بلاها.
package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrOff = errors.New("الواتساب مو مربوط بعد — شغّل وضع التجربة")

const graphBase = "https://graph.facebook.com/v20.0"

type Client struct {
	token, phoneID, appSecret, verifyToken string
	http                                   *http.Client
	base                                   string
}

func New(token, phoneID, appSecret, verifyToken string) *Client {
	return &Client{token: token, phoneID: phoneID, appSecret: appSecret, verifyToken: verifyToken,
		http: &http.Client{Timeout: 20 * time.Second}, base: graphBase}
}

// On الإرسال الحقيقي مربوط؟
func (c *Client) On() bool { return c != nil && c.token != "" && c.phoneID != "" && c.appSecret != "" }

// VerifyToken يطابق hub.verify_token بأول ربط للـwebhook.
func (c *Client) VerifyToken(t string) bool {
	return c != nil && c.verifyToken != "" && hmac.Equal([]byte(t), []byte(c.verifyToken))
}

// ValidSignature X-Hub-Signature-256 = sha256=HMAC(body, appSecret).
func (c *Client) ValidSignature(body []byte, header string) bool {
	if c == nil || c.appSecret == "" || !strings.HasPrefix(header, "sha256=") {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}
	m := hmac.New(sha256.New, []byte(c.appSecret))
	m.Write(body)
	return hmac.Equal(got, m.Sum(nil))
}

// SendText رسالة نصية (تنقبل بس داخل ٢٤ ساعة من آخر رسالة للزبون — قاعدة Meta).
func (c *Client) SendText(ctx context.Context, to, text string) (string, error) {
	return c.send(ctx, map[string]any{"messaging_product": "whatsapp", "to": to, "type": "text",
		"text": map[string]any{"body": text, "preview_url": false}})
}

// SendTemplate أول رسالة تبدأها الشركة — قالب موافق عليه من Meta.
func (c *Client) SendTemplate(ctx context.Context, to, name, lang string, params []string) (string, error) {
	tpl := map[string]any{"name": name, "language": map[string]any{"code": lang}}
	if len(params) > 0 {
		ps := make([]map[string]any, 0, len(params))
		for _, p := range params {
			ps = append(ps, map[string]any{"type": "text", "text": p})
		}
		tpl["components"] = []map[string]any{{"type": "body", "parameters": ps}}
	}
	return c.send(ctx, map[string]any{"messaging_product": "whatsapp", "to": to, "type": "template", "template": tpl})
}

func (c *Client) send(ctx context.Context, body map[string]any) (string, error) {
	if !c.On() {
		return "", ErrOff
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/"+c.phoneID+"/messages", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("whatsapp %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var out struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	_ = json.Unmarshal(data, &out)
	if len(out.Messages) == 0 {
		return "", nil
	}
	return out.Messages[0].ID, nil
}

// Inbound رسالة واردة من الزبون.
type Inbound struct {
	MetaID string
	From   string // رقم الزبون بلا +
	Name   string
	Text   string
	At     time.Time
}

// ParseWebhook يطلّع الرسائل النصية من جسم الـwebhook (الباقي — صور وحالات — يتجاهل).
func ParseWebhook(body []byte) ([]Inbound, error) {
	var p struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Contacts []struct {
						Profile struct {
							Name string `json:"name"`
						} `json:"profile"`
						WaID string `json:"wa_id"`
					} `json:"contacts"`
					Messages []struct {
						ID        string `json:"id"`
						From      string `json:"from"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
						Button struct {
							Text string `json:"text"`
						} `json:"button"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Inbound
	for _, e := range p.Entry {
		for _, ch := range e.Changes {
			names := map[string]string{}
			for _, ct := range ch.Value.Contacts {
				names[ct.WaID] = ct.Profile.Name
			}
			for _, m := range ch.Value.Messages {
				text := m.Text.Body
				if m.Type == "button" {
					text = m.Button.Text
				}
				if strings.TrimSpace(text) == "" {
					text = "[رسالة " + m.Type + " — الوكيل يقرا النص بس]"
				}
				var at time.Time
				var sec int64
				if _, err := fmt.Sscan(m.Timestamp, &sec); err == nil {
					at = time.Unix(sec, 0)
				}
				out = append(out, Inbound{MetaID: m.ID, From: m.From, Name: names[m.From], Text: text, At: at})
			}
		}
	}
	return out, nil
}
