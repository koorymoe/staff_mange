package errlog

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestExecLogsWithCaller(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)
	Exec(nil, nil)
	if buf.Len() != 0 {
		t.Fatal("nil error must not log")
	}
	Exec(nil, errors.New("boom"))
	out := buf.String()
	if !strings.Contains(out, "boom") || !strings.Contains(out, "errlog_test.go") {
		t.Fatalf("got %q", out)
	}
}
