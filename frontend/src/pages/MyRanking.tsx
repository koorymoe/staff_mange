import { useEffect, useState } from 'react'
import { api, type Stats, type RoleKpiLeaderboard } from '../api'
import { useSession } from '../session'
import { roleLabels } from '../session'

const levels = [
  { level: 1, label: 'متدرب', min: 0 },
  { level: 2, label: 'فني مبتدئ', min: 3 },
  { level: 3, label: 'فني', min: 6 },
  { level: 4, label: 'فني متمرس', min: 10 },
  { level: 5, label: 'فني خبير', min: 15 },
]

const BOOKINGS_PER_RANK = 10

export default function MyRanking() {
  const { employee } = useSession()
  const [stats, setStats] = useState<Stats | null>(null)
  const [board, setBoard] = useState<RoleKpiLeaderboard | null>(null)
  const [period, setPeriod] = useState<'weekly' | 'monthly'>('weekly')

  const isTechnician = employee?.role === 'TECHNICIAN'

  useEffect(() => {
    if (isTechnician) api.getStats().then(setStats).catch(() => setStats(null))
  }, [isTechnician])

  useEffect(() => {
    if (!employee) return
    api.getRoleKpiLeaderboard(employee.role).then(setBoard).catch(() => setBoard(null))
  }, [employee])

  const skillCount = employee?.skills.filter((s) => s.canPerform).length || 0
  const currentLevel = [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
  const nextLevel = levels.find((l) => l.min > skillCount)

  const myTechStat = stats?.technicianStats.find((s) => s.employeeId === employee?.id)
  const completedCount = myTechStat?.completed || 0

  const rank = Math.floor(completedCount / BOOKINGS_PER_RANK) + 1
  const remainingForNextRank = BOOKINGS_PER_RANK - (completedCount % BOOKINGS_PER_RANK)

  const sortedTechs = stats
    ? [...stats.technicianStats].sort((a, b) => b.completed - a.completed)
    : []
  const myPosition = sortedTechs.findIndex((s) => s.employeeId === employee?.id)

  const list = board ? (period === 'weekly' ? board.weekly : board.monthly) : []
  const myIndex = list.findIndex((e) => e.employeeId === employee?.id)
  const myEntry = myIndex >= 0 ? list[myIndex] : null

  // ── أرقام الموظف بالفترة المختارة ──
  const myRank = myIndex >= 0 ? myIndex + 1 : null
  const roleLabel = employee ? roleLabels[employee.role] : ''
  // معدل الإنجاز: المنجز من الي انكلّف بيه. الي خلّص ٨ من ٨ مو مثل
  // الي خلّص ٨ من ٢٠ — والرقم المطلق لحاله يخفي هذا الفرق.
  const completionRate = myEntry && myEntry.assignedBookings > 0
    ? Math.round((myEntry.completedBookings / myEntry.assignedBookings) * 100)
    : null
  const periodDays = period === 'weekly' ? 7 : 30
  const periodLabel = period === 'weekly' ? 'الأسبوع الماضي' : 'الشهر الماضي'
  const top3 = list.slice(0, 3)

  return (
    <div dir="rtl" className="space-y-5">
      {/* ═══ العنوان + التبديل ═══ */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-amber-50 text-xl">🏆</span>
          <div>
            <h2 className="text-2xl font-black text-[#0f2040]">تصنيفي</h2>
            <p className="text-xs text-slate-500">
              مقارنة بين {roleLabel} فقط
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          {(['monthly', 'weekly'] as const).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`rounded-lg px-4 py-2 text-xs font-bold transition ${
                period === p ? 'bg-[#2c5aad] text-white shadow-md' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              📅 {p === 'weekly' ? 'أسبوعي' : 'شهري'}
            </button>
          ))}
        </div>
      </div>

      <p className="rounded-lg bg-slate-50 px-3 py-2 text-[11px] text-slate-500">
        ⓘ الترتيب بين أصحاب نفس الدور — آخر {periodDays} يوم، والمقارنة مع {periodLabel}
      </p>

      {/* ═══ البطاقات الأربع ═══ */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <RankCard
          icon="🏆" label="ترتيبي الحالي"
          value={myRank ? `#${myRank}` : '—'}
          tone="emerald"
          delta={myEntry?.rankDelta ?? 0}
          deltaSuffix="مركز"
          note={myRank === 1 ? '🎉 حافظ على مركزك!' : myRank ? `من ${list.length}` : undefined}
        />
        <RankCard
          icon="⭐" label="نقاط التقييم"
          value={myEntry ? String(myEntry.points) : '—'}
          tone="amber"
          delta={myEntry?.pointsDelta ?? 0}
          deltaSuffix="نقطة"
        />
        <RankCard
          icon="🗂️" label="حجوزات منجزة"
          value={myEntry ? String(myEntry.completedBookings) : '—'}
          tone="sky"
          note={myEntry && myEntry.assignedBookings > 0 ? `من ${myEntry.assignedBookings} مكلّف بيها` : undefined}
        />
        <RankCard
          icon="📈" label="معدل الإنجاز"
          value={completionRate !== null ? `${completionRate}%` : '—'}
          tone="violet"
          note={completionRate === null ? 'ماكو حجوزات بالفترة' : undefined}
          bar={completionRate ?? undefined}
        />
      </div>

      {/* ═══ أعلى ٣ ═══ */}
      {top3.length > 0 && (
        <div className="rounded-2xl border border-amber-200 bg-amber-50/40 p-5">
          <h3 className="mb-4 text-sm font-extrabold text-[#0f2040]">
            👑 أعلى ٣ {roleLabel} {period === 'weekly' ? 'هذا الأسبوع' : 'هذا الشهر'}
          </h3>
          {/* الترتيب البصري: الثالث يمين، الأول وسط وأكبر، الثاني يسار */}
          <div className="flex items-end justify-center gap-3">
            {[top3[2], top3[0], top3[1]].map((e, slot) => {
              if (!e) return <div key={slot} className="flex-1" />
              const place = e === top3[0] ? 1 : e === top3[1] ? 2 : 3
              const medal = place === 1 ? '🥇' : place === 2 ? '🥈' : '🥉'
              const placeLabel = place === 1 ? 'الأول' : place === 2 ? 'الثاني' : 'الثالث'
              const isMe = e.employeeId === employee?.id
              return (
                <div
                  key={e.employeeId}
                  className={`flex-1 rounded-2xl border-2 bg-white p-4 text-center ${
                    place === 1 ? 'border-amber-300 shadow-lg' : 'border-slate-200'
                  } ${isMe ? 'ring-2 ring-sky-300' : ''}`}
                >
                  <div className="mb-1 text-2xl">{medal}</div>
                  <p className="truncate text-xs font-bold text-slate-700">{e.employeeName}</p>
                  <p className={`mt-1 text-2xl font-black ${place === 1 ? 'text-amber-600' : 'text-slate-600'}`}>
                    {e.points}
                  </p>
                  {e.pointsDelta !== 0 && (
                    <p className={`text-[10px] font-bold ${e.pointsDelta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
                      {e.pointsDelta > 0 ? '▲' : '▼'} {Math.abs(e.pointsDelta)}
                    </p>
                  )}
                  <span className={`mt-2 inline-block rounded-md px-2 py-0.5 text-[10px] font-bold ${
                    place === 1 ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-600'
                  }`}>
                    {placeLabel}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {isTechnician && (
        <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 p-6 text-white shadow-lg shadow-brand-900/20">
            <p className="text-sm text-brand-100">المستوى (حسب المهارات المعتمدة)</p>
            <p className="mt-1 text-3xl font-extrabold">
              {currentLevel.level} - {currentLevel.label}
            </p>
            <p className="mt-2 text-sm text-brand-100">
              {skillCount} مهارة معتمدة
              {nextLevel
                ? ` - يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي إلى "${nextLevel.label}"`
                : ' - وصلت لأعلى مستوى!'}
            </p>
          </div>

          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <p className="text-sm text-slate-500">الرانك</p>
            <p className="mt-1 text-3xl font-extrabold text-emerald-700">{rank}</p>
            <p className="mt-2 text-sm text-slate-500">
              {completedCount} حجز منجز - يحتاج {remainingForNextRank} حجز إضافي للصعود للرانك التالي
            </p>
            {myPosition >= 0 && (
              <p className="mt-2 text-sm text-slate-500">
                ترتيبك بين الفنيين: <span className="font-bold text-brand-700">#{myPosition + 1}</span> من{' '}
                {sortedTechs.length}
              </p>
            )}
          </div>
        </div>
      )}

      {/* ═══ الجدول + شرح النقاط ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* الترتيب الكامل */}
        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm lg:col-span-2">
          <div className="border-b border-slate-100 px-5 py-4">
            <h3 className="text-sm font-extrabold text-[#0f2040]">👥 ترتيب {roleLabel}</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-right text-xs">
              <thead className="bg-slate-50 text-slate-600">
                <tr>
                  <th className="px-4 py-2.5 font-bold">#</th>
                  <th className="px-4 py-2.5 font-bold">الموظف</th>
                  <th className="px-4 py-2.5 font-bold">نقاط التقييم</th>
                  <th className="px-4 py-2.5 font-bold">الحجوزات المنجزة</th>
                  <th className="px-4 py-2.5 font-bold">الالتزام</th>
                  <th className="px-4 py-2.5 font-bold">التغيير</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {list.map((e, i) => {
                  const isMe = e.employeeId === employee?.id
                  // الالتزام = أيام الحضور من أيام الفترة (عدا الجمع)
                  const workDays = period === 'weekly' ? 6 : 26
                  const rate = Math.min(100, Math.round((e.attendedDays / workDays) * 100))
                  const commit = rate >= 90
                    ? { text: 'ممتاز', cls: 'bg-emerald-50 text-emerald-700' }
                    : rate >= 75
                      ? { text: 'جيد جداً', cls: 'bg-sky-50 text-sky-700' }
                      : rate >= 50
                        ? { text: 'جيد', cls: 'bg-amber-50 text-amber-700' }
                        : { text: 'يحتاج تحسين', cls: 'bg-red-50 text-red-700' }
                  return (
                    <tr key={e.employeeId} className={isMe ? 'bg-sky-50/60' : 'hover:bg-slate-50'}>
                      <td className="px-4 py-3">
                        <span className={`inline-flex h-6 w-6 items-center justify-center rounded-md text-[11px] font-black ${
                          i === 0 ? 'bg-amber-100 text-amber-800'
                          : i === 1 ? 'bg-slate-200 text-slate-700'
                          : i === 2 ? 'bg-orange-100 text-orange-800'
                          : 'text-slate-400'
                        }`}>{i + 1}</span>
                      </td>
                      <td className="px-4 py-3 font-bold text-slate-800">
                        {e.employeeName}{isMe && <span className="mr-1 text-[10px] text-sky-600">(أنت)</span>}
                      </td>
                      <td className={`px-4 py-3 font-black ${e.points < 0 ? 'text-red-600' : 'text-slate-800'}`}>{e.points}</td>
                      <td className="px-4 py-3 text-slate-600">{e.completedBookings}</td>
                      <td className="px-4 py-3">
                        <span className={`rounded-md px-2 py-0.5 text-[10px] font-bold ${commit.cls}`} title={`${e.attendedDays} يوم حضور`}>
                          {commit.text}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        {e.pointsDelta === 0 ? (
                          <span className="text-slate-400">0 —</span>
                        ) : (
                          <span className={`font-bold ${e.pointsDelta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
                            {e.pointsDelta > 0 ? '▲' : '▼'} {e.pointsDelta > 0 ? '+' : ''}{e.pointsDelta}
                          </span>
                        )}
                      </td>
                    </tr>
                  )
                })}
                {list.length === 0 && (
                  <tr><td colSpan={6} className="p-6 text-center text-slate-400">ماكو بيانات تقييم بعد</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* ═══ كيف تنحسب النقاط — الحقيقة مو معادلة مزيّنة ═══ */}
        <div className="space-y-4">
          <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <h3 className="mb-3 text-sm font-extrabold text-[#0f2040]">⚖️ شلون تنحسب النقاط؟</h3>

            {/* ⚠️ ما نعرض معادلة أوزان (٤٠٪ حجوزات + ٢٠٪ استجابة...) لأن
                النظام **ما يشتغل بيها**. النقاط تجي من تقييم المدير
                اليدوي. عرض معادلة ما تنطبق يخلي الموظف يشتغل على أرقام
                ما إلها تأثير، ويفقد ثقته بالشاشة أول ما يكتشف. */}
            <div className="space-y-2.5 text-[11px]">
              <div className="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
                <p className="font-bold text-slate-700">⭐ نقاط التقييم</p>
                <p className="mt-0.5 leading-relaxed text-slate-500">
                  تجي من تقييم المدير المباشر — يزيد نقاط على الشغل الزين، ويخصم على المخالفة مع سبب مكتوب.
                </p>
              </div>
              <div className="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
                <p className="font-bold text-slate-700">🗂️ الحجوزات المنجزة</p>
                <p className="mt-0.5 leading-relaxed text-slate-500">
                  تنعدّ تلقائياً من النظام — الحجوزات الي انكلّفت بيها ووصلت «منجز».
                </p>
              </div>
              <div className="rounded-xl border border-slate-100 bg-slate-50/60 p-3">
                <p className="font-bold text-slate-700">🕐 الالتزام بالدوام</p>
                <p className="mt-0.5 leading-relaxed text-slate-500">
                  من بصمات حضورك — أيام الحضور من أيام العمل بالفترة.
                </p>
              </div>
            </div>

            <p className="mt-3 rounded-lg bg-sky-50 px-3 py-2 text-[10px] leading-relaxed text-sky-800">
              الترتيب على <b>نقاط التقييم</b>، وعند التساوي الأكثر إنجازاً يتقدّم.
            </p>
          </div>

          {/* إنجازاتك بالفترة */}
          {myEntry && (
            <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
              <h3 className="mb-3 text-sm font-extrabold text-[#0f2040]">
                🏅 إنجازاتك {period === 'weekly' ? 'هذا الأسبوع' : 'هذا الشهر'}
              </h3>
              <div className="grid grid-cols-3 gap-2">
                <Achievement
                  icon="✅" label="أيام الحضور"
                  value={`${myEntry.attendedDays}`}
                  ok={myEntry.attendedDays > 0}
                />
                <Achievement
                  icon="⚡" label="معدل الإنجاز"
                  value={completionRate !== null ? `${completionRate}%` : '—'}
                  ok={(completionRate ?? 0) >= 80}
                />
                <Achievement
                  icon="📊" label="تقييمات"
                  value={`${myEntry.evaluationCount}`}
                  ok={myEntry.points >= 0}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

/* ───── بطاقة رقم بأعلى الصفحة ───── */

function RankCard({ icon, label, value, tone, delta, deltaSuffix, note, bar }: {
  icon: string; label: string; value: string
  tone: 'emerald' | 'amber' | 'sky' | 'violet'
  delta?: number; deltaSuffix?: string; note?: string; bar?: number
}) {
  const tones: Record<string, { text: string; bg: string; bar: string }> = {
    emerald: { text: 'text-emerald-700', bg: 'bg-emerald-50', bar: 'bg-emerald-500' },
    amber:   { text: 'text-amber-700',   bg: 'bg-amber-50',   bar: 'bg-amber-500' },
    sky:     { text: 'text-sky-700',     bg: 'bg-sky-50',     bar: 'bg-sky-500' },
    violet:  { text: 'text-violet-700',  bg: 'bg-violet-50',  bar: 'bg-violet-500' },
  }
  const t = tones[tone]
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <p className="text-[11px] font-medium text-slate-500">{label}</p>
        <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-sm ${t.bg}`}>{icon}</span>
      </div>
      <p className={`mt-1 text-2xl font-black ${t.text}`}>{value}</p>
      {delta !== undefined && delta !== 0 && (
        <p className={`text-[10px] font-bold ${delta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
          {delta > 0 ? '▲' : '▼'} {delta > 0 ? '+' : ''}{delta} {deltaSuffix}
        </p>
      )}
      {note && <p className="mt-0.5 text-[10px] text-slate-500">{note}</p>}
      {bar !== undefined && (
        <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-slate-100">
          <div className={`h-full rounded-full ${t.bar}`} style={{ width: `${Math.min(100, bar)}%` }} />
        </div>
      )}
    </div>
  )
}

/* ───── إنجاز صغير ───── */

function Achievement({ icon, label, value, ok }: {
  icon: string; label: string; value: string; ok: boolean
}) {
  return (
    <div className={`rounded-xl border p-2.5 text-center ${ok ? 'border-emerald-200 bg-emerald-50/60' : 'border-slate-200 bg-slate-50'}`}>
      <div className="text-base">{icon}</div>
      <p className={`mt-0.5 text-sm font-black ${ok ? 'text-emerald-700' : 'text-slate-500'}`}>{value}</p>
      <p className="text-[9px] leading-tight text-slate-500">{label}</p>
    </div>
  )
}
