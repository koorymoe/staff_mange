import { useEffect, useState } from 'react'
import { api, type CoreEditLog, type CoreTable, type CoreTablePage } from '../api'
import { useSession } from '../session'

// ═══ 🗄️ لوحة الأدمن التلقائية — ون بيس (11-10) ═══
// مثل أدمن جانغو: كل جدول بالقاعدة يطلعله تصفّح وبحث بلا كود خاص.
// تصفّح للمالك (أو مدير نطاه المالك 👑). التعديل والإضافة بإنطاء منفصل، وكل
// تغيير ينسجل (قبل/بعد). ماكو حذف، والسجلات ما تتعدّل، والفلوس للمالك بس.

export default function CoreAdminPage() {
  const { employee, permissions } = useSession()
  // المالك، أو مدير نظام نطاه المالك هالشاشة (👑 صلاحيات المالك)
  const owner = employee?.actualRole === 'OWNER'
  const isOwner = owner || permissions.includes('owner:system.admin_tables')
  const canEdit = owner || permissions.includes('owner:system.admin_edit')
  const canCreate = owner || permissions.includes('owner:system.admin_create')
  const canLog = owner || permissions.includes('owner:system.admin_log')
  // الصف المفتوح للتعديل (أو null للإضافة)، والقيم بالنموذج
  const [form, setForm] = useState<{ row: Record<string, unknown> | null; values: Record<string, string> } | null>(null)
  const [saving, setSaving] = useState(false)
  const [formErr, setFormErr] = useState('')
  const [refresh, setRefresh] = useState(0)
  const [log, setLog] = useState<CoreEditLog[] | null>(null)
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
  }, [table, page, search, refresh])
  const asText = (v: unknown) => v == null ? '' : typeof v === 'object' ? JSON.stringify(v) : String(v)
  const openEdit = (row: Record<string, unknown> | null) => {
    if (!data) return
    const values: Record<string, string> = {}
    for (const c of data.columns) values[c.name] = row ? asText(row[c.name]) : ''
    setForm({ row, values }); setFormErr('')
  }
  const save = () => {
    if (!form || !data) return
    const changed: Record<string, unknown> = {}
    for (const c of data.columns) {
      if (data.locked.includes(c.name) || (form.row && data.pk.includes(c.name))) continue
      const now = form.values[c.name]
      const was = form.row ? asText(form.row[c.name]) : ''
      if (now !== was) changed[c.name] = now === '' ? null : now
    }
    if (Object.keys(changed).length === 0) { setFormErr('ماكو تغيير'); return }
    setSaving(true); setFormErr('')
    const key: Record<string, unknown> = {}
    if (form.row) for (const k of data.pk) key[k] = form.row[k]
    ;(form.row ? api.coreEditRow(data.table, key, changed) : api.coreCreateRow(data.table, changed))
      .then(() => { setForm(null); setRefresh((n) => n + 1) })
      .catch((e: Error) => setFormErr(e.message))
      .finally(() => setSaving(false))
  }
  if (!isOwner) return <p className="p-6 text-center text-slate-500">الصفحة غير موجودة</p>
  const shown = (tables ?? []).filter((t) => t.name.toLowerCase().includes(filter.trim().toLowerCase()))
  const pages = data ? Math.max(1, Math.ceil(data.total / data.size)) : 1
  const cell = (v: unknown) => v == null ? <span className="text-slate-300">—</span> : typeof v === 'object' ? JSON.stringify(v).slice(0, 120) : String(v).slice(0, 120)
  return (
    <div className="space-y-4 p-4" dir="rtl">
      <header className="space-y-1">
        <h1 className="text-xl font-extrabold text-slate-900">🗄️ لوحة الأدمن</h1>
        <p className="text-sm text-slate-600">كل جداول قاعدة البيانات — تصفّح وبحث{canEdit ? '، وتعديل' : ''}{canCreate ? ' وإضافة' : ''}. ماكو حذف، والسجلات ما تتعدّل، وكلمات السر والتوكنات مخفية دائماً.</p>
        {canLog && (
          <button type="button" onClick={() => (log ? setLog(null) : api.getCoreEditLog().then(setLog).catch((e: Error) => setError(e.message)))}
            className="rounded-lg px-3 py-1 text-xs ring-1 ring-slate-300">{log ? 'سكّر السجل' : '🧾 سجل التعديلات'}</button>
        )}
      </header>
      {log && (
        <ul className="max-h-72 space-y-1 overflow-y-auto rounded-xl bg-white p-2 text-xs ring-1 ring-slate-200">
          {log.length === 0 && <li className="p-2 text-slate-500">ماكو تعديلات بعد.</li>}
          {log.map((l, i) => {
            const keys = Object.keys(l.after ?? {}).filter((k) => JSON.stringify((l.before ?? {})[k]) !== JSON.stringify((l.after ?? {})[k]))
            return (
              <li key={i} className="border-b border-slate-100 p-1 last:border-0">
                <span className="tabular-nums text-slate-500">{l.at}</span> · <b>{l.by}</b> {l.action === 'create' ? 'ضاف صف بـ' : 'عدّل'} <code dir="ltr">{l.table}</code>
                {l.action === 'update' && <span className="text-slate-600"> — {keys.map((k) => `${k}: ${asText((l.before ?? {})[k]) || '—'} ← ${asText((l.after ?? {})[k]) || '—'}`).join(' · ')}</span>}
              </li>
            )
          })}
        </ul>
      )}
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
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-xs text-slate-500"><span dir="ltr">{data.table}</span> · {data.total} صف · صفحة {data.page} من {pages}</p>
                    {!data.editable && <span className="rounded-lg bg-amber-50 px-2 py-0.5 text-xs font-bold text-amber-800">🔒 {data.readonlyReason}</span>}
                    {data.editable && canCreate && <button type="button" onClick={() => openEdit(null)} className="ms-auto rounded-lg bg-emerald-600 px-3 py-1 text-xs font-bold text-white">＋ صف جديد</button>}
                  </div>
                  {form && (
                    <div className="space-y-2 rounded-xl bg-white p-3 ring-2 ring-brand-300">
                      <p className="text-sm font-extrabold text-slate-800">{form.row ? 'تعديل صف' : 'صف جديد'} <span className="text-xs font-normal text-slate-500">— ينسجل بسجل التعديلات</span></p>
                      <div className="grid max-h-[50vh] gap-2 overflow-y-auto sm:grid-cols-2">
                        {data.columns.map((c) => {
                          const off = data.locked.includes(c.name) || (!!form.row && data.pk.includes(c.name))
                          return (
                            <label key={c.name} className="block text-xs">
                              <span className="text-slate-500" dir="ltr">{c.name} <span className="opacity-60">({c.type})</span>{off ? ' 🔒' : ''}</span>
                              <input id={`edit-${c.name}`} disabled={off} value={off && data.locked.includes(c.name) ? '•••' : form.values[c.name] ?? ''}
                                onChange={(e) => setForm((f) => f && { ...f, values: { ...f.values, [c.name]: e.target.value } })}
                                className="mt-0.5 w-full rounded-lg border border-slate-300 px-2 py-1 text-sm disabled:bg-slate-100" />
                            </label>
                          )
                        })}
                      </div>
                      {formErr && <p role="alert" className="text-sm font-bold text-red-700">{formErr}</p>}
                      <div className="flex gap-2">
                        <button type="button" disabled={saving} onClick={save} className="rounded-xl bg-brand-700 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">{saving ? '…' : 'احفظ'}</button>
                        <button type="button" onClick={() => setForm(null)} className="rounded-xl px-5 py-2 text-sm ring-1 ring-slate-300">إلغاء</button>
                      </div>
                    </div>
                  )}
                  <div className="overflow-x-auto rounded-xl bg-white ring-1 ring-slate-200">
                    <table className="text-xs">
                      <thead className="bg-slate-50 text-slate-500"><tr>{data.editable && canEdit && <th />}{data.columns.map((c) => <th key={c.name} className="whitespace-nowrap p-2 text-start" dir="ltr" title={c.type}>{c.name}</th>)}</tr></thead>
                      <tbody>
                        {data.rows.map((r, i) => (
                          <tr key={i} className="border-t border-slate-100">
                            {data.editable && canEdit && <td className="p-1"><button type="button" onClick={() => openEdit(r)} className="rounded px-2 py-0.5 text-xs text-brand-700 ring-1 ring-brand-200">تعديل</button></td>}
                            {data.columns.map((c) => <td key={c.name} className="max-w-[260px] truncate whitespace-nowrap p-2 text-slate-700">{cell(r[c.name])}</td>)}
                          </tr>
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
