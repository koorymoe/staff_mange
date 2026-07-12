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
  }, [employee?.role])

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

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">تصنيفي</h2>
      <p className="mt-1 text-slate-500">
        تقييمك ونقاطك، وترتيبك بين زملائك أصحاب نفس الدور ({employee ? roleLabels[employee.role] : ''}).
      </p>

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

      {/* KPI leaderboard — scoped to my own role */}
      <div className="mt-6 rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="flex items-center justify-between border-b border-slate-100 p-5">
          <div>
            <h3 className="font-bold text-brand-900">تقييم الأداء (KPI)</h3>
            <p className="text-sm text-slate-500">مجموع نقاط التقييم وترتيبك بين نظرائك</p>
          </div>
          <div className="flex gap-2">
            {(['weekly', 'monthly'] as const).map((p) => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={`rounded-lg px-3 py-1.5 text-sm font-bold transition-colors ${period === p ? 'bg-brand-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'}`}
              >
                {p === 'weekly' ? 'أسبوعي' : 'شهري'}
              </button>
            ))}
          </div>
        </div>

        {myEntry && (
          <div className="grid grid-cols-3 gap-3 border-b border-slate-100 p-5">
            <div className="rounded-xl bg-brand-50 p-3 text-center">
              <p className="text-xl font-bold text-brand-700">{myEntry.points}</p>
              <p className="text-xs text-slate-500">نقاط التقييم</p>
            </div>
            <div className="rounded-xl bg-emerald-50 p-3 text-center">
              <p className="text-xl font-bold text-emerald-700">#{myIndex + 1}</p>
              <p className="text-xs text-slate-500">ترتيبي بين {list.length}</p>
            </div>
            <div className="rounded-xl bg-blue-50 p-3 text-center">
              <p className="text-xl font-bold text-blue-700">{myEntry.completedBookings}</p>
              <p className="text-xs text-slate-500">حجوزات منجزة</p>
            </div>
          </div>
        )}

        <div className="overflow-x-auto">
          <table className="w-full text-right text-sm">
            <thead className="bg-slate-50 text-slate-600">
              <tr>
                <th className="px-4 py-2">#</th>
                <th className="px-4 py-2">الموظف</th>
                <th className="px-4 py-2">النقاط</th>
                <th className="px-4 py-2">حجوزات منجزة</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {list.map((e, i) => (
                <tr key={e.employeeId} className={e.employeeId === employee?.id ? 'bg-brand-50' : ''}>
                  <td className="px-4 py-2 font-bold text-slate-500">{i + 1}</td>
                  <td className="px-4 py-2 font-medium text-slate-800">{e.employeeName}</td>
                  <td className={`px-4 py-2 font-bold ${e.points < 0 ? 'text-red-600' : 'text-emerald-600'}`}>{e.points}</td>
                  <td className="px-4 py-2 text-slate-500">{e.completedBookings}</td>
                </tr>
              ))}
              {list.length === 0 && (
                <tr><td colSpan={4} className="p-4 text-center text-slate-400">لا توجد بيانات تقييم بعد</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
