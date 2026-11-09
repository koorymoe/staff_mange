package errlog

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
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

func TestErrSkipsNoRows(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)
	Err(sql.ErrNoRows)
	Err(fmt.Errorf("wrap: %w", sql.ErrNoRows))
	if buf.Len() != 0 {
		t.Fatalf("ErrNoRows must not log: %q", buf.String())
	}
	Err(errors.New("real"))
	if !strings.Contains(buf.String(), "real") {
		t.Fatal("real error must log")
	}
}
