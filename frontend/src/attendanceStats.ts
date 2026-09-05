// ═══ حسابات سجل الدوام ═══
//
// الشاشة تعرض أربع أرقام عن الشهر: الساعات، أيام الحضور، أيام
// التأخير، أيام الغياب. أول اثنين يجون جاهزين من السيرفر. الأخيرين
// ما موجودين، ولازم ينحسبون — وكل واحد منهن يحتاج **تعريف**، والتعريف
// الغلط يطلع أرقام تظلم الموظف.

import type { DailyAttendance, MonthlyAttendanceReport } from './api'

/** دقائق من نص وقت "HH:MM" — يرجّع null لو الشكل مو مفهوم. */
function minutesOfDay(hhmm: string | null | undefined): number | null {
  if (!hhmm) return null
  const m = /^(\d{1,2}):(\d{2})/.exec(hhmm.trim())
  if (!m) return null
  return Number(m[1]) * 60 + Number(m[2])
}

/** دقائق من تاريخ ISO بالتوقيت المحلي. */
function minutesOfIso(iso: string | null): number | null {
  if (!iso) return null
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return null
  return d.getHours() * 60 + d.getMinutes()
}

/**
 * سماحية التأخير بالدقائق.
 *
 * ⚠️ بدون سماحية، الموظف الي يبصم ٨:٠١ ودوامه ٨:٠٠ ينحسب «متأخر» —
 * وهذا يخلي عدّاد التأخير يمتلئ بأرقام ما إلها معنى إداري، فينتجاهل
 * كله. السماحية تخلي الرقم يعني «تأخر فعلاً» مو «ما ضبط الثانية».
 */
export const LATE_GRACE_MINUTES = 10

/** هل هذا اليوم تأخير؟ يحتاج وقت بداية دوام معرّف للموظف. */
export function isLateDay(day: DailyAttendance, shiftStart: string | null): boolean {
  const start = minutesOfDay(shiftStart)
  if (start === null) return false // ماكو دوام محدد = ماكو تأخير نحكم بيه
  const came = minutesOfIso(day.firstCheckIn)
  if (came === null) return false
  return came > start + LATE_GRACE_MINUTES
}

/**
 * أيام العطلة الأسبوعية — الجمعة بالعراق.
 *
 * getDay(): 0 أحد ... 5 جمعة، 6 سبت.
 */
export function isWeekend(d: Date): boolean {
  return d.getDay() === 5
}

/**
 * أيام الغياب بالشهر.
 *
 * ⚠️ نعدّ **الأيام الي مرّت بس** — مو الشهر كله. لو عددنا الشهر كامل،
 * الموظف يوم ٣ من الشهر يشوف «٢٧ يوم غياب» وهو دوامه منتظم. الرقم
 * يخوّف بلا سبب ويخلي الشاشة تنتجاهل.
 *
 * ⚠️ والجمعة ما تنعد غياب — عطلة مو تغيّب.
 *
 * ⚠️ الإجازات المعتمدة مو محسوبة هنا (النظام عنده شاشة إجازات
 * منفصلة). يعني الرقم = «أيام بلا بصمة»، وممكن تكون إجازة مصدّقة.
 * لهذا اللافتة تگول «بلا بصمة» مو «تغيّب».
 */
export function countAbsentDays(report: MonthlyAttendanceReport | null, month: string): number {
  if (!report) return 0
  const [y, m] = month.split('-').map(Number)
  if (!y || !m) return 0

  const present = new Set(report.days.map((d) => d.date.slice(0, 10)))
  const today = new Date()
  const lastDayOfMonth = new Date(y, m, 0).getDate()

  // لو الشهر هو الحالي نوقف عند اليوم، وإلا نكمل الشهر كله
  const isCurrentMonth = today.getFullYear() === y && today.getMonth() + 1 === m
  const lastDay = isCurrentMonth ? today.getDate() : lastDayOfMonth

  let absent = 0
  for (let day = 1; day <= lastDay; day++) {
    const d = new Date(y, m - 1, day)
    if (isWeekend(d)) continue
    const key = `${y}-${String(m).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    if (!present.has(key)) absent++
  }
  return absent
}

/** عدد أيام التأخير بالشهر. */
export function countLateDays(report: MonthlyAttendanceReport | null, shiftStart: string | null): number {
  if (!report || !shiftStart) return 0
  return report.days.filter((d) => isLateDay(d, shiftStart)).length
}

/** ═══ آخر الحركات ═══
 *
 * النظام يسمح بأكثر من جلسة باليوم. يعني الخروج بنص اليوم والرجوع
 * بعده هو **استراحة** بالمعنى العملي — مو انصراف.
 *
 * التمييز: آخر خروج باليوم = انصراف، وأي خروج قبله = استراحة.
 * بلا هذا التمييز الموظف يشوف «تسجيل انصراف» أربع مرات بيوم واحد.
 */
export type Movement = {
  kind: 'in' | 'break' | 'back' | 'out'
  label: string
  at: string
}

export function movementsOf(day: DailyAttendance | null): Movement[] {
  if (!day || day.sessions.length === 0) return []
  const out: Movement[] = []
  const sorted = [...day.sessions].sort(
    (a, b) => new Date(a.checkIn).getTime() - new Date(b.checkIn).getTime(),
  )
  sorted.forEach((s, i) => {
    out.push({
      kind: i === 0 ? 'in' : 'back',
      label: i === 0 ? 'تسجيل حضور' : 'عودة من الاستراحة',
      at: s.checkIn,
    })
    if (s.checkOut) {
      const isLast = i === sorted.length - 1
      out.push({
        kind: isLast ? 'out' : 'break',
        label: isLast ? 'تسجيل انصراف' : 'استراحة — خروج',
        at: s.checkOut,
      })
    }
  })
  // الأحدث فوق
  return out.reverse()
}
