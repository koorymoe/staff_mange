import { useEffect, useState } from 'react'
import { api } from '../api'
import PageHeader from '../components/PageHeader'
import { LineIcon } from '../components/LineIcon'
import { useSession, isAdminOrOwner } from '../session'
import { dayBaghdad } from '../utils/format'
import { askInput } from '../ui/dialogs'
import { swallow } from '../utils/swallow'

import { notify } from '../ui/notify'
// ═══ علاقات الموردين — قرار (ع) 10-15 ═══
// العلاقات العامة ووحدة التقنيين: كل مورد (مشتريات أو تقنيين)، صارلنا شكد ما
// تواصلنا وياه، وآخر حركة (تواصل أو شراء) شنو ومتى ومنو. وتذكير يومي للي عبر الحد.

interface Row {
  kind: 'PROC' | 'TECH'; id: string; name: string; phone: string | null; specialty: string | null
  lastContactAt: string | null; lastContact: string | null; lastContactBy: string | null
  lastBuyAt: string | null; lastBuy: string | null; contacts90: number; buys90: number; daysSince: number | null; overdue: boolean
}
interface Ev { at: string; type: string; text: string | null; by: string | null }

const when = (s: string | null) => dayBaghdad(s) || '—'

export default function SupplierRelationsPage() {
  const { employee } = useSession()
  const isAdmin = isAdminOrOwner(employee)
  const [data, setData] = useState<{ rows: Row[]; contactDays: number; kinds: Record<string, string> } | null>(null)
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  const [kind, setKind] = useState<'' | 'PROC' | 'TECH'>('')
  const [onlyOverdue, setOnlyOverdue] = useState(false)
  const [q, setQ] = useState('')
  const [open, setOpen] = useState<Row | null>(null)
  const [tl, setTl] = useState<{ key: string; evs: Ev[] } | null>(null)
  const [cType, setCType] = useState('CALL')
  const [cNote, setCNote] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    let alive = true
    api.supplierRelations().then((d) => { if (alive) setData(d as typeof data & object) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [reload])

  const openKey = open ? `${open.kind}:${open.id}` : ''
  useEffect(() => {
    if (!open) return
    let alive = true
    api.supplierTimeline(open.kind, open.id).then((evs) => { if (alive) setTl({ key: `${open.kind}:${open.id}`, evs }) }).catch(swallow)
    return () => { alive = false }
  }, [open, reload])
  const evs = tl?.key === openKey ? tl.evs : null

  const save = async () => {
    if (!open) return
    setBusy(true)
    try { await api.addSupplierContact(open.kind, open.id, cType, cNote); setCNote(''); setReload((n) => n + 1) }
    catch (e) { notify(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const editDays = async () => {
    const v = (await askInput('بعد كم يوم بلا تواصل يطلع تذكير؟', String(data?.contactDays ?? 30)))
    if (!v || !(Number(v) > 0)) return
    try { await api.setSupplierContactDays(Number(v)); setReload((n) => n + 1) } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') }
  }

  const rows = (data?.rows ?? [])
    .filter((r) => !kind || r.kind === kind)
    .filter((r) => !onlyOverdue || r.overdue)
    .filter((r) => !q.trim() || `${r.name} ${r.phone ?? ''} ${r.specialty ?? ''}`.includes(q.trim()))
    .sort((a, b) => (b.daysSince ?? 99999) - (a.daysSince ?? 99999))
  const overdueN = (data?.rows ?? []).filter((r) => r.overdue).length
  const typeLabel = (t: string) => (t === 'PURCHASE' ? '🛒 شراء' : data?.kinds[t] ?? t)

  return (
    <div dir="rtl" className="space-y-4">
      <PageHeader title="🤝 علاقات الموردين" subtitle="موردين المشتريات والتقنيين — صارلنا شكد ما تواصلنا، وآخر الحركات شنو ومتى ومنو" />
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}
      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
        {([['', 'الكل'], ['PROC', '🏬 موردين المشتريات'], ['TECH', '🏪 موردين التقنيين']] as const).map(([k, l]) => (
          <button key={k} onClick={() => setKind(k)} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${kind === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>{l}</button>
        ))}
        <button onClick={() => setOnlyOverdue((v) => !v)} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${onlyOverdue ? 'bg-red-600 text-white' : 'bg-red-50 text-red-700'}`}>⏰ صارلهم أكثر من {data?.contactDays ?? 30} يوم ({overdueN})</button>
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث: اسم، رقم، تخصص" className="min-w-[180px] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
        {isAdmin && <button onClick={editDays} className="text-xs text-brand-700 underline">تغيير حد التذكير</button>}
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        {/* (ك 11-23) الجدول حسب صورة زينا: رأس بأيقونات، وكل مورد بطاقة — الليلي مقاس من الصورة (.sp-*) */}
        <div className="sp-wrap overflow-x-auto xl:col-span-2">
          <div className="min-w-[640px]">
            <div className="sp-head sp-grid">
              {([['store', 'المورد'], ['tag', 'النوع'], ['pulse', 'آخر حركة'], ['phone', 'آخر تواصل'], ['cart', 'آخر شراء'], ['cal', '٩٠ يوم']] as const).map(([i, h]) => (
                <span key={h} className="sp-th"><LineIcon n={i} className="h-4 w-4" />{h}</span>
              ))}
            </div>
            <div className="mt-2 space-y-2">
              {rows.map((r) => (
                <div key={r.kind + r.id} role="button" tabIndex={0} onClick={() => setOpen(r)} onKeyDown={(e) => { if (e.key === 'Enter') setOpen(r) }}
                  className={`sp-row sp-grid ${open?.id === r.id && open.kind === r.kind ? 'sp-row-on' : ''}`}>
                  <span className="flex items-center gap-3">
                    <span className="sp-ibox"><LineIcon n="briefcase" className="h-5 w-5" /></span>
                    <span><b className="sp-name">{r.name}</b><span className="sp-phone block tabular-nums">{r.phone}{r.specialty ? ` · ${r.specialty}` : ''}</span></span>
                  </span>
                  <span className="sp-type"><LineIcon n={r.kind === 'TECH' ? 'store' : 'cube'} className="h-5 w-5 sp-type-ico" />{r.kind === 'TECH' ? 'تقنيين' : 'مشتريات'}</span>
                  <span>
                    <span className={`sp-pill ${r.overdue || r.daysSince == null ? 'sp-pill-bad' : 'sp-pill-ok'}`}>
                      <LineIcon n={r.daysSince == null ? 'ban' : r.overdue ? 'clock' : 'checkCircle'} className="h-4 w-4 shrink-0" />
                      <span>{r.daysSince == null ? <>ماكو<br />ولا حركة</> : r.daysSince === 0 ? 'اليوم' : `قبل ${r.daysSince} يوم`}</span>
                    </span>
                  </span>
                  <span className="sp-cell">{r.lastContactAt ? <>{when(r.lastContactAt)} · {r.lastContact}{r.lastContactBy && <span className="sp-muted"> ({r.lastContactBy})</span>}</> : '—'}</span>
                  <span className="sp-cell">{r.lastBuyAt ? `${when(r.lastBuyAt)} · ${r.lastBuy}` : '—'}</span>
                  <span className="sp-90">
                    <span className="inline-flex items-center gap-2"><LineIcon n="chat" className="h-4 w-4" />{r.contacts90} تواصل</span>
                    <span className="inline-flex items-center gap-2"><LineIcon n="cart" className="h-4 w-4" />{r.buys90} شراء</span>
                  </span>
                </div>
              ))}
              {data && rows.length === 0 && <p className="sp-cell p-4 text-center">ماكو موردين بهذا الفلتر</p>}
            </div>
          </div>
        </div>

        <div className="space-y-3">
          {!open && <p className="rounded-2xl bg-white p-4 text-sm text-slate-400 shadow-sm">اختار مورد حتى تشوف حركاته وتسجّل تواصل.</p>}
          {open && (
            <div className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
              <div className="flex items-center justify-between"><p className="font-bold">{open.name}</p><button onClick={() => setOpen(null)} className="text-xs text-slate-400">✕</button></div>
              <div className="space-y-2 rounded-lg border border-slate-200 p-2">
                <p className="text-xs font-bold text-slate-500">سجّل تواصل</p>
                <div className="flex flex-wrap gap-1">
                  {Object.entries(data?.kinds ?? {}).map(([k, l]) => (
                    <button key={k} onClick={() => setCType(k)} className={`rounded-lg px-2 py-1 text-xs font-bold ${cType === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100'}`}>{l}</button>
                  ))}
                </div>
                <textarea value={cNote} onChange={(e) => setCNote(e.target.value)} rows={2} placeholder="شنو صار؟ (عروض جديدة، أسعار، مشكلة…)" className="w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
                <button disabled={busy} onClick={save} className="w-full rounded-lg bg-[#0f2040] py-1.5 text-sm font-bold text-white disabled:opacity-50">حفظ</button>
              </div>
              <p className="text-xs font-bold text-slate-500">آخر الحركات</p>
              {!evs && <p className="text-xs text-slate-400">جاري الجلب...</p>}
              {evs && evs.length === 0 && <p className="text-xs text-slate-400">ماكو ولا حركة مسجّلة وياه.</p>}
              <div className="max-h-[420px] space-y-1 overflow-auto">
                {evs?.map((e, i) => (
                  <div key={i} className="border-r-2 border-brand-200 pr-2 text-sm">
                    <p className="text-xs text-slate-400">{new Date(e.at).toLocaleString('en-GB', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })} · {e.by ?? '—'}</p>
                    <p><b>{typeLabel(e.type)}</b>{e.text ? ` — ${e.text}` : ''}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
