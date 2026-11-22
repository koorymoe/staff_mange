import Modal from '../../ui/Modal'
import { useState } from 'react'
import { apiRequest, thumbUrl } from '../../api'
import type { PlanFileRow, SheetExtraField, SheetSlot } from './planTypes'
import { SHEET_SLOTS, sheetDate, sheetSections } from '../../utils/planSetPrint'
import { COMPANY_UNITED_AR, COMPANY_UNITED_SUB } from '../../company'

// ═══ فورما المخطط — دازاي 11-12 ═══
// «جوه كل تصميم لازم اكو فورما: لوغو الأماني، اسم المخطط، الملاحظات، السكيل… لو ع اليمين
// لو جوه خلي المصممه هيه تحدد… لازم يظهر التاريخ واللوغو، والباقي هيه تعدل عليهن».
// اللوغو والتاريخ ثابتين؛ المكان والاسم والملاحظات والسكيل بيد المصممة. تنطبع بالـPDF.

const LOGO = `${import.meta.env.BASE_URL}apple-touch-icon.png`
const NAVY = '#2b3074'

export default function PlanSheetEditor({ file, typeName, setId, projectName: pn0, projectAddress: pa0, packageName, canEditFields, onClose, onSaved }: {
  file: PlanFileRow
  typeName: string
  setId?: string
  projectName?: string | null
  projectAddress?: string | null
  packageName?: string
  // (ل) 11-22: صلاحية «فورما المخطط: إضافة خانات»
  canEditFields?: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const [pos, setPos] = useState<'right' | 'bottom'>(file.sheetPos ?? 'right')
  const [name, setName] = useState(file.sheetName ?? typeName)
  const [notes, setNotes] = useState(file.sheetNotes ?? '')
  const [scale, setScale] = useState(file.sheetScale ?? '')
  // (ل) 11-22 دازاي «اسم المشروع وعنوانه منين أضيفهن؟ — بس مرة وحدة»: ينحفظن للسيت كله
  const [projectName, setProjectName] = useState(pn0 ?? '')
  const [projectAddress, setProjectAddress] = useState(pa0 ?? '')
  const [extra, setExtra] = useState<SheetExtraField[]>(file.sheetExtra ?? [])
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const save = async () => {
    setBusy(true); setErr('')
    try {
      if (setId && (projectName !== (pn0 ?? '') || projectAddress !== (pa0 ?? ''))) {
        await apiRequest(`/plan-sets/${setId}`, { method: 'PUT', body: JSON.stringify({ sheetProjectName: projectName, sheetProjectAddress: projectAddress }) })
      }
      await apiRequest(`/plan-files/${file.id}/sheet`, { method: 'PUT', body: JSON.stringify({ pos, name, notes, scale, ...(canEditFields ? { extra } : {}) }) })
      onSaved()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }

  // (ل) 11-22 دازاي: المعاينة نفس فورما الـPDF (sheetSections) وبنفس النسبة: 50 ملم من 420 = 11.9٪
  const secs = sheetSections({ projectName, projectAddress, packageName, boardName: name || typeName, scale, date: sheetDate(file), notes, page: '01', extra })
  const vert = pos === 'right'
  const block = (
    <div className={`flex ${vert ? 'h-full w-[11.9%] shrink-0 flex-col' : 'w-full'} overflow-hidden rounded border-2 bg-white text-[7px] leading-tight`} style={{ borderColor: NAVY }} data-testid="sheet-preview">
      <div className="flex items-center justify-center p-1"><img src={LOGO} alt="الأماني" className="h-7 w-7 object-contain" /></div>
      {secs.map((sec, si) => sec.key === 'company'
        ? <div key={si} className={`p-1 text-center ${vert ? 'border-t' : 'border-s'}`} style={{ borderColor: NAVY }}><p className="font-bold" style={{ color: NAVY }}>{COMPANY_UNITED_AR}</p><p className="text-[5px] text-slate-600">{COMPANY_UNITED_SUB}</p></div>
        : <div key={si} className={`space-y-0.5 p-1 ${sec.grow ? 'flex-1' : ''} ${vert ? 'border-t' : 'flex-1 border-s'} ${sec.key === 'extra' ? 'bg-amber-50' : ''}`} style={{ borderColor: NAVY }}>
            {sec.rows.map(([k, v]) => <div key={k}><p className="font-bold" style={{ color: NAVY }}>{k}:</p><p className="whitespace-pre-wrap break-words text-slate-700">{v}</p></div>)}
          </div>)}
    </div>
  )
  const setX = (i: number, patch: Partial<SheetExtraField>) => setExtra((xs) => xs.map((x, k) => (k === i ? { ...x, ...patch } : x)))
  return (
    <Modal title={<>🗂️ فورما المخطط — {typeName}</>} onClose={onClose} size="xl" busy={busy}>
        <div data-testid="sheet-editor" className="grid gap-4 md:grid-cols-[1fr_1.2fr]">
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
            {setId && (
              <fieldset className="grid gap-2 rounded-lg border border-slate-200 p-2 sm:grid-cols-2" data-testid="sheet-project">
                <legend className="px-1 text-xs font-bold text-slate-500">🏢 المشروع — ينكتب مرة وحدة ويطلع بكل لوحات السيت</legend>
                <label className="block"><span className="font-bold text-slate-700">اسم المشروع</span>
                  <input value={projectName} onChange={(e) => setProjectName(e.target.value)} maxLength={200} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
                <label className="block"><span className="font-bold text-slate-700">عنوان المشروع</span>
                  <input value={projectAddress} onChange={(e) => setProjectAddress(e.target.value)} maxLength={200} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
              </fieldset>
            )}
            <label className="block"><span className="font-bold text-slate-700">اسم اللوحة</span>
              <input value={name} onChange={(e) => setName(e.target.value)} maxLength={120} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
            <label className="block"><span className="font-bold text-slate-700">الملاحظات</span>
              <textarea value={notes} onChange={(e) => setNotes(e.target.value)} maxLength={600} rows={3} className="mt-1 w-full rounded-lg border border-slate-200 p-2" /></label>
            <label className="block"><span className="font-bold text-slate-700">المقياس</span>
              <input value={scale} onChange={(e) => setScale(e.target.value)} maxLength={40} placeholder="مثلاً 1:100" dir="ltr" className="mt-1 w-full rounded-lg border border-slate-200 p-2 text-left" /></label>
            {canEditFields && (
              <fieldset className="space-y-2 rounded-lg border border-amber-200 bg-amber-50/50 p-2" data-testid="sheet-extra">
                <legend className="px-1 font-bold text-slate-700">➕ خانات إضافية</legend>
                {extra.map((x, i) => (
                  <div key={i} className="space-y-1 rounded-lg bg-white p-2 shadow-sm">
                    <div className="flex gap-1">
                      <input value={x.label} onChange={(e) => setX(i, { label: e.target.value })} maxLength={40} placeholder="اسم الخانة (مثلاً: المهندس المشرف)" className="min-w-0 flex-1 rounded border border-slate-200 p-1.5" />
                      <select value={x.after} onChange={(e) => setX(i, { after: e.target.value as SheetSlot })} className="rounded border border-slate-200 p-1.5 text-xs" aria-label="مكان الخانة">
                        {SHEET_SLOTS.map((o) => <option key={o.key} value={o.key}>{o.label}</option>)}
                      </select>
                      <button type="button" onClick={() => setExtra((xs) => xs.filter((_, k) => k !== i))} className="rounded px-2 text-rose-600 hover:bg-rose-50" aria-label="شيل الخانة">✕</button>
                    </div>
                    <textarea value={x.value} onChange={(e) => setX(i, { value: e.target.value })} maxLength={300} rows={2} placeholder="المعلومة" className="w-full rounded border border-slate-200 p-1.5" />
                  </div>
                ))}
                {extra.length < 8 && <button type="button" onClick={() => setExtra((xs) => [...xs, { label: '', value: '', after: 'notes' }])} className="rounded-lg border border-dashed border-amber-400 px-3 py-1.5 text-xs font-bold text-amber-800">+ أضف خانة</button>}
              </fieldset>
            )}
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
    </Modal>
  )
}
