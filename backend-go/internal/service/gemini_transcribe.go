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
	return NewGeminiTranscriberModel(key, "")
}

// NewGeminiTranscriberModel بموديل محدد (GEMINI_STT_MODEL) — فاضي = gemini-flash-latest.
func NewGeminiTranscriberModel(key, model string) *GeminiTranscriber {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if model == "" {
		model = "gemini-flash-latest"
	}
	return &GeminiTranscriber{key: key, http: &http.Client{Timeout: 60 * time.Second},
		url: "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent"}
}

func (g *GeminiTranscriber) Transcribe(ctx context.Context, audio []byte, contentType string) (string, error) {
	if g == nil {
		return "", errors.New("gemini off")
	}
	mime := contentType
	if i := strings.Index(mime, ";"); i > 0 {
		mime = mime[:i]
	}
	if mime == "audio/mpeg" {
		mime = "audio/mp3"
	}
	// (دازاي 10-29) «يطول هواي وهو يحلل الصوت» — Flash يفكر قبل ما يكتب بشكل افتراضي،
	// والكتابة الحرفية ما تحتاج تفكير: نطفّيه، وإذا الموديل رفض نعيد بدونه.
	t, status, err := g.call(ctx, audio, mime, true)
	if status == http.StatusBadRequest {
		t, _, err = g.call(ctx, audio, mime, false)
	}
	return t, err
}

func (g *GeminiTranscriber) call(ctx context.Context, audio []byte, mime string, noThink bool) (string, int, error) {
	cfg := map[string]any{"temperature": 0}
	if noThink {
		cfg["thinkingConfig"] = map[string]any{"thinkingBudget": 0}
	}
	body, _ := json.Marshal(map[string]any{
		"contents": []any{map[string]any{"parts": []any{
			map[string]any{"inline_data": map[string]any{"mime_type": mime, "data": base64.StdEncoding.EncodeToString(audio)}},
			map[string]any{"text": geminiSTTPrompt + DefaultVoiceLexicon.PromptHint()},
		}}},
		"generationConfig": cfg,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.key)
	resp, err := g.http.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", resp.StatusCode, fmt.Errorf("gemini %d", resp.StatusCode)
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
		return "", resp.StatusCode, errors.New("gemini: رد فاضي")
	}
	var b strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		b.WriteString(p.Text)
	}
	text := strings.Trim(strings.TrimSpace(b.String()), "\"«»")
	if text == "" {
		return "", resp.StatusCode, errors.New("gemini: ما فهم الصوت")
	}
	return text, resp.StatusCode, nil
}
