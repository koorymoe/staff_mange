package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ═══ 🎙️ تحويل فويس ايريس لنص بـGemini (دازاي 10-23) ═══
// Whisper الصغير بالسيرفر بطيء ويغلط بالعراقي. فويس المالك والمدير ويا ايريس يروح
// لـGemini (المفتاح الموجود)، وإذا فشل يرجع لـWhisper. فويسات الموظفين ما تمر من هنا.

const geminiSTTPrompt = `حوّل هذا التسجيل الصوتي لنص مكتوب حرفياً، باللهجة العراقية مثل ما انحچى (لا تترجمه للفصحى ولا تصلّح الكلام).
أسماء الأشخاص والأماكن والأرقام اكتبها مثل ما تنسمع. رجّع النص بس — بلا مقدمة ولا شرح ولا علامات تنصيص. إذا ماكو كلام واضح رجّع فاضي.`

type GeminiTranscriber struct {
	key  string
	url  string
	http *http.Client
}

// NewGeminiTranscriber nil إذا ماكو مفتاح.
func NewGeminiTranscriber(key string) *GeminiTranscriber {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	return &GeminiTranscriber{key: key, http: &http.Client{Timeout: 45 * time.Second},
		url: "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent"}
}

func (g *GeminiTranscriber) Transcribe(ctx context.Context, audio []byte, contentType string) (string, error) {
	if g == nil {
		return "", errors.New("gemini off")
	}
	mime := contentType
	if i := strings.Index(mime, ";"); i > 0 {
		mime = mime[:i]
	}
	body, _ := json.Marshal(map[string]any{
		"contents": []any{map[string]any{"parts": []any{
			map[string]any{"inline_data": map[string]any{"mime_type": mime, "data": base64.StdEncoding.EncodeToString(audio)}},
			map[string]any{"text": geminiSTTPrompt},
		}}},
		"generationConfig": map[string]any{"temperature": 0},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.key)
	resp, err := g.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini %d", resp.StatusCode)
	}
	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Candidates) == 0 {
		return "", errors.New("gemini: رد فاضي")
	}
	var b strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		b.WriteString(p.Text)
	}
	text := strings.Trim(strings.TrimSpace(b.String()), "\"«»")
	if text == "" {
		return "", errors.New("gemini: ما فهم الصوت")
	}
	return text, nil
}
