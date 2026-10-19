import { useRef, useState } from 'react'

// ═══ تغطية صور الموظفين قبل ما يطلع السكرين لهايكو ═══
// القيد المكتوب: صور الموظفين ما تنرسل لأي مزوّد خارجي. السكرين يلصقه
// المستخدم من جهازه فالنظام ما يعرف وين الصور بيه، والكشف بالخادم يعني
// نرسل الصورة أصلاً. فالتغطية يدوية وبالمتصفح: الي يطلع هي المغطّاة بس.

type Box = { x: number; y: number; w: number; h: number } // نسب من ٠ إلى ١

function render(src: string, boxes: Box[]): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => {
      const c = document.createElement('canvas')
      c.width = img.width
      c.height = img.height
      const ctx = c.getContext('2d')
      if (!ctx) { reject(new Error('المتصفح ما يدعم الرسم')); return }
      ctx.drawImage(img, 0, 0)
      ctx.fillStyle = '#0f2040'
      for (const b of boxes) ctx.fillRect(b.x * c.width, b.y * c.height, b.w * c.width, b.h * c.height)
      resolve(c.toDataURL('image/jpeg', 0.85))
    }
    img.onerror = () => reject(new Error('الصورة ما انقرت'))
    img.src = src
  })
}

export default function ScreenshotRedactor({ src, onDone, onCancel }: {
  src: string
  onDone: (redacted: string) => void
  onCancel: () => void
}) {
  const [boxes, setBoxes] = useState<Box[]>([])
  const [drag, setDrag] = useState<{ x0: number; y0: number; x1: number; y1: number } | null>(null)
  const [confirmed, setConfirmed] = useState(false)
  const [err, setErr] = useState('')
  const area = useRef<HTMLDivElement>(null)

  const pos = (e: React.PointerEvent) => {
    const r = area.current!.getBoundingClientRect()
    return {
      x: Math.min(1, Math.max(0, (e.clientX - r.left) / r.width)),
      y: Math.min(1, Math.max(0, (e.clientY - r.top) / r.height)),
    }
  }
  const norm = (d: { x0: number; y0: number; x1: number; y1: number }): Box => ({
    x: Math.min(d.x0, d.x1), y: Math.min(d.y0, d.y1), w: Math.abs(d.x1 - d.x0), h: Math.abs(d.y1 - d.y0),
  })

  const approve = async () => {
    try { onDone(await render(src, boxes)) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر التجهيز') }
  }

  const shown = drag ? [...boxes, norm(drag)] : boxes

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div dir="rtl" className="flex max-h-full w-full max-w-3xl flex-col gap-3 overflow-y-auto rounded-2xl bg-white p-4 shadow-xl">
        <div>
          <p className="text-base font-extrabold text-[#0f2040]">🛡️ غطِّ صور الموظفين قبل الإرسال</p>
          <p className="mt-1 text-xs text-slate-600">
            السكرين يروح لنموذج خارجي، وصور الموظفين ما تطلع برّا الشركة. اسحب مستطيل على كل صورة موظف ظاهرة —
            التغطية تصير بجهازك، والي ينرسل هي النسخة المغطّاة بس.
          </p>
        </div>
        <div
          ref={area}
          data-testid="redact-area"
          className="relative cursor-crosshair touch-none select-none self-center"
          onPointerDown={(e) => { e.currentTarget.setPointerCapture(e.pointerId); const p = pos(e); setDrag({ x0: p.x, y0: p.y, x1: p.x, y1: p.y }) }}
          onPointerMove={(e) => { if (drag) { const p = pos(e); setDrag({ ...drag, x1: p.x, y1: p.y }) } }}
          onPointerUp={() => {
            if (drag) { const b = norm(drag); if (b.w > 0.01 && b.h > 0.01) setBoxes((s) => [...s, b]) }
            setDrag(null)
          }}>
          <img src={src} alt="السكرين" draggable={false} className="block max-h-[60vh] max-w-full rounded-lg border" />
          {shown.map((b, i) => (
            <div key={i} className="pointer-events-none absolute bg-[#0f2040]"
              style={{ left: `${b.x * 100}%`, top: `${b.y * 100}%`, width: `${b.w * 100}%`, height: `${b.h * 100}%` }} />
          ))}
        </div>
        <div className="flex flex-wrap items-center gap-2 text-xs">
          <span className="text-slate-500">مغطّى: {boxes.length}</span>
          <button type="button" disabled={!boxes.length} onClick={() => setBoxes((s) => s.slice(0, -1))}
            className="rounded-lg border px-2 py-1 text-slate-700 disabled:opacity-40">↩ تراجع</button>
        </div>
        <label className="flex items-center gap-2 rounded-lg bg-amber-50 px-3 py-2 text-sm font-bold text-amber-900">
          <input id="redact-confirm" type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} />
          تأكدت ماكو صورة موظف ظاهرة بالسكرين
        </label>
        {err && <p className="text-xs font-bold text-red-700">{err}</p>}
        <div className="flex justify-end gap-2">
          <button type="button" onClick={onCancel} className="rounded-xl border px-4 py-2 text-sm text-slate-600">إلغاء</button>
          <button type="button" disabled={!confirmed} onClick={() => void approve()}
            className="rounded-xl bg-gradient-to-l from-violet-600 to-fuchsia-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-40">اعتمد الصورة</button>
        </div>
      </div>
    </div>
  )
}
