package handler

import (
	"encoding/json"
	"net/http"
)

type errorEnvelope struct {
	Error string `json:"error"`
}

// WriteJSON يرجّع الجسم مباشرة بدون أي غلاف — يطابق شكل استجابات الباك إند
// القديم (TypeScript/Express) بالضبط حتى يقدر الفرونت إند يتعامل مع الاثنين
// بدون أي تعديل.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: message})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
