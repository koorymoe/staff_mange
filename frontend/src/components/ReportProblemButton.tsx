import { useCallback, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useLocation } from 'react-router-dom'
import { api, fileUrl, type BotReport } from '../api'

// 🆘 بلّغ عن مشكلة بالنظام — من داخل النظام (دازاي 11-05). نفس شروط البوت: صورة + شرح واضح.
// الموظف معروف من دخوله (بلا ربط)، والشاشة الي كان بيها تنسجّل لحالها، والبلاغ يوصل للمالك
// بالتلغرام ويا صوره. والرد يرجعله إشعار بالنظام.
const MIN_RUNES = 25, MIN_WORDS = 5, MAX_PHOTOS = 5

export default function ReportProblemButton() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <button onClick={() => setOpen(true)} title="بلّغ عن مشكلة بالنظام"
        className="flex h-10 items-center gap-1 rounded-full px-3 text-sm font-bold text-rose-600 transition-all hover:bg-rose-50">
        🆘 <span className="hidden sm:inline">بلّغ</span>
      </button>
      {/* ⚠️ على body: الشريط العلوي يحبس fixed جوّاه (backdrop-filter) */}
      {open && createPortal(<ReportModal onClose={() => setOpen(false)} />, document.body)}
    </>
  )
}

function ReportModal({ onClose }: { onClose: () => void }) {
  const { pathname } = useLocation()
  const page = pathname.replace(/^\/staff_mange/, '') || '/'
  const [body, setBody] = useState('')
  const [photos, setPhotos] = useState<string[]>([])
  const [uploading, setUploading] = useState(false)
  const [sending, setSending] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)
  const [mine, setMine] = useState<{ items: BotReport[]; statuses: Record<string, string> } | null>(null)

  const loadMine = useCallback(() => { api.getMyBotReports().then(setMine).catch(() => setMine({ items: [], statuses: {} })) }, [])
  useEffect(() => { loadMine() }, [loadMine])

  const addFiles = async (files: File[]) => {
    const imgs = files.filter((f) => f.type.startsWith('image/')).slice(0, MAX_PHOTOS - photos.length)
    if (!imgs.length) return
    setUploading(true); setMsg(null)
    try {
      for (const f of imgs) {
        const r = await api.uploadFile(f, 'botreports')
        setPhotos((p) => [...p, r.url])
      }
    } catch (e) {
      setMsg({ ok: false, text: e instanceof Error ? e.message : 'تعذر رفع الصورة' })
    } finally { setUploading(false) }
  }
  // لصق لقطة الشاشة مباشرة (Ctrl+V)
  const onPaste = (e: React.ClipboardEvent) => {
    const files = [...e.clipboardData.items].filter((i) => i.kind === 'file').map((i) => i.getAsFile()).filter(Boolean) as File[]
    if (files.length) { e.preventDefault(); addFiles(files) }
  }

  const words = body.trim().split(/\s+/).filter(Boolean).length
  const clear = [...body.trim()].length >= MIN_RUNES && words >= MIN_WORDS
  const send = async () => {
    setSending(true); setMsg(null)
    try {
      const r = await api.submitBotReport({ body: body.trim(), photos, page })
      setMsg({ ok: true, text: `✅ وصل بلاغك رقم #${r.number}. راح يندرس ويوصلك إشعار من تتغيّر حالته. شكراً 🌷` })
      setBody(''); setPhotos([]); loadMine()
    } catch (e) {
      setMsg({ ok: false, text: e instanceof Error ? e.message : 'تعذر الإرسال' })
    } finally { setSending(false) }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" onClick={onClose}>
      <div dir="rtl" className="mt-10 w-full max-w-lg space-y-3 rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()} onPaste={onPaste}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-black text-[#0f2040]">🆘 بلّغ عن مشكلة بالنظام</h2>
          <button onClick={onClose} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
        </div>
        <p className="text-xs text-slate-500">اكتب بأي شاشة صارت، وشنو جنت تسوي، وشنو طلعلك بالضبط — ويا صورة للشاشة. الشاشة الحالية تنسجّل لحالها: <bdi dir="ltr" className="font-mono">{page}</bdi></p>
        <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={4} maxLength={4000}
          placeholder="مثلاً: بشاشة فاتورة الليدر، من أضغط «حفظ» يطلع «خطأ» وما تنحفظ الفاتورة"
          className="w-full rounded-xl border border-slate-300 p-3 text-sm outline-none focus:border-brand-500" />
        {body && !clear && <p className="text-xs text-amber-600">الشرح قصير شوية — اكتب أكثر حتى نفهم المشكلة.</p>}
        <div className="flex flex-wrap items-center gap-2">
          {photos.map((p, i) => (
            <div key={p} className="relative">
              <img src={fileUrl(p)} alt={`صورة ${i + 1}`} className="h-20 w-20 rounded-lg border object-cover" />
              <button onClick={() => setPhotos((x) => x.filter((y) => y !== p))} className="absolute -left-1 -top-1 h-5 w-5 rounded-full bg-rose-600 text-xs text-white">×</button>
            </div>
          ))}
          {photos.length < MAX_PHOTOS && (
            <label className="flex h-20 w-20 cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-slate-300 text-xs text-slate-500 hover:bg-slate-50">
              {uploading ? '…' : <>📷<span>صورة</span></>}
              <input type="file" accept="image/*" multiple className="hidden" onChange={(e) => { addFiles([...(e.target.files ?? [])]); e.target.value = '' }} />
            </label>
          )}
        </div>
        <p className="text-[11px] text-slate-400">تگدر تلصق لقطة الشاشة (Ctrl+V). غطّي أسماء الزبائن بالصورة إذا ما يحتاجها البلاغ.</p>
        {msg && <p className={`rounded-lg p-2 text-sm ${msg.ok ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>{msg.text}</p>}
        <button onClick={send} disabled={sending || uploading || !clear || photos.length === 0}
          className="w-full rounded-xl bg-[#0f2040] py-2.5 font-bold text-white disabled:opacity-40">
          {sending ? 'لحظة…' : photos.length === 0 ? 'ارفق صورة حتى ترسل' : 'إرسال البلاغ'}
        </button>
        {mine && mine.items.length > 0 && (
          <div className="border-t border-slate-100 pt-3">
            <p className="mb-2 text-xs font-bold text-slate-500">بلاغاتي</p>
            <ul className="max-h-48 space-y-1.5 overflow-y-auto">
              {mine.items.map((r) => (
                <li key={r.id} className="rounded-lg bg-slate-50 p-2 text-xs">
                  <span className="font-bold">#{r.number}</span> · {mine.statuses[r.status] ?? r.status}
                  <p className="line-clamp-1 text-slate-500">{r.body}</p>
                  {r.adminNote && <p className="text-emerald-700">↩️ {r.adminNote}</p>}
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  )
}
