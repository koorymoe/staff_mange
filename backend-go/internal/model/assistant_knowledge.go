package model

import "time"

// AssistantKnowledge سطر معرفة "تعلمه" المساعد الذكي من محادثة مع موظف —
// حقيقة عامة، معلومة علّمه إياها موظف، أو نتيجة بحث ويب مفيدة. نخزنها
// ونعيد استرجاعها بمحادثات لاحقة (RAG بسيط بدون قاعدة بيانات متجهات) حتى
// المساعد يبني معرفة حقيقية مع الوقت بدل ما ينسى كل شي بعد كل رد.
type AssistantKnowledge struct {
	ID                    string    `db:"id" json:"id"`
	Topic                 string    `db:"topic" json:"topic"`
	Content               string    `db:"content" json:"content"`
	LearnedFromEmployeeID *string   `db:"learnedFromEmployeeId" json:"learnedFromEmployeeId"`
	CreatedAt             time.Time `db:"createdAt" json:"createdAt"`
}
