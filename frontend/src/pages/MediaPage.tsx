import { useEffect, useState } from 'react'
import { api, type MediaBrief } from '../api'
import { useSession } from '../session'

// ═══ الإعلام والعلاقات العامة — طلب (ع) 10-05 ═══
// المشاريع الي توصل مرحلة «📸 الإعلام» قبل التنفيذ تتحوّل هنا بكل تفاصيلها:
// منو المهندس المشرف، شنو المشروع، وين الموقع، متى يبدي وشكد ياخذ ومتى
// يخلص — حتى فريق الإعلام يطلع يصوّر الشغل وينشره.

const STATUS: Record<MediaBrief['status'], { label: string; cls: string }> = {
  NEW: { label: '🆕 جديد', cls: 'bg-pink-100 text-pink-800' },
  SCHEDULED: { label: '📅 انحدد موعد التصوير', cls: 'bg-sky-100 text-sky-800' },
  SHOT: { label: '🎬 صوّرنا', cls: 'bg-amber-100 text-amber-800' },
  PUBLISHED: { label: '✅ انتشر', cls: 'bg-emerald-100 text-emerald-800' },
  CANCELLED: { label: '✖ انلغى', cls: 'bg-slate-100 text-slate-600' },
}
const dt = (s: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' }) : '—')
const d = (s: string | null) => (s ? new Date(s).toLocaleDateString('ar-IQ') : '—')

function BriefCard({ b, canEdit, onSaved }: { b: MediaBrief; canEdit: boolean; onSaved: (n: MediaBrief) => void }) {
  const [shootAt, setShootAt] = useState('')
  const [url, setUrl] = useState(b.publishedUrl ?? '')
  const [notes, setNotes] = useState(b.mediaNotes ?? '')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const map = b.locationUrl || (b.lat != null && b.lng != null ? `https://maps.google.com/?q=${b.lat},${b.lng}` : null)

  const save = async (status: MediaBrief['status']) => {
    setBusy(true); setErr('')
    try {
      onSaved(await api.updateMediaBrief(b.id, { status, shootAt: status === 'SCHEDULED' ? shootAt : b.shootAt, publishedUrl: url || null, mediaNotes: notes || null }))
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <p className="text-base font-extrabold text-slate-800">{b.projectName} <span className="text-xs font-normal text-slate-500">· {b.projectCode}</span></p>
          {b.workType && <p className="text-xs text-slate-500">نوع الشغل: {b.workType}</p>}
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-bold ${STATUS[b.status].cls}`}>{STATUS[b.status].label}</span>
      </div>

      <div className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
        <p>👷 <b>المهندس المشرف:</b> {b.engineerName ?? '—'}
          {b.engineerPhone && <> · <a href={`tel:${b.engineerPhone}`} className="font-bold text-sky-700 underline" dir="ltr">{b.engineerPhone}</a></>}</p>
        <p>📍 <b>الموقع:</b> {b.location ?? '—'} {map && <a href={map} target="_blank" rel="noreferrer" className="font-bold text-sky-700 underline">الخريطة</a>}</p>
        <p>🚀 <b>يبدي الشغل:</b> {dt(b.startAt)}</p>
        <p>⏱️ <b>المدة:</b> {b.duration ?? '—'}</p>
        <p>🏁 <b>يخلص تقريباً:</b> {b.expectedEndAt ? d(b.expectedEndAt) : (b.deliveryDate ?? '—')}</p>
        <p>📌 <b>مرحلة المشروع هسه:</b> {b.stage}</p>
      </div>
      {(b.notes || b.projectTask) && (
        <p className="mt-2 rounded-lg bg-slate-50 p-2 text-xs text-slate-700">📝 {b.notes ?? b.projectTask}</p>
      )}
      <p className="mt-2 text-[11px] text-slate-400">حوّله {b.createdByName ?? '—'} · {d(b.createdAt)}{b.mediaEmployeeName && <> · آخر تحديث من {b.mediaEmployeeName}</>}</p>

      {b.shootAt && <p className="mt-2 text-sm">📅 <b>موعد التصوير:</b> {dt(b.shootAt)}</p>}
      {b.publishedUrl && <p className="mt-1 text-sm">🔗 <a href={b.publishedUrl} target="_blank" rel="noreferrer" className="font-bold text-sky-700 underline">المنشور</a></p>}

      {canEdit && b.status !== 'PUBLISHED' && b.status !== 'CANCELLED' && (
        <div className="mt-3 space-y-2 border-t border-slate-100 pt-3">
          {b.status === 'NEW' && (
            <div className="flex flex-wrap items-center gap-2">
              <input type="datetime-local" value={shootAt} onChange={(e) => setShootAt(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1.5 text-sm" />
              <button type="button" disabled={busy || !shootAt} onClick={() => void save('SCHEDULED')} className="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">📅 حدد موعد التصوير</button>
            </div>
          )}
          {b.status === 'SCHEDULED' && (
            <button type="button" disabled={busy} onClick={() => void save('SHOT')} className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">🎬 صوّرنا</button>
          )}
          {b.status === 'SHOT' && (
            <div className="flex flex-wrap items-center gap-2">
              <input value={url} onChange={(e) => setUrl(e.target.value)} placeholder="رابط المنشور" dir="ltr" className="min-w-0 flex-1 rounded-lg border border-slate-200 px-2 py-1.5 text-sm" />
              <button type="button" disabled={busy || !url.trim()} onClick={() => void save('PUBLISHED')} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">✅ نشرنا</button>
            </div>
          )}
          <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} placeholder="ملاحظات الإعلام (اختياري)" className="w-full rounded-lg border border-slate-200 p-2 text-xs" />
          <button type="button" disabled={busy} onClick={() => { if (window.confirm('تلغي تصوير هالمشروع؟')) void save('CANCELLED') }} className="text-xs text-slate-500 hover:text-red-600">✖ إلغاء</button>
          {err && <p className="text-xs text-red-600">{err}</p>}
        </div>
      )}
    </div>
  )
}

export default function MediaPage() {
  const { employee, permissions } = useSession()
  const canEdit = employee?.role === 'ADMIN' || employee?.role === 'MEDIA' || permissions.includes('media')
  const [rows, setRows] = useState<MediaBrief[] | null>(null)
  const [err, setErr] = useState('')
  const [showDone, setShowDone] = useState(false)
  useEffect(() => { void api.getMediaBriefs().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  const list = (rows ?? []).filter((b) => showDone || (b.status !== 'PUBLISHED' && b.status !== 'CANCELLED'))

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">📸 الإعلام والعلاقات العامة</h2>
          <p className="text-sm text-slate-500">المشاريع المحوّلة للتصوير بكل تفاصيلها. حدد موعد التصوير، وبعدها «صوّرنا»، وبالأخير «نشرنا» ويا الرابط.</p>
        </div>
        <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={showDone} onChange={(e) => setShowDone(e.target.checked)} /> اعرض المنشور والملغي</label>
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {rows === null && !err ? <p className="text-slate-400">جاري التحميل…</p> : (
        <div className="grid gap-3 lg:grid-cols-2">
          {list.map((b) => <BriefCard key={b.id} b={b} canEdit={canEdit} onSaved={(n) => setRows((r) => (r ?? []).map((x) => (x.id === n.id ? n : x)))} />)}
          {list.length === 0 && <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-400">ماكو مشاريع محوّلة للإعلام هسه.</p>}
        </div>
      )}
    </div>
  )
}
