import { useCallback, useEffect, useMemo, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 📞 استفسارات الطاقة الشمسية (دازاي 11-12) ═══
// المسجّل يسجّل الزبون ← التقنيين يتواصلون ويسجّلون النتيجة ← ابو الجودة يتصل بالزبون ويتأكد:
// اتصل بيك التقني؟ چنت بس تستفسر؟ انطاك عرض سعر مناسب؟ والي التقني كاتب عليه «ما رد» هم يتأكد منه.

type Inq = {
  id: string; number: number; customerName: string; phone: string; area: string | null; note: string | null
  createdBy: string | null; createdAt: string; techId: string | null; techName: string | null
  techOutcome: string | null; techNote: string | null; techContactAt: string | null; quoteGiven: boolean | null; quoteAmount: number | null
  qualityBy: string | null; qualityAt: string | null; qCustomerReached: boolean | null; qTechCalled: string | null
  qIntent: string | null; qGotQuote: string | null; qQuoteFair: string | null; qNote: string | null
  stage: 'NEW' | 'TECH_DONE' | 'CLOSED'; overdue: boolean; mismatch: boolean
}
type Access = { intake: boolean; tech: boolean; quality: boolean; admin: boolean }

const OUTCOME: Record<string, string> = { JUST_ASKING: 'بس يستفسر', WANTS_INSTALL: 'يريد ينصب منظومة', NO_ANSWER: 'ما رد', NOT_INTERESTED: 'ما مهتم' }
const YNU: Record<string, string> = { YES: 'إي', NO: 'لا', UNSURE: 'ما يدري' }
const STAGE: Record<string, { t: string; cls: string }> = {
  NEW: { t: 'ينتظر التقني', cls: 'bg-amber-100 text-amber-800' },
  TECH_DONE: { t: 'ينتظر ابو الجودة', cls: 'bg-sky-100 text-sky-800' },
  CLOSED: { t: 'مسكّر', cls: 'bg-emerald-100 text-emerald-700' },
}
const input = 'w-full rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-sm'
const when = (s: string | null) => (s ? new Date(s).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }) : '')

function Choice({ id, label, value, onChange, opts }: { id: string; label: string; value: string; onChange: (v: string) => void; opts: Record<string, string> }) {
  return (
    <fieldset className="text-sm">
      <legend className="mb-1 font-bold text-slate-700">{label}</legend>
      <div className="flex flex-wrap gap-1">
        {Object.entries(opts).map(([k, t]) => (
          <button key={k} id={`${id}-${k}`} type="button" onClick={() => onChange(k)}
            className={`rounded-full px-3 py-1 text-xs font-bold ${value === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-700'}`}>{t}</button>
        ))}
      </div>
    </fieldset>
  )
}

export default function SolarInquiriesPage() {
  const [rows, setRows] = useState<Inq[] | null>(null)
  const [acc, setAcc] = useState<Access | null>(null)
  const [tab, setTab] = useState<'NEW' | 'TECH_DONE' | 'CLOSED' | 'ALL'>('NEW')
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)
  const [busy, setBusy] = useState(false)
  const [form, setForm] = useState({ customerName: '', phone: '', area: '', note: '' })
  const [open, setOpen] = useState<string | null>(null)
  const [tc, setTc] = useState({ outcome: '', note: '', quoteGiven: '', quoteAmount: '' })
  const [qc, setQc] = useState({ reached: '', techCalled: '', intent: '', gotQuote: '', quoteFair: '', note: '' })

  const load = useCallback(() => {
    apiRequest<{ items: Inq[]; access: Access }>('/solar-inquiries')
      .then((r) => { setRows(r.items); setAcc(r.access) })
      .catch((e: Error) => setMsg({ ok: false, t: e.message }))
  }, [])
  useEffect(() => { load() }, [load])

  const run = async (f: () => Promise<unknown>, ok: string) => {
    setBusy(true); setMsg(null)
    try { await f(); setMsg({ ok: true, t: ok }); setOpen(null); load() } catch (e) { setMsg({ ok: false, t: (e as Error).message }) } finally { setBusy(false) }
  }
  const create = () => run(async () => {
    await apiRequest('/solar-inquiries', { method: 'POST', body: JSON.stringify(form) })
    setForm({ customerName: '', phone: '', area: '', note: '' })
  }, 'انسجل الاستفسار وانرسل للتقنيين ✅')
  const claim = (id: string) => run(() => apiRequest(`/solar-inquiries/${id}/claim`, { method: 'POST' }), 'صار عليك — اتصل بالزبون وسجّل النتيجة')
  const contact = (id: string) => run(() => apiRequest(`/solar-inquiries/${id}/contact`, {
    method: 'POST',
    body: JSON.stringify({ outcome: tc.outcome, note: tc.note, quoteGiven: tc.quoteGiven === '' ? null : tc.quoteGiven === 'YES', quoteAmount: tc.quoteAmount ? Number(tc.quoteAmount) : null }),
  }), 'انسجل اتصالك وراح لابو الجودة يتأكد')
  const quality = (id: string) => run(() => apiRequest(`/solar-inquiries/${id}/quality`, {
    method: 'POST',
    body: JSON.stringify({ customerReached: qc.reached === 'YES', techCalled: qc.techCalled, intent: qc.intent, gotQuote: qc.gotQuote, quoteFair: qc.quoteFair, note: qc.note }),
  }), 'انسجلت المتابعة ✅')

  const shown = useMemo(() => (rows ?? []).filter((r) => tab === 'ALL' || r.stage === tab), [rows, tab])
  const count = (s: string) => (rows ?? []).filter((r) => r.stage === s).length
  const mismatches = (rows ?? []).filter((r) => r.mismatch).length

  return (
    <div dir="rtl" className="space-y-4 p-4">
      <div>
        <h1 className="text-2xl font-black text-[#0f2040]">📞 استفسارات الطاقة الشمسية</h1>
        <p className="text-xs text-slate-500">يتسجّل الزبون ← التقني يتواصل وياه ← ابو الجودة يتأكد إن التقني اتصل وقدّم عرض مناسب</p>
      </div>
      {msg && <p role={msg.ok ? 'status' : 'alert'} className={`rounded-xl p-2 text-sm ${msg.ok ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'}`}>{msg.t}</p>}
      {mismatches > 0 && (acc?.admin || acc?.quality) && (
        <p className="rounded-xl bg-red-50 p-2 text-sm font-bold text-red-800">🚩 {mismatches} استفسار: التقني كاتب «اتصلت» والزبون يگول ما اتصل بيه أحد</p>
      )}

      {(acc?.intake || acc?.admin) && (
        <section className="grid gap-2 rounded-2xl bg-white p-4 shadow-sm sm:grid-cols-2">
          <h2 className="text-base font-extrabold sm:col-span-2">➕ تسجيل استفسار جديد</h2>
          <label className="text-xs">اسم الزبون<input id="si-name" className={input} value={form.customerName} onChange={(e) => setForm({ ...form, customerName: e.target.value })} /></label>
          <label className="text-xs">رقم الهاتف<input id="si-phone" dir="ltr" className={input} value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="07xx xxx xxxx" /></label>
          <label className="text-xs">المنطقة<input id="si-area" className={input} value={form.area} onChange={(e) => setForm({ ...form, area: e.target.value })} /></label>
          <label className="text-xs">شنو سأل؟<input id="si-note" className={input} value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} placeholder="مثلاً: يسأل عن منظومة ٥ كيلو للبيت" /></label>
          <div className="sm:col-span-2"><button type="button" disabled={busy || !form.customerName.trim() || form.phone.replace(/\D/g, '').length < 10} onClick={() => void create()} className="rounded-xl bg-orange-500 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">سجّل وارسل للتقنيين</button></div>
        </section>
      )}

      {(acc?.tech || acc?.quality || acc?.admin || (rows?.length ?? 0) > 0) && (
        <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
          <div className="flex flex-wrap gap-2">
            {([['NEW', `ينتظر التقني (${count('NEW')})`], ['TECH_DONE', `ينتظر ابو الجودة (${count('TECH_DONE')})`], ['CLOSED', `مسكّرة (${count('CLOSED')})`], ['ALL', 'الكل']] as const).map(([k, t]) => (
              <button key={k} type="button" onClick={() => setTab(k)} className={`rounded-full px-3 py-1 text-xs font-bold ${tab === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-700'}`}>{t}</button>
            ))}
          </div>
          {rows === null && <p className="text-sm text-slate-400">جاري التحميل…</p>}
          {rows && shown.length === 0 && <p className="rounded-xl bg-slate-50 p-6 text-center text-sm text-slate-400">ماكو شي هنا 👌</p>}
          {shown.map((r) => (
            <div key={r.id} className={`rounded-xl border p-3 ${r.mismatch ? 'border-red-300 bg-red-50/40' : r.overdue ? 'border-amber-300' : 'border-slate-200'}`}>
              <div className="flex flex-wrap items-center gap-2 text-sm">
                <b>#{r.number} {r.customerName}</b>
                <a href={`tel:${r.phone}`} dir="ltr" className="text-brand-700 underline">{r.phone}</a>
                {r.area && <span className="text-slate-500">· {r.area}</span>}
                <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${STAGE[r.stage].cls}`}>{STAGE[r.stage].t}</span>
                {r.overdue && <span className="rounded-full bg-amber-200 px-2 py-0.5 text-xs font-bold text-amber-900">⏰ أكثر من ٢٤ ساعة بلا تواصل</span>}
                {r.mismatch && <span className="rounded-full bg-red-200 px-2 py-0.5 text-xs font-bold text-red-900">🚩 الزبون يگول التقني ما اتصل</span>}
              </div>
              {r.note && <p className="mt-1 text-sm text-slate-700">سأل: {r.note}</p>}
              <p className="text-xs text-slate-500">سجّله {r.createdBy ?? '—'} · {when(r.createdAt)}{r.techName ? ` · التقني: ${r.techName}` : ''}</p>
              {r.techOutcome && (
                <p className="mt-1 rounded-lg bg-sky-50 p-2 text-sm">📱 التقني ({when(r.techContactAt)}): <b>{OUTCOME[r.techOutcome]}</b>{r.techNote ? ` — ${r.techNote}` : ''}
                  {r.quoteGiven != null && <> · عرض سعر: <b>{r.quoteGiven ? `إي${r.quoteAmount ? ` (${r.quoteAmount.toLocaleString('en-US')} د.ع)` : ''}` : 'لا'}</b></>}</p>
              )}
              {r.qualityAt && (
                <p className="mt-1 rounded-lg bg-emerald-50 p-2 text-sm">⭐ ابو الجودة {r.qualityBy} ({when(r.qualityAt)}): {r.qCustomerReached
                  ? <>اتصل بيه التقني؟ <b>{YNU[r.qTechCalled ?? ''] ?? '—'}</b> · نيته: <b>{OUTCOME[r.qIntent ?? ''] ?? '—'}</b> · انطوه عرض؟ <b>{YNU[r.qGotQuote ?? ''] ?? '—'}</b> · العرض مناسب للسوق؟ <b>{YNU[r.qQuoteFair ?? ''] ?? '—'}</b></>
                  : <b>ما وصل للزبون</b>}{r.qNote ? ` — ${r.qNote}` : ''}</p>
              )}

              <div className="mt-2 flex flex-wrap gap-2">
                {(acc?.tech || acc?.admin) && r.stage === 'NEW' && !r.techId && <button type="button" disabled={busy} onClick={() => void claim(r.id)} className="rounded-lg bg-orange-500 px-3 py-1 text-xs font-bold text-white">📞 آخذه أنا</button>}
                {(acc?.tech || acc?.admin) && r.stage === 'NEW' && <button type="button" onClick={() => { setOpen(open === r.id + 't' ? null : r.id + 't'); setTc({ outcome: '', note: '', quoteGiven: '', quoteAmount: '' }) }} className="rounded-lg bg-[#0f2040] px-3 py-1 text-xs font-bold text-white">سجّل نتيجة الاتصال</button>}
                {(acc?.quality || acc?.admin) && r.stage !== 'CLOSED' && <button type="button" onClick={() => { setOpen(open === r.id + 'q' ? null : r.id + 'q'); setQc({ reached: '', techCalled: '', intent: '', gotQuote: '', quoteFair: '', note: '' }) }} className="rounded-lg bg-emerald-700 px-3 py-1 text-xs font-bold text-white">⭐ اتصلت بالزبون (ابو الجودة)</button>}
              </div>

              {open === r.id + 't' && (
                <div className="mt-2 space-y-2 rounded-lg bg-slate-50 p-3">
                  <Choice id={`tc-out-${r.number}`} label="شنو صار؟" value={tc.outcome} onChange={(v) => setTc({ ...tc, outcome: v })} opts={OUTCOME} />
                  {tc.outcome && tc.outcome !== 'NO_ANSWER' && (
                    <>
                      <Choice id={`tc-q-${r.number}`} label="انطيته عرض سعر؟" value={tc.quoteGiven} onChange={(v) => setTc({ ...tc, quoteGiven: v })} opts={{ YES: 'إي', NO: 'لا' }} />
                      {tc.quoteGiven === 'YES' && <label className="block text-xs">مبلغ العرض (د.ع)<input type="number" className={input} value={tc.quoteAmount} onChange={(e) => setTc({ ...tc, quoteAmount: e.target.value })} /></label>}
                    </>
                  )}
                  <label className="block text-xs">شنو حچيتوا؟<input id={`tc-note-${r.number}`} className={input} value={tc.note} onChange={(e) => setTc({ ...tc, note: e.target.value })} /></label>
                  <button type="button" disabled={busy || !tc.outcome} onClick={() => void contact(r.id)} className="rounded-xl bg-[#0f2040] px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">احفظ</button>
                </div>
              )}
              {open === r.id + 'q' && (
                <div className="mt-2 space-y-2 rounded-lg bg-emerald-50/60 p-3">
                  {r.techOutcome === 'NO_ANSWER' && <p className="text-xs font-bold text-amber-800">التقني كاتب «ما رد» — اسأل الزبون: اتصل بيك التقني؟</p>}
                  {!r.techOutcome && <p className="text-xs font-bold text-amber-800">التقني بعده ما سجّل اتصال — اسأل الزبون إذا أحد تواصل وياه</p>}
                  <Choice id={`qc-r-${r.number}`} label="رد عليك الزبون؟" value={qc.reached} onChange={(v) => setQc({ ...qc, reached: v })} opts={{ YES: 'إي', NO: 'لا' }} />
                  {qc.reached === 'YES' && (
                    <>
                      <Choice id={`qc-tc-${r.number}`} label="اتصل بيك التقني؟" value={qc.techCalled} onChange={(v) => setQc({ ...qc, techCalled: v })} opts={YNU} />
                      <Choice id={`qc-in-${r.number}`} label="شنو چان يريد؟" value={qc.intent} onChange={(v) => setQc({ ...qc, intent: v })} opts={{ JUST_ASKING: 'بس يستفسر', WANTS_INSTALL: 'يريد ينصب منظومة', NOT_INTERESTED: 'ما مهتم' }} />
                      <Choice id={`qc-gq-${r.number}`} label="انطوك عرض سعر؟" value={qc.gotQuote} onChange={(v) => setQc({ ...qc, gotQuote: v })} opts={YNU} />
                      <Choice id={`qc-qf-${r.number}`} label="العرض مناسب ومنافس للسوق؟" value={qc.quoteFair} onChange={(v) => setQc({ ...qc, quoteFair: v })} opts={YNU} />
                    </>
                  )}
                  <label className="block text-xs">ملاحظة<input id={`qc-note-${r.number}`} className={input} value={qc.note} onChange={(e) => setQc({ ...qc, note: e.target.value })} /></label>
                  <button type="button" disabled={busy || !qc.reached || (qc.reached === 'YES' && (!qc.techCalled || !qc.intent || !qc.gotQuote || !qc.quoteFair))}
                    onClick={() => void quality(r.id)} className="rounded-xl bg-emerald-700 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">احفظ المتابعة</button>
                </div>
              )}
            </div>
          ))}
        </section>
      )}
    </div>
  )
}
