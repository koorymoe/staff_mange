import { useEffect, useMemo, useState } from 'react'
import { api, type CoreMatrix, type CoreMatrixRow } from '../api'
import { useSession } from '../session'

// ═══ 🏴‍☠️ مصفوفة الصلاحيات — ون بيس (11-10) ═══
// «هاي الأشياء مهمة جداً، إذا نريد ندقق بيهن» — كل فعل بالنظام، منو يگدر عليه،
// ومساراته. المصدر: النواة نفسها (/api/core/matrix)، فالي تشوفه هنا هو الي
// يطبّقه السيرفر بالضبط. للمالك بس.

function who(r: CoreMatrixRow): { text: string; tone: string } {
  if (r.public) return { text: 'عام (بلا دخول)', tone: 'bg-red-100 text-red-800' }
  if (r.ownerOnly) return { text: 'المالك بس', tone: 'bg-purple-100 text-purple-800' }
  const parts: string[] = []
  if (r.roles?.length) parts.push('دور: ' + r.roles.join('، '))
  if (r.perms?.length) parts.push('صلاحية: ' + r.perms.join('، '))
  if (r.leader) parts.push('الليدر')
  if (r.hasCheck) parts.push('فحص على الشي نفسه')
  if (r.anyone && !parts.length) return { text: 'أي موظف', tone: 'bg-slate-100 text-slate-700' }
  if (r.anyone) parts.unshift('أي موظف +')
  const t = parts.join(' · ') + (r.strict ? ' (حتى المدير)' : '')
  return { text: t || '—', tone: r.hasCheck && !r.roles?.length && !r.perms?.length ? 'bg-sky-100 text-sky-800' : 'bg-emerald-100 text-emerald-800' }
}

export default function CoreMatrixPage() {
  const { employee, permissions } = useSession()
  // المالك، أو مدير نظام نطاه المالك هالشاشة (👑 صلاحيات المالك)
  const isOwner = employee?.actualRole === 'OWNER' || permissions.includes('owner:system.matrix')
  const [data, setData] = useState<CoreMatrix | null>(null)
  const [error, setError] = useState('')
  const [q, setQ] = useState('')
  const [kind, setKind] = useState<'all' | 'owner' | 'rule' | 'anyone' | 'unused'>('all')
  useEffect(() => {
    if (!isOwner) return
    let alive = true
    api.getCoreMatrix().then((d) => { if (alive) setData(d) }).catch((e: Error) => { if (alive) setError(e.message) })
    return () => { alive = false }
  }, [isOwner])
  const rows = useMemo(() => {
    const t = q.trim().toLowerCase()
    return (data?.actions ?? []).filter((r) => {
      if (kind === 'owner' && !r.ownerOnly) return false
      if (kind === 'anyone' && !(r.anyone && !r.roles?.length && !r.perms?.length)) return false
      if (kind === 'rule' && (r.ownerOnly || (r.anyone && !r.roles?.length && !r.perms?.length))) return false
      if (kind === 'unused' && !r.unused) return false
      if (!t) return true
      return [r.action, r.label, ...(r.routes ?? []), ...(r.roles ?? []), ...(r.perms ?? [])].some((x) => x.toLowerCase().includes(t))
    })
  }, [data, q, kind])
  if (!isOwner) return <p className="p-6 text-center text-slate-500">الصفحة غير موجودة</p>
  if (error) return <p role="alert" className="m-4 rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>
  if (!data) return <p className="p-6 text-sm text-slate-500">جاري التحميل…</p>
  const a = data.actions
  const count = (f: (r: CoreMatrixRow) => boolean) => a.filter(f).length
  return (
    <div className="space-y-4 p-4" dir="rtl">
      <header className="space-y-1">
        <h1 className="text-xl font-extrabold text-slate-900">🏴‍☠️ مصفوفة الصلاحيات</h1>
        <p className="text-sm text-slate-600">كل فعل بالنظام ومنو يگدر عليه — نفس الي يطبّقه السيرفر بالضبط. الرفض دائماً «مو موجود»، والمحاولة تنحسب مخالفة.</p>
      </header>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
        {([['all', 'كل الأفعال', a.length], ['rule', 'بقاعدة', count((r) => !r.ownerOnly && !(r.anyone && !r.roles?.length && !r.perms?.length))],
          ['anyone', 'أي موظف', count((r) => !!r.anyone && !r.roles?.length && !r.perms?.length)], ['owner', 'المالك بس', count((r) => !!r.ownerOnly)],
          ['unused', 'بلا مسار', count((r) => !!r.unused)]] as const).map(([k, l, n]) => (
          <button key={k} type="button" onClick={() => setKind(k)} aria-pressed={kind === k}
            className={`rounded-xl p-3 text-start ring-1 ${kind === k ? 'bg-brand-700 text-white ring-brand-700' : 'bg-white text-slate-800 ring-slate-200'}`}>
            <div className="text-2xl font-extrabold tabular-nums">{n}</div><div className="text-xs">{l}</div>
          </button>
        ))}
      </div>
      <input id="core-matrix-q" value={q} onChange={(e) => setQ(e.target.value)} placeholder="دوّر: فعل، مسار، صلاحية، دور…"
        className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm" />
      <p className="text-xs text-slate-500">{rows.length} فعل</p>
      <div className="overflow-x-auto rounded-xl bg-white ring-1 ring-slate-200">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="p-2 text-start">الفعل</th><th className="p-2 text-start">منو يگدر</th><th className="p-2 text-start">المسار</th></tr></thead>
          <tbody>
            {rows.slice(0, 400).map((r) => {
              const w = who(r)
              return (
                <tr key={r.action} className="border-t border-slate-100 align-top">
                  <td className="p-2"><div className="font-bold text-slate-800">{r.label}</div><code className="text-[11px] text-slate-500" dir="ltr">{r.action}</code></td>
                  <td className="p-2"><span className={`inline-block rounded-lg px-2 py-0.5 text-xs font-bold ${w.tone}`}>{w.text}</span></td>
                  <td className="p-2"><code className="block text-[11px] text-slate-600" dir="ltr">{r.routes?.join(' · ') || '—'}</code></td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {rows.length > 400 && <p className="p-2 text-center text-xs text-slate-500">أول ٤٠٠ — ضيّق البحث</p>}
      </div>
    </div>
  )
}
