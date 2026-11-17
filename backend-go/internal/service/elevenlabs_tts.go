package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ═══ 🔊 كمال يحچي (دازاي 11-17) ═══
// «نسويله صوت… اريده يرد علي بكلام مفهوم وصحيح». رد كمال المكتوب يتحوّل لصوت عراقي من
// ElevenLabs (نفس مفتاح Scribe). النص ينظّف قبل (بلا إيموجي ولا تنسيق)، وكل نص ينحفظ صوته
// بالذاكرة حتى إعادة السماع ما تنحسب مرتين.
//
// المفتاح ELEVENLABS_API_KEY، والصوت ELEVENLABS_VOICE_ID (من مكتبة الأصوات — مثل صوت الاستقبال).
// بلا الاثنين = الميزة مطفية (404).

const (
	ttsMaxRunes  = 1500
	ttsCacheSize = 64
)

var ErrTTSOff = errors.New("صوت كمال مو مربوط — حط ELEVENLABS_API_KEY وELEVENLABS_VOICE_ID")

type ElevenLabsTTS struct {
	key, voice, model, base string
	http                    *http.Client
	mu                      sync.Mutex
	cache                   map[string][]byte
	order                   []string
}

func NewElevenLabsTTS(key, voice, model string) *ElevenLabsTTS {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(voice) == "" {
		return nil
	}
	if model == "" {
		model = "eleven_flash_v2_5" // أسرع موديل، ويدعم العربي
	}
	return &ElevenLabsTTS{key: key, voice: voice, model: model, base: "https://api.elevenlabs.io",
		http: &http.Client{Timeout: 40 * time.Second}, cache: map[string][]byte{}}
}

// SpeakText النص الي ينگال: نظيف ومقصوص.
func SpeakText(s string) string {
	s = speakable(s)
	if r := []rune(s); len(r) > ttsMaxRunes {
		s = string(r[:ttsMaxRunes])
	}
	return strings.TrimSpace(s)
}

// Speak يرجّع mp3.
func (t *ElevenLabsTTS) Speak(ctx context.Context, text string) ([]byte, error) {
	if t == nil {
		return nil, ErrTTSOff
	}
	text = SpeakText(text)
	if text == "" {
		return nil, errors.New("ماكو كلام ينگال")
	}
	sum := sha256.Sum256([]byte(t.voice + "|" + t.model + "|" + text))
	key := hex.EncodeToString(sum[:])
	t.mu.Lock()
	if b, ok := t.cache[key]; ok {
		t.mu.Unlock()
		return b, nil
	}
	t.mu.Unlock()

	body, _ := json.Marshal(map[string]any{"text": text, "model_id": t.model,
		"voice_settings": map[string]any{"stability": 0.5, "similarity_boost": 0.8}})
	url := fmt.Sprintf("%s/v1/text-to-speech/%s?output_format=mp3_44100_64", t.base, t.voice)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", t.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := t.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(b))
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("elevenlabs tts %d: %s", resp.StatusCode, msg)
	}
	t.mu.Lock()
	if _, ok := t.cache[key]; !ok {
		t.cache[key] = b
		t.order = append(t.order, key)
		if len(t.order) > ttsCacheSize {
			delete(t.cache, t.order[0])
			t.order = t.order[1:]
		}
	}
	t.mu.Unlock()
	return b, nil
}
