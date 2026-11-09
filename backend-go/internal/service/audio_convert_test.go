package service

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestToMP3FromM4A(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("no ffmpeg")
	}
	dir := t.TempDir()
	in := dir + "/a.m4a"
	if err := exec.Command("ffmpeg", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "sine=duration=2", "-c:a", "aac", in).Run(); err != nil {
		t.Skip(err)
	}
	b, _ := os.ReadFile(in)
	out, err := ToMP3(context.Background(), b)
	if err != nil || len(out) < 1000 {
		t.Fatalf("convert: %v (%d bytes)", err, len(out))
	}
	if !(out[0] == 'I' && out[1] == 'D' && out[2] == '3') && !(out[0] == 0xFF) {
		t.Fatalf("not mp3: % x", out[:4])
	}
}
