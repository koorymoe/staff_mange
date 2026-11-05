import { useCallback, useEffect, useState } from 'react'
import { api, fileUrl, type BotReport } from '../api'
import { roleLabel } from '../roleLabels'
import { useSession } from '../session'
import { askConfirm } from '../ui/dialogs'

// 📩 بلاغات الموظفين من بوت التلغرام (دازاي 11-05) — للمالك ومن ينطيه الصلاحية.
// كل تغيير حالة أو رد يوصل للموظف بالبوت فوراً.
const ORDER = ['NEW', 'NEEDS_INFO', 'IN_PROGRESS', 'DONE', 'NOT_A_BUG', 'REJECTED']

export default function BotReportsPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [view, setView] = useState<'reports' | 'links'>('reports')
  const [status, setStatus] = useState('NEW')
  const [data, setData] = useState<{ items: BotReport[]; statuses: Record<string, string> } | null>(null)
  const [err, setErr] = useState('')
  const [open, setOpen] = useState<string | null>(null)

  const load = useCallback(() => {
    api.getBotReports(status).then((d) => { setErr(''); setData(d) }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب البلاغات'))
  }, [status])
  useEffect(() => { load() }, [load])

  const labels = data?.statuses ?? {}
  return (
    <div dir="rtl" className="mx-auto max-w-5xl space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">📩 بلاغات الموظفين</h1>
        <p className="mt-1 text-sm text-blue-100">مشاكل النظام الي دزها الموظفين لبوت التلغرام (صورة + شرح). أي حالة أو رد تحطه يوصل للموظف بالبوت.</p>
      </div>
      {isOwner && (
        <div className="flex gap-2">
          {([['reports', '📩 البلاغات'], ['links', '🔗 الحسابات المربوطة']] as const).map(([k, l]) => (
            <button key={k} type="button" onClick={() => setView(k)}
              className={`rounded-xl px-4 py-2 text-sm font-extrabold ${view === k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200'}`}>{l}</button>
          ))}
        </div>
      )}
      {view === 'links' && isOwner ? <LinksView /> : <>
      <div className="flex flex-wrap gap-2">
        {[...ORDER, ''].map((s) => (
          <button key={s || 'all'} type="button" onClick={() => setStatus(s)}
            className={`rounded-xl px-3 py-1.5 text-sm font-bold ${status === s ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
            {s ? (labels[s] ?? s) : 'الكل'}
          </button>
        ))}
      </div>
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}
      {!data && !err && <p className="py-10 text-center text-slate-400">جاري التحميل…</p>}
      {data && data.items.length === 0 && <p className="py-10 text-center text-slate-400">ماكو بلاغات بهالحالة.</p>}
      <div className="space-y-3">
        {data?.items.map((r) => (
          <ReportCard key={r.id} r={r} labels={labels} expanded={open === r.id}
            onToggle={() => setOpen(open === r.id ? null : r.id)} onSaved={load} />
        ))}
      </div>
      </>}
    </div>
  )
}

// الحسابات المربوطة بالتلغرام — الفك بس من هنا (دازاي 11-05: «محد يفك الربط الا اني»)
function LinksView() {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof api.getTelegramLinks>> | null>(null)
  const [err, setErr] = useState('')
  const load = useCallback(() => { api.getTelegramLinks().then((r) => { setErr(''); setRows(r) }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب')) }, [])
  useEffect(() => { load() }, [load])
  const unlink = async (id: string, name: string) => {
    if (!(await askConfirm(`تفك ربط ${name} بالتلغرام؟ يوصله إشعار، ويگدر يربط من جديد.`))) return
    try { await api.ownerUnlinkTelegram(id); load() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر فك الربط') }
  }
  if (err) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (!rows.length) return <p className="py-10 text-center text-slate-400">ماكو أحد رابط حسابه بعد.</p>
  return (
    <div className="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200">
      <table className="w-full text-sm">
        <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="p-3 text-right">الموظف</th><th className="p-3 text-right">التلغرام</th><th className="p-3 text-right">من</th><th /></tr></thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.employeeId} className="border-t border-slate-100">
              <td className="p-3 font-bold">{r.name} <span className="text-xs font-normal text-slate-500">({roleLabel(r.role) || r.role}){r.status !== 'ACTIVE' ? ' · مو فعّال' : ''}</span></td>
              <td className="p-3" dir="ltr">{r.tgUsername ? `@${r.tgUsername}` : '—'}</td>
              <td className="p-3 text-xs text-slate-500">{new Date(r.linkedAt).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</td>
              <td className="p-3 text-left"><button onClick={() => unlink(r.employeeId, r.name)} className="text-xs font-bold text-red-600 hover:underline">فك الربط</button></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function ReportCard({ r, labels, expanded, onToggle, onSaved }: {
  r: BotReport; labels: Record<string, string>; expanded: boolean; onToggle: () => void; onSaved: () => void
}) {
  const [st, setSt] = useState(r.status)
  const [reply, setReply] = useState('')
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState('')
  const save = async () => {
    setSaving(true); setMsg('')
    try {
      await api.updateBotReport(r.id, { status: st, reply: reply.trim() || undefined })
      setReply(''); setMsg('انحفظ ووصل للموظف ✅'); onSaved()
    } catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setSaving(false) }
  }
  const when = new Date(r.createdAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'medium', timeStyle: 'short' })
  return (
    <div className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
      <button type="button" onClick={onToggle} className="flex w-full flex-wrap items-center justify-between gap-2 text-right">
        <span className="font-extrabold text-[#0f2040]">#{r.number} · {r.employeeName} <span className="text-xs font-normal text-slate-500">({roleLabel(r.employeeRole) || r.employeeRole})</span></span>
        <span className="text-xs text-slate-500">{r.source === 'system' ? '🖥️ النظام' : '📱 البوت'} · {labels[r.status] ?? r.status} · {when} · 📷 {r.photos.length}</span>
      </button>
      {r.page && <p className="mt-1 text-xs text-slate-400">الشاشة: <bdi dir="ltr" className="font-mono">{r.page}</bdi></p>}
      <p className={`mt-2 whitespace-pre-wrap text-sm text-slate-700 ${expanded ? '' : 'line-clamp-2'}`}>{r.body}</p>
      {expanded && (
        <div className="mt-3 space-y-3">
          <div className="flex flex-wrap gap-2">
            {r.photos.map((p, i) => (
              <a key={i} href={fileUrl(p)} target="_blank" rel="noreferrer">
                <img src={fileUrl(p)} alt={`صورة ${i + 1}`} className="h-32 w-auto rounded-lg border object-contain" />
              </a>
            ))}
          </div>
          {r.adminNote && <p className="rounded-lg bg-slate-50 p-2 text-xs text-slate-600">آخر رد: {r.adminNote}</p>}
          <div className="flex flex-col gap-2 sm:flex-row">
            <select value={st} onChange={(e) => setSt(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
              {ORDER.map((s) => <option key={s} value={s}>{labels[s] ?? s}</option>)}
            </select>
            <input value={reply} onChange={(e) => setReply(e.target.value)} maxLength={2000}
              placeholder="رد يوصل للموظف بالبوت (اختياري)" className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm" />
            <button type="button" onClick={save} disabled={saving || (st === r.status && !reply.trim())}
              className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-40">
              {saving ? 'لحظة…' : 'حفظ وإرسال'}
            </button>
          </div>
          {msg && <p className="text-xs text-slate-600">{msg}</p>}
        </div>
      )}
    </div>
  )
}
