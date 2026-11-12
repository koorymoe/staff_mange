import { useState } from 'react'
import { apiRequest, thumbUrl } from '../../api'
import type { PlanFileRow } from './planTypes'
import { sheetDate } from '../../utils/planSetPrint'

// ═══ فورما المخطط — دازاي 11-12 ═══
// «جوه كل تصميم لازم اكو فورما: لوغو الأماني، اسم المخطط، الملاحظات، السكيل… لو ع اليمين
// لو جوه خلي المصممه هيه تحدد… لازم يظهر التاريخ واللوغو، والباقي هيه تعدل عليهن».
// اللوغو والتاريخ ثابتين؛ المكان والاسم والملاحظات والسكيل بيد المصممة. تنطبع بالـPDF.

const LOGO = `${import.meta.env.BASE_URL}apple-touch-icon.png`
const NAVY = '#2b3074'

export default function PlanSheetEditor({ file, typeName, onClose, onSaved }: {
  file: PlanFileRow
  typeName: string
  onClose: () => void
  onSaved: () => void
}) {
  const [pos, setPos] = useState<'right' | 'bottom'>(file.sheetPos ?? 'right')
  const [name, setName] = useState(file.sheetName ?? typeName)
  const [notes, setNotes] = useState(file.sheetNotes ?? '')
  const [scale, setScale] = useState(file.sheetScale ?? '')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const save = async () => {
    setBusy(true); setErr('')
    try {
      await apiRequest(`/plan-files/${file.id}/sheet`, { method: 'PUT', body: JSON.stringify({ pos, name, notes, scale }) })
      onSaved()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const block = (
    <div className={`flex ${pos === 'right' ? 'h-full w-28 flex-col' : 'w-full'} overflow-hidden rounded border-2 bg-white text-[9px] leading-tight`} style={{ borderColor: NAVY }} data-testid="sheet-preview">
      <div className={`flex items-center justify-center p-1.5 ${pos === 'right' ? 'border-b' : 'border-l'}`} style={{ borderColor: NAVY }}>
        <img src={LOGO} alt="الأماني" className="h-8 w-8 rounded-full object-contain" />
      </div>
      {[['اسم المخطط', name || '—'], ['الملاحظات', notes || '—'], ['السكيل', scale || '—'], ['التاريخ', sheetDate(file)]].map(([k, v]) => (
        <div key={k} className={`flex-1 p-1 ${pos === 'right' ? 'border-b last:border-b-0' : 'border-l last:border-l-0'}`} style={{ borderColor: NAVY }}>
          <p className="font-bold" style={{ color: NAVY }}>{k}</p>
          <p className="whitespace-pre-wrap break-words text-slate-700">{v}</p>
        </div>
      ))}
    </div>
  )
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4" onClick={onClose}>
      <div dir="rtl" className="w-full max-w-3xl rounded-2xl bg-white p-5 shadow-2xl" onClick={(e) => e.stopPropagation()} data-testid="sheet-editor">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-lg font-extrabold" style={{ color: NAVY }}>🗂️ فورما المخطط — {typeName}</h3>
          <button type="button" onClick={onClose} className="text-2xl text-slate-400" aria-label="سد">×</button>
        </div>
        <div className="grid gap-4 md:grid-cols-[1fr_1.2fr]">
          <div className="space-y-3 text-sm">
            <fieldset>
              <legend className="mb-1 font-bold text-slate-700">مكان الفورما</legend>
              <div className="flex gap-2">
                {([['right', '⇥ على اليمين'], ['bottom', '⤓ جوّه (تحت)']] as const).map(([v, l]) => (
                  <button key={v} type="button" onClick={() => setPos(v)}
                    className={`flex-1 rounded-lg border px-3 py-2 font-bold ${pos === v ? 'border-amber-500 bg-amber-50 text-amber-800' : 'border-slate-200 text-slate-600'}`}>{l}</button>
                ))}
              </div>
            </fieldset>
            <label className="block"><span className="font-bold text-slate-700">اسم المخطط</span>
              <input value={name} onChange={(e) => setName(e.target.value)} maxLength={120} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
            <label className="block"><span className="font-bold text-slate-700">الملاحظات</span>
              <textarea value={notes} onChange={(e) => setNotes(e.target.value)} maxLength={600} rows={3} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
            <label className="block"><span className="font-bold text-slate-700">السكيل</span>
              <input value={scale} onChange={(e) => setScale(e.target.value)} maxLength={40} placeholder="مثلاً 1:100" dir="ltr" className="mt-1 w-full rounded-lg border border-slate-200 p-2 text-left" /></label>
            <p className="rounded-lg bg-slate-50 p-2 text-xs text-slate-500">🔒 لوغو الأماني والتاريخ ({sheetDate(file)}) يطلعون دائماً بالفورما.</p>
            {err && <p className="text-xs text-red-600">{err}</p>}
            <div className="flex gap-2">
              <button type="button" disabled={busy} onClick={() => void save()} className="rounded-lg px-5 py-2 text-sm font-bold text-white disabled:opacity-50" style={{ background: NAVY }}>{busy ? 'يحفظ…' : 'احفظ الفورما'}</button>
              <button type="button" onClick={onClose} className="text-sm text-slate-500">إلغاء</button>
            </div>
          </div>
          <div>
            <p className="mb-1 text-xs font-bold text-slate-500">معاينة — هيچ تطلع بالـPDF</p>
            <div className={`flex aspect-[1.414/1] gap-1.5 rounded-lg border border-slate-300 bg-slate-50 p-2 ${pos === 'right' ? 'flex-row' : 'flex-col'}`}>
              {pos === 'right' && block}
              <div className="grid min-h-0 flex-1 place-items-center overflow-hidden rounded bg-white">
                <img src={thumbUrl(file.fileKey, 480)} alt="" className="max-h-full max-w-full object-contain" />
              </div>
              {pos === 'bottom' && block}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
