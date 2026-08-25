// ═══ مساحة العمل ═══
//
// «أريد كأنما آني فاتح أقوى تطبيقات المحاكاة — شي منظّم ومرتّب وقوي».
//
// التخطيط المتعارف عليه بكل أدوات المحاكاة (MATLAB/Simulink، Packet
// Tracer، Fusion): **كتالوگ على جنب، لوح بالنص، خصائص على الجنب
// الثاني، وشريط أدوات فوگ ونتائج تحت**. ما نخترع تخطيطاً جديداً —
// الفني الي استعمل أي أداة قبل يعرف وين يدوّر بلا ما نعلّمه.
//
// ⚠️ المحرّك ما يشتغل مع كل حركة: يشتغل لمن تضغط **تشغيل**. المحاكاة
// المستمرة تخلّي اللوح يترمّش وأنت تسحب قطعة، وتخفي متى تغيّرت
// النتيجة فعلاً.

import { useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import SimGate from '../sim/SimGate'
import Canvas from './Canvas'
import { CABLE_BY_ID, LINK_PARAMS, MEDIUM_AR } from './cables'
import { DOMAINS, PARTS, PART_BY_ID } from './catalog'
import { electricalEngine } from './engines/electrical'
import { networkEngine } from './engines/network'
import { solarEngine } from './engines/solar'
import { Symbol } from './symbols'
import type { DomainEngine, DomainId, LabDoc, ParamDef, SimResult } from './types'

const ENGINES: Record<DomainId, DomainEngine> = {
  electrical: electricalEngine,
  solar: solarEngine,
  network: networkEngine,
}

export default function LabWorkbench() {
  return <SimGate><Bench /></SimGate>
}

function Bench() {
  const [domain, setDomain] = useState<DomainId>('network')
  const [docs, setDocs] = useState<Record<DomainId, LabDoc>>({
    network: { domain: 'network', nodes: [], links: [] },
    solar: { domain: 'solar', nodes: [], links: [] },
    electrical: { domain: 'electrical', nodes: [], links: [] },
  })
  const [selected, setSelected] = useState<string | null>(null)
  const [pendingPart, setPendingPart] = useState<string | null>(null)
  const [result, setResult] = useState<SimResult | null>(null)

  const doc = docs[domain]
  const setDoc = useCallback((updater: (d: LabDoc) => LabDoc) => {
    setDocs((all) => ({ ...all, [domain]: updater(all[domain]) }))
    // ⚠️ النتيجة تنمسح مع أي تعديل: إبقاء قراءات قديمة فوگ مخطط
    // تغيّر يعني الفني يقرأ رقماً ما يخص الي قدّامه.
    setResult(null)
  }, [domain])

  const parts = useMemo(() => PARTS.filter((p) => p.domain === domain), [domain])
  const selNode = doc.nodes.find((n) => n.id === selected)
  const selPart = selNode ? PART_BY_ID[selNode.partId] : null
  const selLink = doc.links.find((l) => l.id === selected)

  const run = () => setResult(ENGINES[domain].run(doc, PART_BY_ID))
  const clearAll = () => { setDoc(() => ({ domain, nodes: [], links: [] })); setSelected(null) }

  const setParam = (id: string, value: string | number | boolean) => {
    if (!selNode) return
    setDoc((d) => ({ ...d, nodes: d.nodes.map((n) => (n.id === selNode.id ? { ...n, params: { ...n.params, [id]: value } } : n)) }))
  }

  const setLinkParam = (id: string, value: string | number | boolean) => {
    if (!selLink) return
    setDoc((d) => ({
      ...d,
      links: d.links.map((l) => (l.id === selLink.id ? { ...l, params: { ...(l.params ?? {}), [id]: value } } : l)),
    }))
  }

  // ⚠️ إزالة التكرار: المحرّك يفحص كل منفذ لحاله، فغلط واحد بمنفذين
  // (الموجب والسالب) يطلع رسالتين متطابقتين. الفني يقراها مرتين ويظن
  // إن عنده غلطتين.
  const msgs = useMemo(() => {
    const seen = new Set<string>()
    return (result?.messages ?? []).filter((m) => {
      const k = m.kind + '|' + m.text
      if (seen.has(k)) return false
      seen.add(k)
      return true
    })
  }, [result])
  const errors = msgs.filter((m) => m.kind === 'error')
  const warns = msgs.filter((m) => m.kind === 'warn')

  return (
    <div dir="rtl" className="space-y-3">
      {/* ⚠️ شريط السلامة دائم — نفس قاعدة بقية المختبر. */}
      <div className="rounded-xl bg-amber-50 px-4 py-2.5 text-[12.5px] font-bold text-amber-900 ring-1 ring-amber-200">
        ⚠️ محاكاة تدريب — القيم نمطية عامة مو كتالوگ موديل بعينه. الرجوع لكتالوگ
        الشركة المصنّعة إلزامي قبل أي تنفيذ حقيقي.
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <Link to="/simulator-lab" className="text-sm text-brand-700 hover:underline">← رجوع للمختبر</Link>
        <h2 className="text-xl font-bold text-brand-900">🧰 مساحة عمل المحاكاة</h2>
      </div>

      {/* ═══ شريط الأدوات ═══ */}
      <div className="flex flex-wrap items-center gap-2 rounded-xl bg-white p-2.5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="inline-flex overflow-hidden rounded-lg ring-1 ring-slate-300">
          {DOMAINS.map((d) => (
            <button
              key={d.id}
              onClick={() => { setDomain(d.id); setSelected(null); setPendingPart(null); setResult(null) }}
              className={`px-3.5 py-1.5 text-xs font-bold transition ${
                domain === d.id ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 hover:bg-slate-50'}`}
            >
              {d.icon} {d.name}
            </button>
          ))}
        </div>

        <div className="mr-auto flex items-center gap-2">
          <button onClick={run}
            className="rounded-lg bg-emerald-600 px-4 py-1.5 text-xs font-bold text-white shadow hover:bg-emerald-700">
            ▶ تشغيل المحاكاة
          </button>
          <button onClick={() => setResult(null)} disabled={!result}
            className="rounded-lg bg-slate-200 px-3 py-1.5 text-xs font-bold text-slate-700 disabled:opacity-40">
            ⏹ إيقاف
          </button>
          <button onClick={clearAll}
            className="rounded-lg bg-white px-3 py-1.5 text-xs font-bold text-red-600 ring-1 ring-red-200 hover:bg-red-50">
            🗑 تفريغ اللوح
          </button>
        </div>
      </div>

      <p className="text-xs text-slate-500">{DOMAINS.find((d) => d.id === domain)?.about}</p>

      {/* ⚠️ اللوح يحتاج مساحة — بالموبايل ما يشتغل، والرسالة الصريحة
          أنظف من محرّر مصغّر ما ينستعمل. */}
      <div className="rounded-2xl bg-white p-8 text-center shadow md:hidden">
        <p className="text-lg font-bold text-slate-700">🖥️ افتحه من الحاسبة</p>
        <p className="mt-2 text-sm text-slate-500">مساحة العمل تحتاج شاشة كبيرة — سحب قطع وربط بشاشة ٦ إنچ ما يشتغل.</p>
      </div>

      <div className="hidden gap-3 md:grid md:grid-cols-[13rem_1fr_17rem]">
        {/* ═══ الكتالوگ ═══ */}
        <div className="rounded-xl bg-white p-3 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-2 text-xs font-bold text-slate-500">القطع — اضغط قطعة ثم اضغط باللوح</h3>
          <div className="space-y-1.5">
            {parts.map((p) => (
              <button
                key={p.id}
                onClick={() => setPendingPart(pendingPart === p.id ? null : p.id)}
                className={`flex w-full items-center gap-2 rounded-lg border p-2 text-right transition ${
                  pendingPart === p.id ? 'border-emerald-500 bg-emerald-50' : 'border-slate-200 hover:border-brand-400 hover:bg-slate-50'}`}
              >
                <svg viewBox={`0 0 ${p.w} ${p.h}`} width={38} height={26} className="shrink-0">
                  <Symbol symbol={p.symbol} w={p.w} h={p.h} accent="#64748b"
                    params={Object.fromEntries(p.params.map((x) => [x.id, x.default]))} />
                </svg>
                <span className="flex-1 text-[12px] font-bold text-slate-700">{p.name}</span>
              </button>
            ))}
          </div>
        </div>

        {/* ═══ اللوح ═══ */}
        <div className="flex min-h-[560px] flex-col gap-2">
          <Canvas
            doc={doc} setDoc={setDoc} result={result}
            selected={selected} setSelected={setSelected}
            pendingPart={pendingPart} onPlaced={() => setPendingPart(null)}
          />

          {/* ═══ شريط النتائج ═══ */}
          <div className="max-h-44 overflow-y-auto rounded-xl bg-white p-3 text-[12px] shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            {!result && <p className="text-slate-400">اضغط «تشغيل المحاكاة» حتى تشوف النتيجة.</p>}
            {result && msgs.length === 0 && <p className="text-slate-400">ماكو ملاحظات.</p>}
            {result && (
              <>
                {errors.length > 0 && (
                  <p className="mb-2 rounded-lg bg-red-50 px-3 py-1.5 font-bold text-red-700">
                    {errors.length} مشكلة خطيرة — تنقرا تحت.
                  </p>
                )}
                <ul className="space-y-1.5">
                  {msgs.map((m, i) => (
                    <li key={i} className={`rounded-lg px-3 py-1.5 leading-relaxed ${
                      m.kind === 'error' ? 'bg-red-50 font-bold text-red-700'
                        : m.kind === 'warn' ? 'bg-amber-50 text-amber-800'
                        : 'bg-slate-50 text-slate-600'}`}>
                      {m.text}
                    </li>
                  ))}
                </ul>
              </>
            )}
          </div>
        </div>

        {/* ═══ الخصائص ═══ */}
        <div className="rounded-xl bg-white p-3 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-2 text-xs font-bold text-slate-500">الخصائص</h3>
          {!selNode && !selLink && (
            <p className="text-[12px] text-slate-400">اضغط قطعة أو كيبلاً باللوح حتى تشوف خصائصه.</p>
          )}

          {/* ═══ خصائص الكيبل ═══
              ⚠️ هنا الفرق الحقيقي عن Packet Tracer: الكيبل مو خط —
              له نوع وطول وترانسيفر بكل طرف، وكلهن ينفحصن. */}
          {selLink && selLink.params && (
            <div className="space-y-3">
              <div>
                <p className="text-sm font-bold text-brand-900">🔌 كيبل</p>
                <p className="mt-0.5 text-[11px] text-slate-500">
                  {CABLE_BY_ID[String(selLink.params.cable)]?.about}
                </p>
                <p className="mt-1 text-[10.5px] text-slate-400">
                  الوسط: {MEDIUM_AR[CABLE_BY_ID[String(selLink.params.cable)]?.medium ?? 'copper']} ·
                  الحد {CABLE_BY_ID[String(selLink.params.cable)]?.maxM} م
                </p>
              </div>
              {LINK_PARAMS.map((lp) => (
                <Field key={lp.id} def={lp} value={selLink.params?.[lp.id]} onChange={(v) => setLinkParam(lp.id, v)} />
              ))}
              <button
                onClick={() => { setDoc((d) => ({ ...d, links: d.links.filter((l) => l.id !== selLink.id) })); setSelected(null) }}
                className="w-full rounded-lg bg-red-50 py-1.5 text-[11px] font-bold text-red-600 ring-1 ring-red-200 hover:bg-red-100"
              >
                🗑 احذف الكيبل
              </button>
            </div>
          )}
          {selNode && selPart && (
            <div className="space-y-3">
              <div>
                <p className="text-sm font-bold text-brand-900">{selPart.name}</p>
                {selPart.about && <p className="mt-0.5 text-[11px] leading-relaxed text-slate-500">{selPart.about}</p>}
              </div>

              {selPart.danger && (
                <p className="rounded-lg bg-red-50 px-2.5 py-1.5 text-[11px] leading-relaxed font-bold text-red-700">
                  ⚠️ {selPart.danger}
                </p>
              )}

              {selPart.params.map((pd) => (
                <Field key={pd.id} def={pd} value={selNode.params[pd.id]} onChange={(v) => setParam(pd.id, v)} />
              ))}

              {result?.nodeReadings[selNode.id] && (
                <div className="rounded-lg bg-slate-900 p-2.5">
                  <p className="mb-1 text-[10px] text-slate-400">القياس</p>
                  {result.nodeReadings[selNode.id].map((r, i) => (
                    <p key={i} className={`font-mono text-[13px] font-bold ${
                      r.tone === 'bad' ? 'text-red-400' : r.tone === 'warn' ? 'text-amber-300' : 'text-emerald-400'}`}>
                      {r.text}
                    </p>
                  ))}
                </div>
              )}

              <button
                onClick={() => {
                  setDoc((d) => ({
                    ...d,
                    nodes: d.nodes.filter((n) => n.id !== selNode.id),
                    links: d.links.filter((l) => l.from.node !== selNode.id && l.to.node !== selNode.id),
                  }))
                  setSelected(null)
                }}
                className="w-full rounded-lg bg-red-50 py-1.5 text-[11px] font-bold text-red-600 ring-1 ring-red-200 hover:bg-red-100"
              >
                🗑 احذف القطعة
              </button>
            </div>
          )}

          <div className="mt-4 border-t border-slate-100 pt-3 text-[10.5px] leading-relaxed text-slate-400">
            <b className="text-slate-500">الاختصارات</b><br />
            اضغط منفذاً ثم منفذاً ثانياً حتى توصّل · Esc يلغي<br />
            Delete يحذف المحدّد · عجلة الماوس تكبّر<br />
            الزر الأيمن مع السحب يحرّك اللوح
          </div>
        </div>
      </div>

      {warns.length > 0 && result && (
        <p className="hidden text-xs text-slate-400 md:block">{warns.length} ملاحظة تحذيرية.</p>
      )}
    </div>
  )
}

function Field({ def, value, onChange }: {
  def: ParamDef
  value: string | number | boolean | undefined
  onChange: (v: string | number | boolean) => void
}) {
  const label = (
    <span className="mb-1 flex items-baseline gap-1 text-[11px] font-bold text-slate-600">
      {def.label}
      {def.unit && <span className="text-slate-400">({def.unit})</span>}
    </span>
  )
  return (
    <label className="block">
      {label}
      {def.kind === 'bool' ? (
        <button
          type="button"
          onClick={() => onChange(!value)}
          className={`w-full rounded-lg px-3 py-1.5 text-[12px] font-bold transition ${
            value ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-slate-600'}`}
        >
          {value ? 'مغلق ✓' : 'مفتوح'}
        </button>
      ) : def.kind === 'select' ? (
        <select
          value={String(value ?? def.default)}
          onChange={(e) => onChange(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-2 py-1.5 text-[12px]"
        >
          {def.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      ) : (
        <input
          type={def.kind === 'number' ? 'number' : 'text'}
          value={String(value ?? def.default)}
          min={def.min} max={def.max}
          dir={def.kind === 'number' ? 'ltr' : undefined}
          onChange={(e) => onChange(def.kind === 'number' ? Number(e.target.value) : e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-2 py-1.5 text-[12px]"
        />
      )}
      {def.help && <span className="mt-0.5 block text-[10px] leading-relaxed text-slate-400">{def.help}</span>}
    </label>
  )
}
