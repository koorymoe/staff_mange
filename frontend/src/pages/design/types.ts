// 🏗️ صيت التصاميم (ع) 11-08 — أنواع مشتركة بين الشاشات
export type Discipline = { code: string; name: string; icon: string }
export type DesignTool = { id: string; discipline: string; name: string; url: string; howTo: string; sort: number; active: boolean }
export type DesignMeta = {
  disciplines: Discipline[]
  tools: DesignTool[]
  canManage: boolean
  canReview: boolean
  projects: { id: string; code: string; name: string }[]
}
export type DesignSetRow = {
  id: string; title: string; projectId: string | null; projectCode: string | null
  customerName: string | null; notes: string | null; createdByName: string | null; createdAt: string; archivedAt: string | null
  sections: number; approved: number; inReview: number; late: number
}
export type DesignFile = {
  id: string; sectionId: string; title: string; fileKey: string | null; fileType: string | null
  url: string | null; toolName: string | null; uploadedByName: string | null; createdAt: string
}
export type DesignSection = {
  id: string; setId: string; discipline: string
  status: 'TODO' | 'IN_PROGRESS' | 'REVIEW' | 'APPROVED' | 'RETURNED'
  assigneeId: string | null; assigneeName: string | null; dueDate: string | null
  reviewNote: string | null; reviewedByName: string | null; reviewedAt: string | null; taskId: string | null
  files: DesignFile[]
}
export type DesignSetDetail = DesignSetRow & { sections: DesignSection[]; canManage: boolean; canReview: boolean }

export const SECTION_STATUS: Record<DesignSection['status'], { label: string; cls: string }> = {
  TODO: { label: 'ما بدأ', cls: 'bg-slate-100 text-slate-600' },
  IN_PROGRESS: { label: '✏️ قيد التصميم', cls: 'bg-sky-100 text-sky-800' },
  REVIEW: { label: '⏳ ينتظر المراجعة', cls: 'bg-amber-100 text-amber-800' },
  APPROVED: { label: '✅ معتمد', cls: 'bg-emerald-100 text-emerald-800' },
  RETURNED: { label: '↩️ رجع للتعديل', cls: 'bg-red-100 text-red-700' },
}
