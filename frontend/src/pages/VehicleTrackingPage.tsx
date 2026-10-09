import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type VehicleTracking, type VehicleTrackRating } from '../api'

// ═══ 🚗 متابعة السيارات — نفس ملف إكسل (ع) 10-09 ═══
// أبو الكميات يعبّي كل يوم جدول لكل السيارات (الغسل + عشر بنود من 0 لـ4 + وصف
// العطل)، والنسبة الموزونة تطلع لحالها. والفترة الشهرية من ١٧ لـ١٦، وإحصائيات
// كل سيارة بنفس أعمدة شيت «الإحصائيات».

const ITEMS: { key: keyof VehicleTrackRating; label: string; w: number }[] = [
  { key: 'exteriorClean', label: 'نظافة الهيكل الخارجي', w: 0.07 },
  { key: 'exteriorCondition', label: 'حالة الهيكل الخارجي', w: 0.1 },
  { key: 'tireCondition', label: 'حالة الإطارات', w: 0.1 },
  { key: 'glassClean', label: 'تنظيف الزجاج', w: 0.05 },
  { key: 'lightsCondition', label: 'حالة اللايتات', w: 0.05 },
  { key: 'technicalFaults', label: 'الأعطال الفنية', w: 0.25 },
  { key: 'interiorClean', label: 'التنظيف الداخلي', w: 0.1 },
  { key: 'seatsCondition', label: 'حالة الكراسي', w: 0.07 },
  { key: 'interiorDirt', label: 'الأوساخ الداخلية', w: 0.08 },
  { key: 'smell', label: 'الرائحة', w: 0.05 },
]
const SCORE_CLS = ['bg-red-500', 'bg-orange-400', 'bg-yellow-400', 'bg-lime-500', 'bg-emerald-600']
const MEANING = ['سيئ جداً / غير آمن', 'ضعيف', 'متوسط', 'جيد', 'ممتاز']

type Draft = Partial<Record<keyof VehicleTrackRating, number | string | null>>

const todayKey = () => new Date(Date.now() + 3 * 3600_000).toISOString().slice(0, 10)
const periodOf = (d: string) => {
  const [y, m, day] = d.split('-').map(Number)
  const dt = new Date(Date.UTC(y, m - 1 - (day < 17 ? 1 : 0), 1))
  return dt.toISOString().slice(0, 7)
}
const shiftPeriod = (p: string, n: number) => {
  const [y, m] = p.split('-').map(Number)
  return new Date(Date.UTC(y, m - 1 + n, 1)).toISOString().slice(0, 7)
}
function dailyScore(d: Draft): number | null {
  let num = 0, den = 0
  for (const it of ITEMS) {
    const v = d[it.key]
    if (typeof v === 'number') { num += v * it.w; den += 4 * it.w }
  }
  return den ? (num / den) * 100 : null
}
const pct = (v: number | null | undefined) => (v == null ? '—' : `${v.toFixed(1)}%`)
const scoreTone = (v: number | null) => (v == null ? 'text-slate-400' : v >= 80 ? 'text-emerald-700' : v >= 60 ? 'text-amber-700' : 'text-red-700')
const GRADE_CLS: Record<string, string> = {
  'ممتاز': 'bg-emerald-100 text-emerald-800', 'جيد جداً': 'bg-lime-100 text-lime-800', 'جيد': 'bg-sky-100 text-sky-800',
  'مقبول': 'bg-amber-100 text-amber-800', 'ضعيف': 'bg-red-100 text-red-800',
}

function ScorePicker({ value, onPick }: { value: number | null | undefined; onPick: (v: number | null) => void }) {
  return (
    <div className="flex gap-0.5">
      {[0, 1, 2, 3, 4].map((n) => (
        <button key={n} type="button" title={MEANING[n]} onClick={() => onPick(value === n ? null : n)}
          className={`h-7 w-7 rounded-md text-xs font-black ${value === n ? `${SCORE_CLS[n]} text-white shadow` : 'bg-slate-100 text-slate-500 hover:bg-slate-200'}`}>{n}</button>
      ))}
    </div>
  )
}

function HelpDot() {
  const [open, setOpen] = useState(false)
  return (
    <span className="relative inline-block">
      <button type="button" onClick={() => setOpen((o) => !o)} aria-label="شلون ينحسب؟"
        className="flex h-6 w-6 items-center justify-center rounded-full border border-white/50 text-xs font-black">؟</button>
      {open && (
        <span className="absolute left-0 top-8 z-30 block w-80 rounded-xl border border-slate-200 bg-white p-3 text-[11px] leading-relaxed text-slate-600 shadow-lg">
          <b className="block text-slate-800">الدرجات</b>
          {MEANING.map((m, i) => <span key={m} className="block">{i} = {m}</span>)}
          <b className="mt-2 block text-slate-800">الأوزان</b>
          {ITEMS.map((it) => <span key={it.key} className="block">{it.label}: {Math.round(it.w * 100)}%</span>)}
          <span className="mt-2 block">الغسل ما يدخل بنسبة اليوم — يدخل بالشهري بوزن 8%.</span>
          <span className="block">التقدير: ممتاز 90+ · جيد جداً 80+ · جيد 70+ · مقبول 60+ · ضعيف.</span>
        </span>
      )}
    </span>
  )
}

export default function VehicleTrackingPage() {
  const [date, setDate] = useState(todayKey())
  const [period, setPeriod] = useState(periodOf(todayKey()))
  const [data, setData] = useState<VehicleTracking | null>(null)
  const [tab, setTab] = useState<'day' | 'stats' | 'log'>('day')
  const [drafts, setDrafts] = useState<Record<string, Draft>>({})
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)

  const load = useCallback(() => {
    api.getVehicleTracking(period).then(setData).catch((e) => setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر' }))
  }, [period])
  useEffect(load, [load])

  // تعبئة جدول اليوم من المسجّل أصلاً
  useEffect(() => {
    if (!data) return
    const next: Record<string, Draft> = {}
    for (const v of data.vehicles) {
      const r = data.ratings.find((x) => x.vehicleId === v.id && x.date === date)
      next[v.id] = r ? { ...r } : {}
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setDrafts(next)
  }, [data, date])

  const set = (vid: string, key: keyof VehicleTrackRating, val: number | string | null) =>
    setDrafts((d) => ({ ...d, [vid]: { ...d[vid], [key]: val } }))

  const save = async () => {
    if (!data) return
    setBusy(true); setMsg(null)
    try {
      await api.saveVehicleTrackingDay(date, data.vehicles.map((v) => {
        const d = drafts[v.id] ?? {}
        const row: Record<string, unknown> = { vehicleId: v.id, wash: d.wash ?? null, faultDescription: d.faultDescription ?? null }
        for (const it of ITEMS) row[it.key] = d[it.key] ?? null
        return row
      }))
      setMsg({ ok: true, t: '✅ انحفظ يوم ' + date })
      load()
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر الحفظ' }) } finally { setBusy(false) }
  }

  const [mergeTo, setMergeTo] = useState<Record<string, string>>({})
  const merge = async (fromId: string) => {
    const toId = mergeTo[fromId]
    if (!toId) return
    setBusy(true); setMsg(null)
    try {
      await api.mergeTrackingVehicle(fromId, toId)
      setMsg({ ok: true, t: '✅ اندمجت السيارتين — التقييمات صارت على سيارة النظام' })
      load()
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر الدمج' }) } finally { setBusy(false) }
  }
  const temps = data?.vehicles.filter((v) => v.temp) ?? []
  const realCars = data?.vehicles.filter((v) => !v.temp) ?? []

  const days = useMemo(() => {
    if (!data) return []
    return [...new Set(data.ratings.map((r) => r.date))].sort().reverse()
  }, [data])

  return (
    <div dir="rtl" className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h1 className="text-2xl font-extrabold">🚗 متابعة السيارات</h1>
            <p className="mt-1 text-sm text-blue-100">تقييم يومي للنظافة والحالة (0–4)، والإحصائيات تنحسب لحالها — الفترة من 17 لـ16.</p>
          </div>
          <HelpDot />
        </div>
        <div className="mt-3 flex flex-wrap items-center gap-2 text-sm">
          <button onClick={() => setPeriod(shiftPeriod(period, -1))} className="rounded-lg bg-white/15 px-3 py-1.5 font-bold">›</button>
          <span className="rounded-lg bg-white/15 px-3 py-1.5 font-bold">الفترة: {data ? `${data.from} ← ${data.to}` : period}</span>
          <button onClick={() => setPeriod(shiftPeriod(period, 1))} className="rounded-lg bg-white/15 px-3 py-1.5 font-bold">‹</button>
        </div>
      </div>

      <div className="flex gap-2 overflow-x-auto">
        {([['day', '📋 التقييم اليومي'], ['stats', '📊 الإحصائيات'], ['log', '🗓️ سجل الفترة']] as const).map(([k, l]) => (
          <button key={k} onClick={() => setTab(k)}
            className={`shrink-0 rounded-xl border px-4 py-2 text-sm font-bold ${tab === k ? 'border-[#2c5aad] bg-[#2c5aad] text-white' : 'border-slate-200 bg-white text-slate-600'}`}>{l}</button>
        ))}
      </div>
      {temps.length > 0 && (
        <div className="rounded-2xl border border-amber-300 bg-amber-50 p-4 text-sm">
          <b className="text-amber-800">🔗 سيارات انضافت من الإكسل برقم مؤقت — إذا هي نفس سيارة موجودة بالنظام اختارها وادمج:</b>
          <div className="mt-2 space-y-2">
            {temps.map((t) => (
              <div key={t.id} className="flex flex-wrap items-center gap-2">
                <span className="min-w-[10rem] font-bold text-slate-700">{t.name}</span>
                <span className="text-slate-500">هي نفس:</span>
                <select value={mergeTo[t.id] ?? ''} onChange={(e) => setMergeTo((m) => ({ ...m, [t.id]: e.target.value }))}
                  className="rounded-lg border border-slate-300 bg-white px-2 py-1.5">
                  <option value="">— اختار سيارة النظام —</option>
                  {realCars.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
                <button disabled={busy || !mergeTo[t.id]} onClick={() => void merge(t.id)}
                  className="rounded-lg bg-amber-600 px-3 py-1.5 font-bold text-white disabled:opacity-40">دمج</button>
              </div>
            ))}
          </div>
          <p className="mt-2 text-xs text-amber-700">السيارة المؤقتة ما تنمسح — بس تنطفي، وتقييماتها تنتقل.</p>
        </div>
      )}
      {msg && <p className={`rounded-xl px-4 py-2 text-sm font-bold ${msg.ok ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>{msg.t}</p>}
      {!data ? <p className="text-slate-400">جاري التحميل…</p> : tab === 'day' ? (
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <label className="text-sm font-bold text-slate-600">اليوم:</label>
            <input type="date" value={date} onChange={(e) => { setDate(e.target.value); setPeriod(periodOf(e.target.value)) }}
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
            <span className="text-xs text-slate-400">{new Date(date).toLocaleDateString('ar-IQ', { weekday: 'long' })}</span>
          </div>
          <div className="grid gap-3 xl:grid-cols-2">
            {data.vehicles.map((v) => {
              const d = drafts[v.id] ?? {}
              const s = dailyScore(d)
              return (
                <div key={v.id} className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-base font-extrabold text-[#0f2040]">🚙 {v.name}</p>
                    <span className={`text-xl font-black ${scoreTone(s)}`}>{pct(s)}</span>
                  </div>
                  <div className="mt-2 flex gap-2">
                    {([[4, '🧽 تم الغسل', 'bg-emerald-600'], [0, '✖ لم يتم الغسل', 'bg-red-500']] as const).map(([val, l, cls]) => (
                      <button key={val} type="button" onClick={() => set(v.id, 'wash', d.wash === val ? null : val)}
                        className={`flex-1 rounded-lg px-2 py-1.5 text-xs font-bold ${d.wash === val ? `${cls} text-white` : 'bg-slate-100 text-slate-500'}`}>{l}</button>
                    ))}
                  </div>
                  <div className="mt-3 grid gap-x-4 gap-y-1.5 sm:grid-cols-2">
                    {ITEMS.map((it) => (
                      <div key={it.key} className="flex items-center justify-between gap-2">
                        <span className="text-xs text-slate-600">{it.label} <b className="text-[10px] text-sky-700">{Math.round(it.w * 100)}%</b></span>
                        <ScorePicker value={d[it.key] as number | null | undefined} onPick={(n) => set(v.id, it.key, n)} />
                      </div>
                    ))}
                  </div>
                  <input value={(d.faultDescription as string) ?? ''} onChange={(e) => set(v.id, 'faultDescription', e.target.value)}
                    placeholder="وصف العطل (إذا أكو)" className="mt-3 w-full rounded-lg border border-slate-200 px-3 py-2 text-sm" />
                </div>
              )
            })}
          </div>
          <div className="sticky bottom-0 z-10 flex justify-center border-t border-slate-200 bg-white/95 py-3 backdrop-blur">
            <button disabled={busy} onClick={() => void save()} className="rounded-xl bg-[#2c5aad] px-8 py-2.5 text-sm font-bold text-white shadow disabled:opacity-50">
              {busy ? 'جاري الحفظ…' : '💾 حفظ اليوم'}
            </button>
          </div>
        </div>
      ) : tab === 'stats' ? (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
          <table className="w-full text-right text-xs">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-3 py-2.5">السيارة</th><th className="px-3 py-2.5">التقييمات</th><th className="px-3 py-2.5">النقاط</th>
                <th className="px-3 py-2.5">النسبة الخام</th><th className="px-3 py-2.5">تم الغسل %</th>
                {ITEMS.map((it) => <th key={it.key} className="px-3 py-2.5">{it.label}<span className="block text-[10px] font-normal text-sky-700">وزنه {Math.round(it.w * 100)}%</span></th>)}
                <th className="px-3 py-2.5">النتيجة الموزونة</th><th className="px-3 py-2.5">التقدير</th>
                <th className="px-3 py-2.5">تم الغسل</th><th className="px-3 py-2.5">لم يتم</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {data.stats.map((s) => (
                <tr key={s.vehicleId}>
                  <td className="whitespace-nowrap px-3 py-2.5 font-bold text-slate-800">{s.vehicleName}</td>
                  <td className="px-3 py-2.5">{s.completed}</td>
                  <td className="whitespace-nowrap px-3 py-2.5">{s.rawSum} / {s.rawMax}</td>
                  <td className="px-3 py-2.5">{pct(s.rawPct)}</td>
                  <td className="px-3 py-2.5">{pct(s.washPct)}</td>
                  {ITEMS.map((it) => <td key={it.key} className={`px-3 py-2.5 ${scoreTone(s.items[it.key] ?? null)}`}>{pct(s.items[it.key])}</td>)}
                  <td className={`px-3 py-2.5 text-sm font-black ${scoreTone(s.weighted)}`}>{pct(s.weighted)}</td>
                  <td className="px-3 py-2.5">{s.grade && <span className={`whitespace-nowrap rounded-lg px-2 py-0.5 font-bold ${GRADE_CLS[s.grade]}`}>{s.grade}</span>}</td>
                  <td className="px-3 py-2.5 text-emerald-700">{s.washDone}</td>
                  <td className="px-3 py-2.5 text-red-600">{s.washNot}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="space-y-3">
          {days.length === 0 && <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-400">ماكو تقييمات بهالفترة.</p>}
          {days.map((day) => (
            <div key={day} className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="flex items-center justify-between border-b border-slate-100 px-4 py-2.5">
                <b className="text-sm text-[#0f2040]">{new Date(day).toLocaleDateString('ar-IQ', { weekday: 'long' })} — {day}</b>
                <button onClick={() => { setDate(day); setTab('day') }} className="text-xs font-bold text-sky-700">✏️ تعديل</button>
              </div>
              <table className="w-full text-right text-xs">
                <thead className="text-slate-500"><tr>
                  <th className="px-3 py-2">السيارة</th><th className="px-3 py-2">الغسل</th>
                  {ITEMS.map((it) => <th key={it.key} className="px-2 py-2">{it.label}</th>)}
                  <th className="px-3 py-2">وصف العطل</th><th className="px-3 py-2">الموزون %</th>
                </tr></thead>
                <tbody className="divide-y divide-slate-100">
                  {data.ratings.filter((r) => r.date === day).map((r) => (
                    <tr key={r.id}>
                      <td className="whitespace-nowrap px-3 py-2 font-bold">{data.vehicles.find((v) => v.id === r.vehicleId)?.name ?? '—'}</td>
                      <td className="whitespace-nowrap px-3 py-2">{r.wash == null ? '—' : r.wash >= 2 ? '✅ تم' : '✖ لم يتم'}</td>
                      {ITEMS.map((it) => {
                        const v = r[it.key] as number | null
                        return <td key={it.key} className="px-2 py-2 text-center">{v == null ? '—' : <span className={`inline-block h-6 w-6 rounded-md leading-6 text-white ${SCORE_CLS[v]}`}>{v}</span>}</td>
                      })}
                      <td className="max-w-[14rem] truncate px-3 py-2 text-slate-500" title={r.faultDescription ?? ''}>{r.faultDescription ?? ''}</td>
                      <td className={`px-3 py-2 font-black ${scoreTone(r.score)}`}>{pct(r.score)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
