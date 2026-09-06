// ═══ عارض منحنيات اللوح ═══
//
// «أريد كأني فاتح ماتلاب» — وهذا أقرب شي إلها بالمختبر: المتدرّب
// يشوف **المنحنى نفسه** مو رقماً نهائياً.
//
// ⚠️⚠️ **الرسم بـSVG مو بمكتبة رسم.** المنحنى ١٢٠ نقطة وثلاث محاور،
// وإضافة مكتبة كاملة لهذا تضيف مئات الكيلوبايتات للحزمة بلا مقابل —
// والمختبر أصلاً يجرّ Babylon وراه.
//
// ⚠️ وكل نقطة تنرسم **من نفس النموذج** الي يحسب عليه المحرّك. عارض
// يرسم من جدول ثانٍ يفترق عن الحساب بأول تعديل، والمتدرّب يشوف قمة
// بمحل والرقم يقول محلاً ثانياً.

import { useMemo, useState } from 'react'

import {
  findMpp, ivCurve, panelModel, simulateDay,
  type PanelSpec,
} from './pv'

interface Props {
  spec: PanelSpec
  /** عدد الألواح بالسلسلة — لعرض جهد الستring. */
  count: number
  /** حرارة الجو بالظهر. */
  ambientC: number
  onClose: () => void
}

const W = 460
const H = 220
const PAD = { l: 46, r: 44, t: 14, b: 26 }

/** محور + شبكة — مشتركة بين المنحنيين. */
function Frame({ xMax, yMax, yMax2, xLabel, yLabel, yLabel2 }: {
  xMax: number; yMax: number; yMax2?: number; xLabel: string; yLabel: string; yLabel2?: string
}) {
  const iw = W - PAD.l - PAD.r
  const ih = H - PAD.t - PAD.b
  return (
    <g>
      {[0, 0.25, 0.5, 0.75, 1].map((f) => (
        <g key={f}>
          <line x1={PAD.l} y1={PAD.t + ih * f} x2={PAD.l + iw} y2={PAD.t + ih * f}
            stroke="#1e293b" strokeWidth={1} />
          <text x={PAD.l - 6} y={PAD.t + ih * f + 3.5} fontSize={8.5} fill="#64748b" textAnchor="end">
            {(yMax * (1 - f)).toFixed(yMax < 20 ? 1 : 0)}
          </text>
          {yMax2 !== undefined && (
            <text x={PAD.l + iw + 6} y={PAD.t + ih * f + 3.5} fontSize={8.5} fill="#a78bfa" textAnchor="start">
              {(yMax2 * (1 - f)).toFixed(0)}
            </text>
          )}
        </g>
      ))}
      {[0, 0.25, 0.5, 0.75, 1].map((f) => (
        <text key={f} x={PAD.l + iw * f} y={H - 8} fontSize={8.5} fill="#64748b" textAnchor="middle">
          {(xMax * f).toFixed(xMax < 30 ? 1 : 0)}
        </text>
      ))}
      <line x1={PAD.l} y1={PAD.t} x2={PAD.l} y2={PAD.t + ih} stroke="#334155" strokeWidth={1.2} />
      <line x1={PAD.l} y1={PAD.t + ih} x2={PAD.l + iw} y2={PAD.t + ih} stroke="#334155" strokeWidth={1.2} />
      <text x={PAD.l + iw / 2} y={H - 0.5} fontSize={8.5} fill="#94a3b8" textAnchor="middle">{xLabel}</text>
      <text x={10} y={PAD.t + ih / 2} fontSize={8.5} fill="#38bdf8" textAnchor="middle"
        transform={`rotate(-90, 10, ${PAD.t + ih / 2})`}>{yLabel}</text>
      {yLabel2 && (
        <text x={W - 8} y={PAD.t + ih / 2} fontSize={8.5} fill="#a78bfa" textAnchor="middle"
          transform={`rotate(-90, ${W - 8}, ${PAD.t + ih / 2})`}>{yLabel2}</text>
      )}
    </g>
  )
}

export default function CurveView({ spec, count, ambientC, onClose }: Props) {
  /** ⚠️ منزلقان: الإشعاع والحرارة. التفاعل هو الي يخلّي المتدرّب
   *  يشوف **القمة تتحرك** — والحركة تعلّم أكثر من صورة ثابتة. */
  const [irr, setIrr] = useState(1000)
  const [amb, setAmb] = useState(ambientC)

  const { curve, mpp, model, day } = useMemo(() => {
    const m = panelModel(spec, irr, amb)
    return {
      model: m,
      curve: ivCurve(m, 120),
      mpp: findMpp(m, 300),
      day: simulateDay(spec, { panels: count, peakIrr: irr, minC: amb - 12, maxC: amb + 6, stepH: 0.5 }),
    }
  }, [spec, irr, amb, count])

  const iw = W - PAD.l - PAD.r
  const ih = H - PAD.t - PAD.b
  const vMax = Math.max(1, model.voc)
  const iMax = Math.max(0.1, model.isc * 1.05)
  const pMax = Math.max(1, mpp.pmax * 1.1)

  const px = (v: number) => PAD.l + (v / vMax) * iw
  const pyI = (i: number) => PAD.t + ih - (i / iMax) * ih
  const pyP = (p: number) => PAD.t + ih - (p / pMax) * ih

  const pathI = curve.map((pt, k) => `${k === 0 ? 'M' : 'L'} ${px(pt.v).toFixed(1)} ${pyI(pt.i).toFixed(1)}`).join(' ')
  const pathP = curve.map((pt, k) => `${k === 0 ? 'M' : 'L'} ${px(pt.v).toFixed(1)} ${pyP(pt.p).toFixed(1)}`).join(' ')

  const dayMax = Math.max(1, ...day.points.map((p) => p.pmax))
  const dx = (h: number) => PAD.l + (h / 24) * iw
  const dy = (p: number) => PAD.t + ih - (p / dayMax) * ih
  const pathD = day.points.map((p, k) => `${k === 0 ? 'M' : 'L'} ${dx(p.hour).toFixed(1)} ${dy(p.pmax).toFixed(1)}`).join(' ')

  const stat = (k: string, v: string, tone = 'text-slate-100') => (
    <div key={k} className="rounded-lg bg-[#0b1220] px-2.5 py-1.5 ring-1 ring-slate-800">
      <p className="text-[9.5px] text-slate-500">{k}</p>
      <p className={`font-mono text-[12.5px] font-bold tabular-nums ${tone}`}>{v}</p>
    </div>
  )

  return (
    <div className="border-t border-slate-800 bg-[#0b1220] p-3">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <button onClick={onClose} className="text-[11px] font-bold text-slate-500 hover:text-slate-300">✕ سكّر</button>
        <p className="text-xs font-bold text-slate-200">📈 منحنيات اللوح</p>
      </div>

      {/* ⚠️ المنزلقان فوگ المنحنى: المتدرّب يحرّك ويشوف الأثر بنفس
          النظرة. تحتهما يعني يحرّك وينزل بصره ويرجع. */}
      <div className="mb-2 grid gap-2 sm:grid-cols-2">
        {([['الإشعاع', irr, 0, 1200, 50, setIrr, 'W/m²'],
           ['حرارة الجو', amb, -10, 55, 1, setAmb, '°C']] as const).map(([label, val, min, max, step, set, unit]) => (
          <label key={label} className="flex items-center gap-2 rounded-lg bg-[#0e1626] px-3 py-1.5 ring-1 ring-slate-800">
            <span className="w-20 text-[11px] text-slate-400">{label}</span>
            <input type="range" min={min} max={max} step={step} value={val}
              onChange={(e) => set(Number(e.target.value))} className="flex-1 accent-sky-500" />
            <span className="w-20 text-left font-mono text-[11.5px] tabular-nums text-slate-200">{val} {unit}</span>
          </label>
        ))}
      </div>

      <div className="grid gap-2 lg:grid-cols-2">
        <div className="rounded-xl bg-black p-2 ring-1 ring-slate-800">
          <p className="mb-0.5 px-1 text-[10.5px] font-bold text-slate-400">منحنى I–V و P–V</p>
          <svg viewBox={`0 0 ${W} ${H}`} className="w-full">
            <Frame xMax={vMax} yMax={iMax} yMax2={pMax} xLabel="الجهد (V)" yLabel="التيار (A)" yLabel2="القدرة (W)" />
            <path d={pathP} fill="none" stroke="#a78bfa" strokeWidth={1.6} />
            <path d={pathI} fill="none" stroke="#38bdf8" strokeWidth={2} />
            {/* ⚠️ القمة مؤشّرة بخط عمودي مو بنقطة: النقطة تضيع بين
                المنحنيين، والخط يوري **الجهد** الي عندها — وهذا الي
                يضبطه الإنفرتر فعلاً. */}
            <line x1={px(mpp.vmp)} y1={PAD.t} x2={px(mpp.vmp)} y2={PAD.t + ih}
              stroke="#f59e0b" strokeWidth={1} strokeDasharray="3 3" />
            <circle cx={px(mpp.vmp)} cy={pyP(mpp.pmax)} r={3.5} fill="#f59e0b" />
            <text x={px(mpp.vmp)} y={PAD.t + 9} fontSize={9} fill="#fbbf24" textAnchor="middle">MPP</text>
          </svg>
        </div>

        <div className="rounded-xl bg-black p-2 ring-1 ring-slate-800">
          <p className="mb-0.5 px-1 text-[10.5px] font-bold text-slate-400">
            إنتاج اليوم — {count} لوح · {day.kwh.toFixed(2)} kWh
          </p>
          <svg viewBox={`0 0 ${W} ${H}`} className="w-full">
            <Frame xMax={24} yMax={dayMax} xLabel="الساعة" yLabel="القدرة (W)" />
            <path d={`${pathD} L ${dx(24)} ${dy(0)} L ${dx(0)} ${dy(0)} Z`} fill="#f59e0b" fillOpacity={0.12} />
            <path d={pathD} fill="none" stroke="#f59e0b" strokeWidth={2} />
          </svg>
        </div>
      </div>

      <div className="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-6">
        {stat('حرارة الخلية', `${model.cellT.toFixed(1)} °C`, model.cellT > 60 ? 'text-amber-300' : 'text-slate-100')}
        {stat('Voc', `${model.voc.toFixed(2)} V`)}
        {stat('Isc', `${model.isc.toFixed(2)} A`)}
        {stat('Vmp', `${mpp.vmp.toFixed(2)} V`)}
        {stat('Pmax للوح', `${mpp.pmax.toFixed(0)} W`)}
        {stat(`ستring ${count}`, `${(model.voc * count).toFixed(0)} V`,
          model.voc * count > 450 ? 'text-red-300' : 'text-slate-100')}
      </div>

      {/* ⚠️ تفسير مكتوب تحت الأرقام: رقم بلا تفسير يُقرا ويُنسى.
          والفرق بين حرارة الجو وحرارة الخلية هو أكثر شي ينُسى. */}
      <p className="mt-2 text-[10.5px] leading-relaxed text-slate-500">
        حرارة الخلية <b className="text-slate-400">{model.cellT.toFixed(0)}°</b> والجو <b className="text-slate-400">{amb}°</b> —
        الفرق <b className="text-amber-400">{(model.cellT - amb).toFixed(0)} درجة</b>، وهو الي ياكل من الإنتاج.
        {' '}جرّب تنزّل الحرارة وشوف القمة شلون ترتفع وتنتقل لليمين.
      </p>
    </div>
  )
}
