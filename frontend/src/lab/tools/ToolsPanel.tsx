// ═══ لوحة أدوات الفحص ═══
//
// ⚠️ **بشكل الترمنال مو بشكل تقرير**: خط أحادي المسافة على أسود،
// واتجاه `ltr`. الفني الي يتدرّب هنا راح يقرا نفس هالأسطر بالضبط
// على جهاز حقيقي — لو عرضناها بجدول ملوّن عربي، يتعلّم يقرا شكلاً
// ما راح يشوفه بحياته.

import { useState } from 'react'

import { linkTest, ping, pingableNodes, traceroute } from './netTools'
import type { LabDoc } from '../types'

interface Props {
  /** ⚠️ المستند **بعد** حقن الأعطال — الأداة تشوف المنظومة بأعطالها.
   *  أداة تقرا نسخة سليمة تعطي جواباً صحيحاً على شبكة معطوبة، والمتدرّب
   *  يستبعد السبب الصح. */
  doc: LabDoc
  onClose: () => void
}

export default function ToolsPanel({ doc, onClose }: Props) {
  const devices = pingableNodes(doc)
  const [from, setFrom] = useState(devices[0]?.id ?? '')
  const [to, setTo] = useState(devices[1]?.id ?? '')
  const [linkId, setLinkId] = useState(doc.links[0]?.id ?? '')
  const [out, setOut] = useState<string[]>([])
  const [ok, setOk] = useState<boolean | null>(null)

  const run = (r: { lines: string[]; ok: boolean }) => { setOut(r.lines); setOk(r.ok) }

  const linkLabel = (id: string) => {
    const l = doc.links.find((x) => x.id === id)
    if (!l) return id
    const nm = (nid: string) => {
      const n = doc.nodes.find((x) => x.id === nid)
      return String(n?.cliState?.hostname ?? n?.params.name ?? n?.params.hostname ?? nid)
    }
    return `${nm(l.from.node)} ⇄ ${nm(l.to.node)}`
  }

  const sel = 'rounded-lg bg-[#0b1220] px-2 py-1.5 text-[11px] text-slate-200 ring-1 ring-slate-700'
  const btn = 'rounded-lg px-3 py-1.5 text-[11px] font-bold transition disabled:opacity-40'

  return (
    <div className="border-t border-slate-800 bg-[#0b1220] p-3">
      <div className="mb-2 flex items-center justify-between">
        <button onClick={onClose} className="text-[11px] font-bold text-slate-500 hover:text-slate-300">✕ سكّر</button>
        <p className="text-xs font-bold text-slate-200">🧪 أدوات الفحص</p>
      </div>

      {devices.length < 2 && (
        <p className="rounded-lg border border-dashed border-slate-700 px-3 py-4 text-center text-[11px] text-slate-500">
          حط جهازين على الأقل (حاسبة أو كاميرا) حتى تفحص الاتصال بينهم.
        </p>
      )}

      {devices.length >= 2 && (
        <div className="mb-2 flex flex-wrap items-center gap-2">
          <span className="text-[11px] text-slate-500">من</span>
          <select value={from} onChange={(e) => setFrom(e.target.value)} className={sel}>
            {devices.map((d) => <option key={d.id} value={d.id}>{d.label} · {d.ip}</option>)}
          </select>
          <span className="text-[11px] text-slate-500">إلى</span>
          <select value={to} onChange={(e) => setTo(e.target.value)} className={sel}>
            {devices.map((d) => <option key={d.id} value={d.id}>{d.label} · {d.ip}</option>)}
          </select>
          <button
            onClick={() => run(ping(doc, from, to))}
            disabled={from === to}
            className={`${btn} bg-sky-600 text-white hover:bg-sky-500`}
          >
            ping
          </button>
          <button
            onClick={() => run(traceroute(doc, from, to))}
            disabled={from === to}
            className={`${btn} bg-slate-700 text-slate-100 hover:bg-slate-600`}
          >
            traceroute
          </button>
          {doc.links.length > 0 && (
            <>
              <span className="mr-2 text-[11px] text-slate-500">الوصلة</span>
              <select value={linkId} onChange={(e) => setLinkId(e.target.value)} className={sel}>
                {doc.links.map((l) => <option key={l.id} value={l.id}>{linkLabel(l.id)}</option>)}
              </select>
              <button
                onClick={() => run(linkTest(doc, linkId))}
                className={`${btn} bg-slate-700 text-slate-100 hover:bg-slate-600`}
              >
                فحص الوصلة
              </button>
            </>
          )}
        </div>
      )}

      <pre
        dir="ltr"
        className={`h-[220px] overflow-auto rounded-xl bg-black p-3 text-left font-mono text-[11.5px] leading-relaxed ring-1 ${
          out.length === 0 ? 'text-slate-600 ring-slate-800'
            : ok ? 'text-emerald-300 ring-emerald-900/60' : 'text-red-300 ring-red-900/60'}`}
      >
{out.length === 0 ? '# اختر جهازين واضغط ping' : out.join('\n')}
      </pre>

      {/* ⚠️ تنبيه دائم: الأداة تنطي **عَرَضاً** مو تشخيصاً. */}
      <p className="mt-1.5 text-[10.5px] text-slate-600">
        الأداة تنطيك العَرَض مثل الميدان — التشخيص شغلك. جرّب traceroute حتى تعرف <b>وين</b> انقطع.
      </p>
    </div>
  )
}
