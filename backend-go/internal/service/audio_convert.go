package service

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ═══ 🎧 توحيد صيغة البصمة (دازاي 10-29) ═══
// الآيفون يسجّل m4a (audio/mp4)، وGemini ما يقبلها — فچان يفشل ويرجع لـWhisper
// (أبطأ بهواي وأضعف بالعراقي: ٣ دقايق وطلّع «هلو» بس)، والبصمة نفسها ما تشتغل
// بكل المتصفحات. هسه كل بصمة تتحول mp3 أحادي ١٦ كيلو: Gemini يقبلها، وWhisper
// أسرع عليها، وتشتغل بأي جهاز.

var ErrNoFFmpeg = errors.New("ffmpeg مو موجود")

// ToMP3 يحوّل أي صوت لـmp3. ملفات مؤقتة (مو pipe) لأن m4a الآيفون فهرسه بآخر الملف.
func ToMP3(ctx context.Context, audio []byte) ([]byte, error) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, ErrNoFFmpeg
	}
	dir, err := os.MkdirTemp("", "voice-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	in, out := filepath.Join(dir, "in"), filepath.Join(dir, "out.mp3")
	if err := os.WriteFile(in, audio, 0o600); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-hide_banner", "-loglevel", "error", "-y", "-i", in,
		"-vn", "-ac", "1", "-ar", "16000", "-b:a", "48k", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		return nil, errors.New("ffmpeg: " + string(b))
	}
	return os.ReadFile(out)
}
