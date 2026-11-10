// 📐 سيت المخططات (ك) 11-10 — أشكال ردود /api/plan-sets
export type PlanGroup = { no: number; name: string; en: string; types: { code: string; name: string }[] }
export type PlanStats = { types: number; projects: number; sets: number; files: number }
export type PlanMeta = {
  groups: PlanGroup[]
  stats: PlanStats
  canManage: boolean
  projects: { id: string; code: string; name: string }[]
}
export type PlanSetRow = {
  id: string
  title: string
  projectId: string | null
  projectCode: string | null
  projectName: string | null
  notes: string | null
  createdByName: string | null
  createdAt: string
  archivedAt: string | null
  files: number
  cover: string | null
}
export type PlanFileRow = {
  id: string
  setId: string
  typeCode: string
  title: string
  fileKey: string
  fileType: string | null
  uploadedByName: string | null
  createdAt: string
}
export type PlanSetDetail = PlanSetRow & { items: PlanFileRow[]; canManage: boolean }
