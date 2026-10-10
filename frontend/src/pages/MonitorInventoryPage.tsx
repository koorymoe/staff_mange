import { useEffect, useMemo, useState } from 'react'
import { api, type TeamInventoryCheck, type TeamInventoryShortageReason } from '../api'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'
import SearchBar from '../components/SearchBar'
import Pager from '../components/Pager'
import { matches } from '../utils/search'

/**
 * ═══ متابعة الجرد — للمراقب ═══
 *
 * «جرد الأدوات يكون عبارة عن صلاحيات… وكون اكو صلاحية اسمها متابعة
 * الجرد أنطيها لإداري الكوادر وللمراقب ولأبو الكميات».
 *
 * ⚠️ هذي **مو** شاشة `/inventory` مال أبو الكميات (الي يضيف ويحذف
 * أدوات ويوافق على الطلبات). هذي **متابعة**: منو جرد ومنو ما جرد،
 * وشنو ناقص وليش. المراقب يشوف ويحاسب، **ما ينفّذ** — نفس مبدأ «يشوف
 * الجودة كاملة بلا ما يتصل بالزبون».
 *
 * والي يشوفها: الليدر (جروده هو) · أو منو انمنح «متابعة الجرد» ·
 * أو المراقب بصلاحيته الافتراضية.
 */
const REASON_LABELS: Record<TeamInventoryShortageReason, string> = {
  FORGOTTEN: 'نسيان بمكان معين',
  DAMAGED: 'تلف — تحتاج بديل',
  UNKNOWN: 'ما يعرف',
}
const ROLE_LABELS: Record<string, string> = {
  LEADER: 'الليدر', EMPLOYEE1: 'الموظف الأول', EMPLOYEE2: 'الموظف الثاني',
}
const PAGE = 10
const fmt = (v: string) => new Date(v).toLocaleString('en-CA', { hour12: false }).slice(0, 16)

export default function MonitorInventoryPage({ embedded }: { embedded?: boolean } = {}) {
  // null = لسه ما انجاب · [] = انجاب وماكو نتيجة. التفريق مقصود.
  const [checks, setChecks] = useState<TeamInventoryCheck[] | null>(null)
  const [failed, setFailed] = useState(false)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [openId, setOpenId] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getTeamInventoryChecks()
      .then((rows) => { if (alive) setChecks(rows) })
      .catch(() => { if (alive) setFailed(true) })
    return () => { alive = false }
  }, [])

  const filtered = useMemo(() => {
    const rows = checks || []
    if (!search.trim()) return rows
    return rows.filter((c) => matches(
      [c.leader?.name, c.employee1?.name, c.employee2?.name,
       ...c.items.filter((i) => !i.present).map((i) => i.toolName)],
      search,
    ))
  }, [checks, search])

  // ⚠️ الأرقام تنحسب من نفس القائمة الي تنعرض — ماكو مصدر ثاني يفترق.
  const stats = useMemo(() => {
    const rows = checks || []
    let missing = 0, damaged = 0
    for (const c of rows) {
      for (const it of c.items) {
        if (it.present) continue
        missing++
        if (it.reason === 'DAMAGED') damaged++
      }
    }
    return { total: rows.length, missing, damaged }
  }, [checks])

  if (failed) return <EmptyState icon="📦" title="تعذر جلب الجرود" reason="جرّب تحدّث الصفحة" />
  if (!checks) return <div className="rounded-2xl bg-white p-8 text-center text-slate-400">جاري الجلب...</div>

  const shown = filtered.slice((page - 1) * PAGE, page * PAGE)

  return (
    <div dir="rtl" className="space-y-4">
      {!embedded && (
        <PageHeader title="📦 متابعة الجرد" subtitle="منو جرد ومنو ما جرد، وشنو ناقص وليش — متابعة بلا تنفيذ" />
      )}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <StatTile label="جرود مسجّلة" icon="📋" tone="info" value={stats.total} />
        {/* ⚠️ «—» لا «٠» لمن ماكو ولا جرد أصلاً: صفر معناه «جردوا وماكو نقص»،
            والفراغ معناه «ماكو جرد بعد» — والخلط يظلم الفريق. */}
        <StatTile label="أدوات ناقصة" icon="⚠️" tone="warning"
          value={stats.total === 0 ? '—' : stats.missing} />
        <StatTile label="منها تالفة" icon="🛠️" tone="danger"
          value={stats.total === 0 ? '—' : stats.damaged} />
      </div>

      <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <SearchBar value={search} onChange={(v) => { setSearch(v); setPage(1) }}
          placeholder="بحث باسم الليدر أو الموظف أو الأداة الناقصة..." />

        {shown.length === 0 ? (
          <EmptyState icon="📦"
            title={filtered.length === 0 && (checks.length > 0) ? 'ماكو نتيجة للبحث' : 'ماكو ولا جرد مسجّل'}
            reason={checks.length === 0 ? 'الجرد يسجّله الليدر من شاشة «الجرد» عنده' : undefined} />
        ) : (
          <div className="mt-3 space-y-2">
            {shown.map((c) => {
              const missing = c.items.filter((i) => !i.present)
              const open = openId === c.id
              return (
                <div key={c.id} className="rounded-xl border border-slate-200">
                  {/* الضغط يفتح التفاصيل **جوّا الصف** — مو بلوحة أسفل الصفحة */}
                  <button onClick={() => setOpenId(open ? null : c.id)}
                    className="flex w-full flex-wrap items-center justify-between gap-2 p-3 text-right">
                    <span>
                      <span className="font-bold text-slate-800">{c.leader?.name || 'ليدر غير معروف'}</span>
                      {(c.employee1 || c.employee2) && (
                        <span className="mr-2 text-xs text-slate-500">
                          مع {[c.employee1?.name, c.employee2?.name].filter(Boolean).join(' و')}
                        </span>
                      )}
                      <span className="block text-xs text-slate-400">{fmt(c.createdAt)}</span>
                    </span>
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                      missing.length === 0 ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-100 text-amber-800'}`}>
                      {missing.length === 0 ? '✓ مكتمل' : `ناقص ${missing.length}`}
                    </span>
                  </button>

                  {open && (
                    <div className="border-t border-slate-100 p-3">
                      {missing.length === 0 ? (
                        <p className="text-sm text-emerald-700">كل الأدوات موجودة عند الفريق</p>
                      ) : (
                        <table className="min-w-full text-right text-sm">
                          <thead>
                            <tr className="text-xs text-slate-500">
                              <th className="py-1 font-semibold">الأداة الناقصة</th>
                              <th className="py-1 font-semibold">عند منو</th>
                              <th className="py-1 font-semibold">السبب</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-slate-100">
                            {missing.map((it) => (
                              <tr key={it.id}>
                                <td className="py-1.5 font-medium text-slate-700">{it.toolName}</td>
                                <td className="py-1.5 text-slate-600">{ROLE_LABELS[it.personRole] || it.personRole}</td>
                                <td className="py-1.5 text-slate-600">
                                  {it.reason ? REASON_LABELS[it.reason] : '—'}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
            <Pager page={page} perPage={PAGE} total={filtered.length} unit="جرد"
              onPage={setPage} onPerPage={() => {}} perPageOptions={[PAGE]} />
          </div>
        )}
      </div>
    </div>
  )
}
