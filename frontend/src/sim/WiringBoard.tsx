import { useEffect, useMemo, useState } from 'react'
import type { SimDevice, Scene, SimAction, Terminal, Wire } from './types'
import { TERMINAL_KIND_LABELS } from './types'

// ═══ محرّك التوصيل ═══
//
// «القفل الإلكتروني بيه ١٥ واير ملوّنات — كل لون شنو يعني وشلون
// يربطهنة بكهرباء».
//
// ⚠️ **ماكو حساب إحداثيات للمؤشر**: كل طرف عنصر `<circle>` لحاله بحدث
// ضغط، وكل سلك `<path>` لحاله. لو اعتمدنا على تحويل إحداثيات الضغط
// (مثل محرّرات الرسم) چان صار الأمر هشّاً مع التكبير والتمرير وRTL —
// وهاي أكثر نقطة تنكسر بمحرّرات اللوحات.
//
// ⚠️ والأسلاك تنرسم من **الحالة** مو من DOM: فالتراجع والحفظ والاستئناف
// من وين وقف يشتغلون لحالهم بلا كود إضافي.
//
// ⚠️ الأجهزة تنرسم بالكود مو بصور. والإحداثيات نسب `0..1` من صندوق
// الجهاز — فلمن تنرفع صورة حقيقية بعدين، نفس الإحداثيات تشتغل فوگها.

const VW = 1000 // عرض فضاء الرسم — ثابت، والحجم الفعلي يتكيّف بـCSS
const VH = 620
const DEV_W = 300
const DEV_H = 380
const PSU_H = 190

interface Props {
  scene: Scene
  devices: Record<string, SimDevice>
  wires: Wire[]
  onAction: (a: SimAction) => void
  onRemoveWire: (w: Wire) => void
  /** أطراف تنبض للتلميح. */
  highlight?: string[]
  readOnly?: boolean
}

interface Placed {
  ref: string
  device: SimDevice
  x: number
  y: number
  w: number
  h: number
}

export default function WiringBoard({ scene, devices, wires, onAction, onRemoveWire, highlight, readOnly }: Props) {
  const [armed, setArmed] = useState<string | null>(null)
  const [hover, setHover] = useState<string | null>(null)

  // Esc يلغي التسليح — أول شي يجرّبه أي واحد لمن يغلط بالاختيار.
  useEffect(() => {
    if (!armed) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setArmed(null) }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [armed])

  const placed: Placed[] = useMemo(
    () =>
      (scene.devices ?? [])
        .map((sd) => {
          const device = devices[sd.deviceId]
          if (!device) return null
          const h = (device.terminals?.length ?? 0) > 4 ? DEV_H : PSU_H
          return {
            ref: sd.ref,
            device,
            x: sd.x * VW - DEV_W / 2,
            y: sd.y * VH - h / 2,
            w: DEV_W,
            h,
          }
        })
        .filter((p): p is Placed => p !== null),
    [scene, devices],
  )

  /** موقع الطرف المطلق بفضاء الرسم. */
  const pointOf = useMemo(() => {
    const map = new Map<string, { x: number; y: number; t: Terminal; ref: string }>()
    for (const p of placed) {
      for (const t of p.device.terminals ?? []) {
        map.set(`${p.ref}:${t.id}`, { x: p.x + t.x * p.w, y: p.y + t.y * p.h, t, ref: p.ref })
      }
    }
    return map
  }, [placed])

  const clickTerminal = (id: string) => {
    if (readOnly) return
    if (!armed) { setArmed(id); return }
    if (armed === id) { setArmed(null); return }
    onAction({ op: 'CONNECT', from: armed, to: id })
    setArmed(null)
  }

  // منحنى ناعم بين طرفين — الانحناء أفقي لأن الأطراف على حافّات جانبية.
  const pathFor = (a: { x: number; y: number }, b: { x: number; y: number }) => {
    const dx = Math.max(40, Math.abs(b.x - a.x) * 0.45)
    return `M ${a.x} ${a.y} C ${a.x - dx} ${a.y}, ${b.x + dx} ${b.y}, ${b.x} ${b.y}`
  }

  const hoverInfo = hover ? pointOf.get(hover) : null

  return (
    <div className="relative">
      <svg viewBox={`0 0 ${VW} ${VH}`} className="w-full select-none rounded-2xl bg-slate-50 ring-1 ring-slate-200">
        {/* شبكة خفيفة تعطي إحساس لوح عمل */}
        <defs>
          <pattern id="simgrid" width="28" height="28" patternUnits="userSpaceOnUse">
            <path d="M28 0H0V28" fill="none" stroke="#e2e8f0" strokeWidth="1" />
          </pattern>
        </defs>
        <rect width={VW} height={VH} fill="url(#simgrid)" />

        {/* ═══ الأسلاك ═══ تنرسم قبل الأجهزة حتى تمر تحتها */}
        {wires.map((w, i) => {
          const a = pointOf.get(w.from)
          const b = pointOf.get(w.to)
          if (!a || !b) return null
          return (
            <g key={`${w.from}-${w.to}-${i}`} className={readOnly ? '' : 'cursor-pointer'}>
              {/* خط شفاف عريض يسهّل الضغط — الخيط الرفيع صعب تصيبه */}
              <path d={pathFor(a, b)} stroke="transparent" strokeWidth={18} fill="none"
                onClick={() => !readOnly && onRemoveWire(w)} />
              <path d={pathFor(a, b)} stroke={a.t.colorHex || '#334155'} strokeWidth={4}
                fill="none" strokeLinecap="round" opacity={0.9} pointerEvents="none" />
            </g>
          )
        })}

        {/* ═══ الأجهزة ═══ */}
        {placed.map((p) => (
          <g key={p.ref}>
            <rect x={p.x} y={p.y} width={p.w} height={p.h} rx={18}
              fill="#ffffff" stroke="#cbd5e1" strokeWidth={2} />
            <rect x={p.x} y={p.y} width={p.w} height={38} rx={18} fill="#0f2040" />
            <rect x={p.x} y={p.y + 20} width={p.w} height={18} fill="#0f2040" />
            <text x={p.x + p.w / 2} y={p.y + 25} textAnchor="middle"
              fill="#ffffff" fontSize={14} fontWeight="700">{p.device.name}</text>
            <text x={p.x + p.w / 2} y={p.y + p.h - 14} textAnchor="middle"
              fill="#94a3b8" fontSize={11}>{p.device.brand} · {p.device.model}</text>

            {(p.device.terminals ?? []).map((t) => {
              const id = `${p.ref}:${t.id}`
              const pt = pointOf.get(id)!
              const isArmed = armed === id
              const isHi = highlight?.includes(id)
              // الأطراف على الحافة اليسرى؟ نحط التسمية يمينها والعكس.
              const labelRight = t.x < 0.5
              return (
                <g key={id} onClick={() => clickTerminal(id)}
                  onMouseEnter={() => setHover(id)} onMouseLeave={() => setHover(null)}
                  className={readOnly ? '' : 'cursor-pointer'}>
                  {/* هالة التلميح أو التسليح */}
                  {(isArmed || isHi) && (
                    <circle cx={pt.x} cy={pt.y} r={16} fill={t.colorHex || '#3b82f6'} opacity={0.25}>
                      <animate attributeName="r" values="12;18;12" dur="1.4s" repeatCount="indefinite" />
                    </circle>
                  )}
                  <circle cx={pt.x} cy={pt.y} r={9} fill={t.colorHex || '#64748b'}
                    stroke={isArmed ? '#0f2040' : '#ffffff'} strokeWidth={isArmed ? 3 : 2} />
                  {/* ⚠️ علامة الخطر: الطرف الي غلطه يحرق الجهاز يتميّز
                      بصرياً قبل ما يلمسه المتدرّب. */}
                  {t.danger && (
                    <text x={pt.x} y={pt.y - 14} textAnchor="middle" fontSize={11}
                      pointerEvents="none">⚠️</text>
                  )}
                  {/* ⚠️ المرساة **معكوسة** عن المتوقع لأن النص عربي:
                      بالنص RTL قيمة `start` تثبّت **يمين** النص فيمتد
                      لليسار فوق الدائرة، و`end` تثبّت يساره فيمتد يميناً.
                      (انكشفت بالفحص: الأسماء چانت تطلع مقطوعة «أ…».) */}
                  <text x={labelRight ? pt.x + 15 : pt.x - 15} y={pt.y + 4}
                    textAnchor={labelRight ? 'end' : 'start'} fontSize={11.5}
                    fill="#334155" pointerEvents="none">{t.label}</text>
                </g>
              )
            })}
          </g>
        ))}
      </svg>

      {/* ═══ بطاقة الطرف عند المرور ═══ خارج الـSVG حتى تنسّق بـTailwind */}
      {hoverInfo && (
        <div className="pointer-events-none absolute bottom-3 right-3 max-w-[19rem] rounded-xl bg-white/95 p-3 text-right shadow-lg ring-1 ring-slate-200 backdrop-blur">
          <p className="flex items-center justify-end gap-2 text-sm font-bold text-slate-800">
            {hoverInfo.t.label}
            <span className="inline-block h-3 w-3 rounded-full ring-1 ring-slate-300"
              style={{ backgroundColor: hoverInfo.t.colorHex || '#64748b' }} />
          </p>
          <p className="mt-0.5 text-[11px] font-semibold text-brand-700">
            {TERMINAL_KIND_LABELS[hoverInfo.t.kind] || hoverInfo.t.kind}
            {hoverInfo.t.signal ? ` · ${hoverInfo.t.signal}` : ''}
          </p>
          {hoverInfo.t.description && (
            <p className="mt-1 text-[11.5px] leading-relaxed text-slate-600">{hoverInfo.t.description}</p>
          )}
          {hoverInfo.t.danger && (
            <p className="mt-1.5 rounded-lg bg-red-50 px-2 py-1 text-[11px] font-bold leading-relaxed text-red-700">
              ⚠️ {hoverInfo.t.danger}
            </p>
          )}
        </div>
      )}

      {armed && (
        <p className="absolute left-3 top-3 rounded-lg bg-brand-600 px-3 py-1.5 text-[11px] font-bold text-white shadow">
          اختار الطرف الثاني — أو Esc للإلغاء
        </p>
      )}
    </div>
  )
}
