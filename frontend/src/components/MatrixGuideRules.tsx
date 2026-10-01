import { useEffect, useState } from 'react'
import { api, type MatrixGuideRule } from '../api'
import { GROUP_LABEL, type EyeGroup } from './matrixEyeColors'

// ═══ تعليمات ماتركس — «الخطوات التعليمية» ═══
// المدير يكتب شنو يگول ماتركس بكل شاشة/زر. وإذا مفتاح هايكو موجود، هايكو
// ياخذ التعليمة أساساً ويصيغ التوجيه حسب موقف الموظف وأرقامه.

const EMPTY: Partial<MatrixGuideRule> = { route: '/', match: '', groups: '', text: '', onlyIfPending: false, priority: 5, enabled: true }

export default function MatrixGuideRules() {
  const [rules, setRules] = useState<MatrixGuideRule[] | null>(null)
  const [modelOn, setModelOn] = useState(false)
  const [form, setForm] = useState<Partial<MatrixGuideRule>>(EMPTY)
  const [editId, setEditId] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    api.getGuideRules().then((r) => { if (alive) { setRules(r.rules); setModelOn(r.modelEnabled) } }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [reload])

  const save = async () => {
    setErr(null)
    try {
      if (editId) await api.updateGuideRule(editId, form)
      else await api.createGuideRule(form)
      setForm(EMPTY); setEditId(null); setReload((n) => n + 1)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
  }
  const toggle = async (r: MatrixGuideRule) => {
    try { await api.updateGuideRule(r.id, { ...r, enabled: !r.enabled }); setReload((n) => n + 1) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }
  const del = async (r: MatrixGuideRule) => {
    if (!confirm('تحذف هاي التعليمة؟')) return
    try { await api.deleteGuideRule(r.id); setReload((n) => n + 1) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }

  const inp = 'w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-extrabold text-[#0f2040]">📚 تعليمات ماتركس</h3>
        <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${modelOn ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}>
          {modelOn ? '🧠 هايكو شغّال — يصيغ التوجيه بنفسه' : '📏 بالتعليمات حرفياً — هايكو ينتظر المفتاح'}
        </span>
      </div>
      <p className="mb-3 text-xs text-slate-500">
        لكل شاشة وزر: شنو يگول ماتركس للموظف. <code>{'{left}'}</code> = عدد الشغل الباقي، <code>{'{work}'}</code> = تفاصيله. «بس إذا عنده شغل باقي» = ما يطلع إلا لو متأخر.
      </p>
      {err && <p className="mb-2 rounded-lg bg-red-50 p-2 text-red-600">{err}</p>}

      <div className="mb-4 grid gap-2 rounded-xl bg-slate-50 p-3 md:grid-cols-4">
        <label className="text-xs">الشاشة (المسار)<input className={inp} dir="ltr" value={form.route ?? ''} onChange={(e) => setForm({ ...form, route: e.target.value })} placeholder="/coordinator" /></label>
        <label className="text-xs">كلمات الزر (فاصل |)<input className={inp} value={form.match ?? ''} onChange={(e) => setForm({ ...form, match: e.target.value })} placeholder="تثبيت|تأكيد" /></label>
        <label className="text-xs">الأدوار
          <select className={inp} value={form.groups ?? ''} onChange={(e) => setForm({ ...form, groups: e.target.value })}>
            <option value="">الكل</option>
            {(Object.keys(GROUP_LABEL) as EyeGroup[]).map((g) => <option key={g} value={g}>{GROUP_LABEL[g]}</option>)}
          </select>
        </label>
        <label className="text-xs">الأولوية<input type="number" className={inp} value={form.priority ?? 5} onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })} /></label>
        <label className="text-xs md:col-span-4">التوجيه<textarea className={inp} rows={2} value={form.text ?? ''} onChange={(e) => setForm({ ...form, text: e.target.value })} placeholder="تواصل ويا الزبون أول، وإذا توصلت اضغط «تم»." /></label>
        <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={!!form.onlyIfPending} onChange={(e) => setForm({ ...form, onlyIfPending: e.target.checked })} /> بس إذا عنده شغل باقي</label>
        <div className="flex gap-2 md:col-span-3 md:justify-end">
          {editId && <button onClick={() => { setForm(EMPTY); setEditId(null) }} className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs">إلغاء</button>}
          <button onClick={save} className="rounded-lg bg-brand-600 px-4 py-1.5 text-xs font-bold text-white">{editId ? 'حفظ التعديل' : '➕ إضافة تعليمة'}</button>
        </div>
      </div>

      {!rules ? <p className="text-slate-400">جاري التحميل…</p> : (
        <ul className="divide-y divide-slate-100">
          {rules.map((r) => (
            <li key={r.id} className={`flex flex-wrap items-start justify-between gap-2 py-2 ${r.enabled ? '' : 'opacity-50'}`}>
              <div className="min-w-0 flex-1">
                <p className="text-slate-800">«{r.text}»</p>
                <p className="text-[11px] text-slate-500" dir="rtl">
                  <code dir="ltr">{r.route}</code>{r.match && <> · زر: {r.match}</>} · {r.groups ? GROUP_LABEL[r.groups as EyeGroup] ?? r.groups : 'كل الأدوار'}{r.onlyIfPending && ' · بس إذا متأخر'} · أولوية {r.priority}
                </p>
              </div>
              <div className="flex gap-1">
                <button onClick={() => { setForm(r); setEditId(r.id) }} className="rounded-md border border-slate-200 px-2 py-0.5 text-xs">✏️</button>
                <button onClick={() => toggle(r)} className="rounded-md border border-slate-200 px-2 py-0.5 text-xs">{r.enabled ? '⏸️' : '▶️'}</button>
                <button onClick={() => del(r)} className="rounded-md border border-red-200 px-2 py-0.5 text-xs text-red-600">🗑️</button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
