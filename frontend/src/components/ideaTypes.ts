// 💡 أنواع «عندي فكرة» (دازاي 11-21 — كايزن)
export const IDEA_CATEGORIES: Record<string, string> = { TIME: '⏱️ وقت', MONEY: '💰 فلوس', QUALITY: '⭐ جودة', SAFETY: '🦺 سلامة', OTHER: '💡 ثاني' }
export const IDEA_STATUS: Record<string, string> = { NEW: '⏳ تنتظر رد', ACCEPTED: '✅ نطبقها', TRYING: '🧪 نجربها', REJECTED: '❌ لا' }
export interface Idea {
  id: string; employeeId: string; employee: string; text: string; category: string; status: string
  reply: string | null; decidedBy: string | null; decidedAt: string | null; late: boolean; createdAt: string
}

