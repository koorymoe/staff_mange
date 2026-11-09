package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ 👂 أذن قابلة للتبديل (دازاي 10-29) ═══
// «شوفلي أداة خاصة… حتى نطلع أقوى نتيجة». ماكو مقارنة منشورة للعراقي بالذات، فنخلي
// أكثر من مزوّد، وزر «قارن» يدز نفس البصمة لكلهم، ودازاي يختار بالأرقام.
//   - gemini   (GEMINI_API_KEY، الموديل GEMINI_STT_MODEL)
//   - deepgram (DEEPGRAM_API_KEY، الموديل DEEPGRAM_MODEL، افتراضي nova-3)
//   - soniox   (SONIOX_API_KEY، الموديل SONIOX_MODEL)
//   - elevenlabs (ELEVENLABS_API_KEY، Scribe — الأول بمقارنة عربي مخلوط إنگليزي)
// الأساسي AI_IRIS_VOICE. كل بصمة توصل mp3 (ToMP3).

type STTProvider interface {
	Transcribe(ctx context.Context, audio []byte, contentType string) (string, error)
}

type NamedSTT struct {
	Name string
	P    STTProvider
	// Timeout مهلة هالمزوّد (صفر = ٢٥ ثانية) — حتى البطيء ما يوگف الباقي.
	Timeout time.Duration
}

func (n NamedSTT) run(ctx context.Context, audio []byte, ct string) (string, error) {
	d := n.Timeout
	if d <= 0 {
		d = 25 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	t, err := n.P.Transcribe(ctx, audio, ct)
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("تجاوز المهلة (%s)", d)
	}
	return t, err
}

// TranscribeChain يجرّب المزوّدات بالترتيب (الأساسي أول): إذا واحد فشل أو تأخر، الي بعده.
// (دازاي 10-29) «يگلي عيدها وبعد فترة يطلع» — فما نگول «عيدها» إلا إذا كلهم فشلوا.
func TranscribeChain(ctx context.Context, chain []NamedSTT, audio []byte, ct string) (text, name string, err error) {
	if len(chain) == 0 {
		return "", "", errors.New("ماكو مزوّد صوت")
	}
	var errs []string
	for _, p := range chain {
		t, e := p.run(ctx, audio, ct)
		if e == nil && strings.TrimSpace(t) != "" {
			return t, p.Name, nil
		}
		if e == nil {
			e = errors.New("نص فاضي")
		}
		errs = append(errs, p.Name+": "+e.Error())
		if ctx.Err() != nil {
			break
		}
	}
	return "", "", errors.New(strings.Join(errs, " | "))
}

type STTResult struct {
	Name    string  `json:"name"`
	Text    string  `json:"text"`
	Seconds float64 `json:"seconds"`
	Error   string  `json:"error,omitempty"`
}

// CompareSTT نفس البصمة لكل المزوّدات بالتوازي.
func CompareSTT(ctx context.Context, all []NamedSTT, audio []byte, ct string) []STTResult {
	out := make([]STTResult, len(all))
	var wg sync.WaitGroup
	for i, p := range all {
		wg.Add(1)
		go func(i int, p NamedSTT) {
			defer wg.Done()
			start := time.Now()
			if p.Timeout <= 0 || p.Timeout > 40*time.Second {
				p.Timeout = 40 * time.Second
			}
			t, err := p.run(ctx, audio, ct)
			r := STTResult{Name: p.Name, Text: DefaultVoiceLexicon.Apply(t), Seconds: time.Since(start).Round(100 * time.Millisecond).Seconds()}
			if err != nil {
				r.Error = err.Error()
			}
			out[i] = r
		}(i, p)
	}
	wg.Wait()
	return out
}

func baseMime(ct string) string {
	if i := strings.Index(ct, ";"); i > 0 {
		ct = ct[:i]
	}
	return ct
}

// ── Deepgram ──

type DeepgramSTT struct {
	key, model string
	http       *http.Client
}

func NewDeepgramSTT(key, model string) *DeepgramSTT {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if model == "" {
		model = "nova-3"
	}
	return &DeepgramSTT{key: key, model: model, http: &http.Client{Timeout: 45 * time.Second}}
}

func (d *DeepgramSTT) Transcribe(ctx context.Context, audio []byte, ct string) (string, error) {
	url := "https://api.deepgram.com/v1/listen?language=ar&smart_format=true&model=" + d.model
	if terms := DefaultVoiceLexicon.TopTerms(80); len(terms) > 0 && strings.HasPrefix(d.model, "nova-3") {
		for _, t := range terms {
			url += "&keyterm=" + urlEscape(t)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(audio))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Token "+d.key)
	req.Header.Set("Content-Type", baseMime(ct))
	resp, err := d.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("deepgram %d: %s", resp.StatusCode, short(raw))
	}
	var out struct {
		Results struct {
			Channels []struct {
				Alternatives []struct {
					Transcript string `json:"transcript"`
				} `json:"alternatives"`
			} `json:"channels"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Results.Channels) == 0 || len(out.Results.Channels[0].Alternatives) == 0 {
		return "", errors.New("deepgram: رد فاضي")
	}
	t := strings.TrimSpace(out.Results.Channels[0].Alternatives[0].Transcript)
	if t == "" {
		return "", errors.New("deepgram: ما فهم الصوت")
	}
	return t, nil
}

// ── Soniox (رفع ملف ← طلب تحويل ← انتظار ← النص) ──

type SonioxSTT struct {
	key, model string
	http       *http.Client
	base       string
}

func NewSonioxSTT(key, model string) *SonioxSTT {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if model == "" {
		model = "stt-async-v3"
	}
	return &SonioxSTT{key: key, model: model, http: &http.Client{Timeout: 30 * time.Second}, base: "https://api.soniox.com/v1"}
}

func (s *SonioxSTT) do(ctx context.Context, method, path string, body io.Reader, ct string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, s.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.key)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("soniox %d: %s", resp.StatusCode, short(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (s *SonioxSTT) Transcribe(ctx context.Context, audio []byte, ct string) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "voice.mp3")
	_, _ = fw.Write(audio)
	_ = mw.Close()
	var file struct {
		ID string `json:"id"`
	}
	if err := s.do(ctx, http.MethodPost, "/files", &buf, mw.FormDataContentType(), &file); err != nil {
		return "", err
	}
	defer func() { _ = s.do(context.Background(), http.MethodDelete, "/files/"+file.ID, nil, "", nil) }()
	body, _ := json.Marshal(map[string]any{
		"model": s.model, "file_id": file.ID, "language_hints": []string{"ar"},
		"context": strings.Join(DefaultVoiceLexicon.TopTerms(150), ", "),
	})
	var tr struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Error  string `json:"error_message"`
	}
	if err := s.do(ctx, http.MethodPost, "/transcriptions", bytes.NewReader(body), "application/json", &tr); err != nil {
		return "", err
	}
	defer func() { _ = s.do(context.Background(), http.MethodDelete, "/transcriptions/"+tr.ID, nil, "", nil) }()
	for tr.Status != "completed" {
		if tr.Status == "error" {
			return "", errors.New("soniox: " + tr.Error)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
		if err := s.do(ctx, http.MethodGet, "/transcriptions/"+tr.ID, nil, "", &tr); err != nil {
			return "", err
		}
	}
	var res struct {
		Text string `json:"text"`
	}
	if err := s.do(ctx, http.MethodGet, "/transcriptions/"+tr.ID+"/transcript", nil, "", &res); err != nil {
		return "", err
	}
	if t := strings.TrimSpace(res.Text); t != "" {
		return t, nil
	}
	return "", errors.New("soniox: ما فهم الصوت")
}

// ── تصحيح بالقاموس بعد التحويل (Haiku) ──

const voicePolishPrompt = `هذا نص محوّل من بصمة صوت عراقية. صلّح بس الكلمات الي واضح إنها انكتبت غلط وهي من القائمة (أسماء وكلمات شغل).
لا تغيّر المعنى ولا اللهجة ولا ترتيب الكلام، ولا تضيف ولا تشيل شي. إذا ماكو غلط رجّع النص نفسه. رجّع النص بس.
القائمة: `

// PolishTranscript يمرّر النص على Haiku ويا القاموس. يرجّع الأصلي إذا صار أي خلل.
func PolishTranscript(ctx context.Context, client *anthropic.Client, model, text string) string {
	if client == nil || DefaultVoiceLexicon == nil || strings.TrimSpace(text) == "" {
		return text
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	llm := anthropicLLM{c: client}
	resp, err := llm.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(model), MaxTokens: 600,
		System:   []anthropic.TextBlockParam{{Text: voicePolishPrompt + strings.Join(DefaultVoiceLexicon.TopTerms(400), "، ")}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(text))},
	})
	if err != nil {
		return text
	}
	RecordAIUsage("VOICE_POLISH", model, resp.Usage)
	var b strings.Builder
	for _, blk := range resp.Content {
		if t, ok := blk.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(t.Text)
		}
	}
	out := strings.TrimSpace(b.String())
	// حارس: إذا تغيّر الطول هواي فهو مو تصحيح
	if out == "" || len([]rune(out)) > len([]rune(text))*3/2+10 || len([]rune(out)) < len([]rune(text))/2 {
		return text
	}
	return out
}

func short(b []byte) string {
	s := strings.TrimSpace(string(b))
	if r := []rune(s); len(r) > 200 {
		s = string(r[:200])
	}
	return s
}

// ── ElevenLabs Scribe (ELEVENLABS_API_KEY، الموديل ELEVENLABS_STT_MODEL) ──

type ElevenLabsSTT struct {
	key, model string
	http       *http.Client
	url        string
}

func NewElevenLabsSTT(key, model string) *ElevenLabsSTT {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if model == "" {
		model = "scribe_v2"
	}
	return &ElevenLabsSTT{key: key, model: model, http: &http.Client{Timeout: 45 * time.Second},
		url: "https://api.elevenlabs.io/v1/speech-to-text"}
}

func (e *ElevenLabsSTT) Transcribe(ctx context.Context, audio []byte, ct string) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("model_id", e.model)
	_ = mw.WriteField("language_code", "ara")
	_ = mw.WriteField("tag_audio_events", "false")
	fw, _ := mw.CreateFormFile("file", "voice.mp3")
	_, _ = fw.Write(audio)
	_ = mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("xi-api-key", e.key)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := e.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("elevenlabs %d: %s", resp.StatusCode, short(raw))
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || strings.TrimSpace(out.Text) == "" {
		return "", errors.New("elevenlabs: ما فهم الصوت")
	}
	return strings.TrimSpace(out.Text), nil
}
