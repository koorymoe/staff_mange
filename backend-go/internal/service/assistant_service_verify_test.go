package service

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestGeminiRequestSearchToolJSON يتأكد إن تفعيل البحث يولّد بالضبط الشكل
// اللي توثقه واجهة Gemini REST API: "tools":[{"google_search":{}}].
func TestGeminiRequestSearchToolJSON(t *testing.T) {
	req := geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: "hi"}}}},
		Tools:    []geminiTool{{}},
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"tools":[{"google_search":{}}]`) {
		t.Fatalf("expected tools field with google_search, got: %s", got)
	}
	t.Logf("marshaled: %s", got)
}

// TestGeminiRequestNoToolsOmitted يتأكد إن عدم تفعيل البحث ما يضيف حقل tools إطلاقاً.
func TestGeminiRequestNoToolsOmitted(t *testing.T) {
	req := geminiRequest{Contents: []geminiContent{{Parts: []geminiPart{{Text: "hi"}}}}}
	b, _ := json.Marshal(req)
	if strings.Contains(string(b), "tools") {
		t.Fatalf("expected no tools field, got: %s", string(b))
	}
}

// TestGeminiResponseWithGroundingMetadataUnmarshals يتأكد إن رد فيه groundingMetadata
// (يوصل لما البحث يشتغل فعلياً) ينفك بدون خطأ ويعطينا النص الصحيح.
func TestGeminiResponseWithGroundingMetadataUnmarshals(t *testing.T) {
	mockJSON := `{
		"candidates": [{
			"content": {"parts": [{"text": "جواب المساعد هنا"}]},
			"groundingMetadata": {
				"searchEntryPoint": {"renderedContent": "..."},
				"groundingChunks": [{"web": {"uri": "https://example.com", "title": "مثال"}}],
				"webSearchQueries": ["استعلام تجريبي"]
			}
		}]
	}`
	var gr geminiResponse
	if err := json.Unmarshal([]byte(mockJSON), &gr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(gr.Candidates) != 1 || len(gr.Candidates[0].Content.Parts) != 1 {
		t.Fatalf("unexpected structure: %+v", gr)
	}
	if gr.Candidates[0].Content.Parts[0].Text != "جواب المساعد هنا" {
		t.Fatalf("unexpected text: %s", gr.Candidates[0].Content.Parts[0].Text)
	}
	if gr.Candidates[0].GroundingMetadata == nil {
		t.Fatalf("expected groundingMetadata to be captured")
	}
}

// TestExtractLearnedKnowledge_WithMarker يتأكد إن العلامة تُستخرج وتُشال من
// النص المعروض للموظف.
func TestExtractLearnedKnowledge_WithMarker(t *testing.T) {
	raw := "جواب عادي للموظف.\n<<LEARN topic=\"إنفرترات\">>الإنفرتر الهجين أفضل للاستخدام المنزلي<<END_LEARN>>"
	cleaned, items := extractLearnedKnowledge(raw)

	if strings.Contains(cleaned, "LEARN") {
		t.Fatalf("marker not stripped, got: %s", cleaned)
	}
	if cleaned != "جواب عادي للموظف." {
		t.Fatalf("unexpected cleaned text: %q", cleaned)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 learned item, got %d", len(items))
	}
	if items[0].Topic != "إنفرترات" {
		t.Fatalf("unexpected topic: %s", items[0].Topic)
	}
	if items[0].Content != "الإنفرتر الهجين أفضل للاستخدام المنزلي" {
		t.Fatalf("unexpected content: %s", items[0].Content)
	}
}

// TestExtractLearnedKnowledge_MultipleMarkers يتأكد من دعم أكثر من علامة بنفس الرد.
func TestExtractLearnedKnowledge_MultipleMarkers(t *testing.T) {
	raw := "جواب.\n<<LEARN topic=\"أ\">>محتوى أ<<END_LEARN>>\n<<LEARN topic=\"ب\">>محتوى ب<<END_LEARN>>"
	cleaned, items := extractLearnedKnowledge(raw)
	if cleaned != "جواب." {
		t.Fatalf("unexpected cleaned text: %q", cleaned)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d: %+v", len(items), items)
	}
}

// TestExtractLearnedKnowledge_NoMarker يتأكد إن رد بدون علامة يرجع كما هو تماماً.
func TestExtractLearnedKnowledge_NoMarker(t *testing.T) {
	raw := "هلا وغلا، شلونك اليوم؟"
	cleaned, items := extractLearnedKnowledge(raw)
	if cleaned != raw {
		t.Fatalf("expected unchanged text, got: %q", cleaned)
	}
	if items != nil {
		t.Fatalf("expected no items, got: %+v", items)
	}
}
