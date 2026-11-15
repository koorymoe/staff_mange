package repository

import "testing"

func TestSequenceFormatsPast9999(t *testing.T) {
	if got := formatQuotationNumber("AM-2026-", 7); got != "AM-2026-0007" {
		t.Fatal(got)
	}
	if got := formatQuotationNumber("AM-2026-", 10000); got != "AM-2026-10000" {
		t.Fatal(got)
	}
	if got := formatProcurementCode(10000); got != "PR-10000" {
		t.Fatal(got)
	}
}

func TestAgentKeyFromNote(t *testing.T) {
	if k := agentKeyFromNote(agentNotePrefix("abc123") + ": work"); k != "abc123" {
		t.Fatal(k)
	}
	if k := agentKeyFromNote("manual note"); k != "" {
		t.Fatal(k)
	}
}
