// ═══ صلاحيات المراقب المفصّلة (دازاي 10-25/26) ═══
// «monitoring» = الحزمة كاملة. وكل صلاحية صغيرة تفتح شاشتها بس — نفس حارس الخادم
// (requireMon[sub])، حتى ما تنعرض شاشة ترجّع 403 وتسجّل مخالفة على الموظف.
export const MON_SUBS = ['mon_inbox', 'mon_reports', 'mon_justifications', 'mon_attendance', 'mon_staff', 'mon_matrix', 'mon_field'] as const
export type MonSub = typeof MON_SUBS[number]

/** مراقب كامل: الدور، أو المدير/المالك، أو صلاحية monitoring، أو auditing (غير مهندس الجودة). */
export function isFullMonitor(role: string | undefined, permissions: string[]): boolean {
  return role === 'MONITOR' || role === 'ADMIN' || role === 'OWNER' || permissions.includes('monitoring')
    || (permissions.includes('auditing') && role !== 'QUALITY_ENGINEER')
}

/** يكدر يفتح شي يحتاج **كل** هالصلاحيات الصغيرة (أو الحزمة كاملة). */
export function monCan(role: string | undefined, permissions: string[], ...subs: MonSub[]): boolean {
  return isFullMonitor(role, permissions) || (subs.length > 0 && subs.every((s) => permissions.includes(s)))
}

/** عنده أي صلاحية مراقب (كاملة أو صغيرة) — لباب «مكتب المراقب». */
export function monAny(role: string | undefined, permissions: string[]): boolean {
  return isFullMonitor(role, permissions) || MON_SUBS.some((s) => permissions.includes(s))
}
