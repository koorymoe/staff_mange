import { useEffect, useState } from 'react'
import { api, type CoreTable, type CoreTablePage } from '../api'
import { useSession } from '../session'

// ═══ 🗄️ لوحة الأدمن التلقائية — ون بيس (11-10) ═══
// مثل أدمن جانغو: كل جدول بالقاعدة يطلعله تصفّح وبحث بلا كود خاص.
// قراءة بس — التعديل بشاشات النظام الي بيها قواعدها. الأسرار تطلع «•••».

export default function CoreAdminPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [tables, setTables] = useState<CoreTable[] | null>(null)
  const [filter, setFilter] = useState('')
  const [table, setTable] = useState('')
  const [page, setPage] = useState(1)
  const [q, setQ] = useState('')
  const [search, setSearch] = useState('')
  const [data, setData] = useState<CoreTablePage | null>(null)
  const [error, setError] = useState('')
  useEffect(() => {
    if (!isOwner) return
    let alive = true
    api.getCoreTables().then((t) => { if (alive) setTables(t) }).catch((e: Error) => { if (alive) setError(e.message) })
    return () => { alive = false }
  }, [isOwner])
  useEffect(() => {
    if (!table) return
    let alive = true
    api.getCoreTable(table, page, search).then((d) => { if (alive) { setData(d); setError('') } }).catch((e: Error) => { if (alive) setError(e.message) })
    return () => { alive = false }
  }, [table, page, search])
  if (!isOwner) return <p className="p-6 text-center text-slate-500">الصفحة غير موجودة</p>
  const shown = (tables ?? []).filter((t) => t.name.toLowerCase().includes(filter.trim().toLowerCase()))
  const pages = data ? Math.max(1, Math.ceil(data.total / data.size)) : 1
  const cell = (v: unknown) => v == null ? <span className="text-slate-300">—</span> : typeof v === 'object' ? JSON.stringify(v).slice(0, 120) : String(v).slice(0, 120)
  return (
    <div className="space-y-4 p-4" dir="rtl">
      <header className="space-y-1">
        <h1 className="text-xl font-extrabold text-slate-900">🗄️ لوحة الأدمن</h1>
        <p className="text-sm text-slate-600">كل جداول قاعدة البيانات — تصفّح وبحث، قراءة بس. كلمات السر والتوكنات مخفية دائماً.</p>
      </header>
      {error && <p role="alert" className="rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>}
      <div className="grid gap-4 lg:grid-cols-[260px_1fr]">
        <aside className="space-y-2">
          <input id="core-admin-filter" value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="دوّر جدول…"
            className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm" dir="ltr" />
          <ul className="max-h-[70vh] overflow-y-auto rounded-xl bg-white ring-1 ring-slate-200">
            {!tables && <li className="p-3 text-sm text-slate-500">جاري التحميل…</li>}
            {shown.map((t) => (
              <li key={t.name}>
                <button type="button" onClick={() => { setTable(t.name); setPage(1); setQ(''); setSearch(''); setData(null) }}
                  className={`flex w-full items-center justify-between gap-2 px-3 py-2 text-start text-sm ${table === t.name ? 'bg-brand-50 font-bold text-brand-800' : 'text-slate-700 hover:bg-slate-50'}`}>
                  <span dir="ltr" className="truncate">{t.name}</span><span className="text-xs tabular-nums text-slate-400">{t.rows}</span>
                </button>
              </li>
            ))}
          </ul>
        </aside>
        <section className="min-w-0 space-y-2">
          {!table && <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-500 ring-1 ring-slate-200">اختار جدول من القائمة</p>}
          {table && (
            <>
              <form className="flex gap-2" onSubmit={(e) => { e.preventDefault(); setPage(1); setSearch(q) }}>
                <input id="core-admin-q" value={q} onChange={(e) => setQ(e.target.value)} placeholder={`دوّر بـ${table}…`}
                  className="min-w-0 flex-1 rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm" />
                <button type="submit" className="rounded-xl bg-brand-700 px-4 text-sm font-bold text-white">بحث</button>
              </form>
              {data && (
                <>
                  <p className="text-xs text-slate-500"><span dir="ltr">{data.table}</span> · {data.total} صف · صفحة {data.page} من {pages}</p>
                  <div className="overflow-x-auto rounded-xl bg-white ring-1 ring-slate-200">
                    <table className="text-xs">
                      <thead className="bg-slate-50 text-slate-500"><tr>{data.columns.map((c) => <th key={c.name} className="whitespace-nowrap p-2 text-start" dir="ltr" title={c.type}>{c.name}</th>)}</tr></thead>
                      <tbody>
                        {data.rows.map((r, i) => (
                          <tr key={i} className="border-t border-slate-100">{data.columns.map((c) => <td key={c.name} className="max-w-[260px] truncate whitespace-nowrap p-2 text-slate-700">{cell(r[c.name])}</td>)}</tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  <div className="flex items-center justify-center gap-2">
                    <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="rounded-lg px-3 py-1 text-sm ring-1 ring-slate-300 disabled:opacity-40">السابق</button>
                    <span className="text-sm tabular-nums">{page} / {pages}</span>
                    <button type="button" disabled={page >= pages} onClick={() => setPage((p) => p + 1)} className="rounded-lg px-3 py-1 text-sm ring-1 ring-slate-300 disabled:opacity-40">التالي</button>
                  </div>
                </>
              )}
            </>
          )}
        </section>
      </div>
    </div>
  )
}
