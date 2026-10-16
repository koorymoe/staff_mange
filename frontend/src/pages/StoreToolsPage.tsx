import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type OnDemandTool, type ToolRequestItem } from '../api'

// ═══ أدوات المخزن التخصصية — قرار (ع) 10-16 ═══
// «أبو الكميات يضيف بيها الأدوات التخصصية وعددهن، ومن يصير نقص عند موظف
// ينطي من عدهن». نفس جدول «أدوات حسب الحاجة» (OnDemandTool) — مو مخزن
// ثاني — حتى الكمية الي تطلع هنا هي نفسها الي بطلبات الأدوات وبالجرد.
export default function StoreToolsPage() {
  const [tools, setTools] = useState<OnDemandTool[]>([])
  const [issued, setIssued] = useState<ToolRequestItem[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [q, setQ] = useState('')
  // إضافة أداة جديدة
  const [newName, setNewName] = useState('')
  const [newQty, setNewQty] = useState('1')
  // زيادة كمية / إعطاء
  const [addFor, setAddFor] = useState<OnDemandTool | null>(null)
  const [addQty, setAddQty] = useState('1')
  const [giveFor, setGiveFor] = useState<OnDemandTool | null>(null)
  const [giveTo, setGiveTo] = useState('')
  const [giveNote, setGiveNote] = useState('')

  const load = () => {
    api.getOnDemandTools().then(setTools).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب الأدوات'))
    api.getToolRequests().then((r) => setIssued(r.filter((x) => x.status === 'APPROVED'))).catch(() => {})
  }
  useEffect(() => {
    load()
    api.getEmployees().then((e) => setEmployees(e.filter((x) => x.status === 'ACTIVE'))).catch(() => {})
  }, [])

  const run = async (fn: () => Promise<unknown>, ok?: string) => {
    setBusy(true); setErr('')
    try { await fn(); load(); if (ok) alert(ok) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }

  const shown = tools.filter((t) => !q.trim() || t.name.includes(q.trim()))
  const holders = useMemo(() => {
    const m = new Map<string, ToolRequestItem[]>()
    for (const r of issued) m.set(r.toolId, [...(m.get(r.toolId) ?? []), r])
    return m
  }, [issued])

  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">🧰 أدوات المخزن التخصصية</h1>
        <p className="mt-1 text-sm text-blue-100">ضيف الأدوات التخصصية وعددها. من ينقص شي عند موظف، انطيه من هنا — العدد ينقص لحاله، ومن يرجّعها يرجع العدد.</p>
      </div>

      {err && <p className="rounded-xl bg-red-50 px-4 py-2 text-sm text-red-700">{err}</p>}

      <div className="flex flex-wrap items-end gap-2 rounded-2xl border border-slate-200 bg-white p-4">
        <div className="flex-1">
          <label className="mb-1 block text-xs font-bold text-slate-500">أداة جديدة</label>
          <input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="مثلاً: جهاز فحص كيبل، كمّاشة RJ45…"
            className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div className="w-28">
          <label className="mb-1 block text-xs font-bold text-slate-500">العدد</label>
          <input type="number" min={1} value={newQty} onChange={(e) => setNewQty(e.target.value)} className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <button disabled={busy || newName.trim().length < 2 || !(Number(newQty) > 0)}
          onClick={() => run(() => api.createOnDemandTool({ name: newName.trim(), barcode: `ST-${Date.now()}`, totalQuantity: Math.round(Number(newQty)) }), undefined).then(() => { setNewName(''); setNewQty('1') })}
          className="rounded-xl bg-[#0f2040] px-5 py-2 text-sm font-bold text-white disabled:opacity-40">➕ ضيف</button>
      </div>

      <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث باسم الأداة…" className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />

      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full text-right text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-3">الأداة</th><th className="p-3">الكل</th><th className="p-3">بالمخزن هسه</th><th className="p-3">عند الموظفين</th><th className="p-3">إجراء</th></tr>
          </thead>
          <tbody>
            {shown.map((t) => {
              const h = holders.get(t.id) ?? []
              return (
                <tr key={t.id} className="border-t border-slate-100 align-top">
                  <td className="p-3 font-bold text-[#0f2040]">{t.name}</td>
                  <td className="p-3 tabular-nums">{t.totalQuantity}</td>
                  <td className={`p-3 font-black tabular-nums ${t.availableQuantity > 0 ? 'text-emerald-700' : 'text-red-600'}`}>{t.availableQuantity}</td>
                  <td className="p-3">
                    {h.length === 0 ? <span className="text-slate-400">—</span> : h.map((r) => (
                      <span key={r.id} className="mb-1 ml-1 inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-bold text-amber-800">
                        {r.employee?.name ?? '—'}
                        <button disabled={busy} title="رجّعها للمخزن" onClick={() => confirm(`${r.employee?.name ?? 'الموظف'} رجّع «${t.name}»؟`) && run(() => api.returnStoreTool(r.id))}
                          className="rounded-full bg-white px-1.5 text-[10px] text-slate-600 hover:bg-emerald-100">↩ رجّع</button>
                      </span>
                    ))}
                  </td>
                  <td className="p-3 whitespace-nowrap">
                    <button disabled={busy || t.availableQuantity <= 0} onClick={() => { setGiveFor(t); setGiveTo(''); setGiveNote('') }}
                      className="ml-1 rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-40">🤝 انطي موظف</button>
                    <button disabled={busy} onClick={() => { setAddFor(t); setAddQty('1') }}
                      className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-700">➕ زيد العدد</button>
                  </td>
                </tr>
              )
            })}
            {shown.length === 0 && <tr><td colSpan={5} className="p-8 text-center text-slate-400">ماكو أدوات بعد — ضيف أول أداة من فوك.</td></tr>}
          </tbody>
        </table>
      </div>

      {giveFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setGiveFor(null)}>
          <div className="w-full max-w-md rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[#0f2040]">🤝 انطي «{giveFor.name}»</h3>
            <p className="mt-1 text-xs text-slate-500">بالمخزن {giveFor.availableQuantity} — راح ينقص واحد.</p>
            <select value={giveTo} onChange={(e) => setGiveTo(e.target.value)} className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm">
              <option value="">— اختار الموظف —</option>
              {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
            </select>
            <input value={giveNote} onChange={(e) => setGiveNote(e.target.value)} placeholder="ملاحظة (اختياري): ليش ناقصته…" className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
            <div className="mt-4 flex gap-2">
              <button disabled={busy || !giveTo} onClick={() => run(() => api.issueStoreTool(giveFor.id, giveTo, giveNote || undefined)).then(() => setGiveFor(null))}
                className="flex-1 rounded-xl bg-[#0f2040] py-2 text-sm font-bold text-white disabled:opacity-40">سلّمه</button>
              <button onClick={() => setGiveFor(null)} className="rounded-xl border border-slate-300 px-4 text-sm">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {addFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setAddFor(null)}>
          <div className="w-full max-w-sm rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[#0f2040]">➕ زيد عدد «{addFor.name}»</h3>
            <input type="number" min={1} value={addQty} onChange={(e) => setAddQty(e.target.value)} className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
            <div className="mt-4 flex gap-2">
              <button disabled={busy || !(Number(addQty) > 0)} onClick={() => run(() => api.addStockIntake({ toolId: addFor.id, quantity: Math.round(Number(addQty)) })).then(() => setAddFor(null))}
                className="flex-1 rounded-xl bg-[#0f2040] py-2 text-sm font-bold text-white disabled:opacity-40">ضيف للمخزن</button>
              <button onClick={() => setAddFor(null)} className="rounded-xl border border-slate-300 px-4 text-sm">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
