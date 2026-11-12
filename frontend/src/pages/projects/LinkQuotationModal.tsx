import { useEffect, useState } from 'react'
import { apiRequest } from '../../api'
import type { Project } from './shared'

// بلاغ #33 (مؤمل 11-12): زبون انسوّاله عرض سعر قبل ما يصعد بالنظام كمشروع —
// هنا يختار العرض الموجود ويربطه بالمشروع بدل ما يسوي عرض ثاني.
type Candidate = { id: string; quotationNumber: string; customerName: string; customerPhone?: string | null; netTotal: number; createdAt: string; match: boolean }
type Linked = { id: string; quotationNumber: string } | null

export default function LinkQuotationModal({ project, onClose, onLinked }: { project: Project; onClose: () => void; onLinked: () => void }) {
  const [q, setQ] = useState('')
  const [rows, setRows] = useState<Candidate[] | null>(null)
  const [linked, setLinked] = useState<Linked>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    apiRequest<Linked>(`/quotation-by-project/${project.id}`).then(setLinked).catch(() => {})
  }, [project.id])

  useEffect(() => {
    const t = setTimeout(() => {
      apiRequest<Candidate[]>(`/projects/${project.id}/quotation-candidates?q=${encodeURIComponent(q.trim())}`)
        .then(r => { setRows(r); setErr('') })
        .catch(e => setErr(e instanceof Error ? e.message : 'تعذر جلب العروض'))
    }, 300)
    return () => clearTimeout(t)
  }, [project.id, q])

  const link = async (id: string) => {
    setBusy(true)
    try {
      await apiRequest(`/projects/${project.id}/link-quotation`, { method: 'POST', body: JSON.stringify({ quotationId: id }) })
      onLinked()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الربط')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4" onClick={onClose}>
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-lg max-h-[85vh] flex flex-col" onClick={e => e.stopPropagation()} data-testid="link-quotation">
        <div className="p-4 border-b flex items-center justify-between">
          <h3 className="font-bold">🔗 ربط عرض سعر موجود — {project.name}</h3>
          <button onClick={onClose} className="text-gray-500 text-xl leading-none">×</button>
        </div>
        <div className="p-4 space-y-3 overflow-y-auto">
          {linked ? (
            <p className="text-sm bg-green-50 text-green-800 rounded-lg p-3">هذا المشروع مربوط بعرض <b>{linked.quotationNumber}</b>. ما يحتاج ربط.</p>
          ) : (
            <>
              <p className="text-xs text-gray-500">تطلع عروض نفس الزبون (الرقم أو الاسم) الي بعدها مو مربوطة بمشروع. تگدر تبحث برقم العرض أو الاسم.</p>
              <input value={q} onChange={e => setQ(e.target.value)} placeholder="رقم العرض أو اسم الزبون"
                className="w-full border rounded-lg px-3 py-2 text-sm" />
              {err && <p className="text-sm text-red-600">{err}</p>}
              {rows === null ? <p className="text-sm text-gray-400">جاري التحميل…</p> : rows.length === 0 ? (
                <p className="text-sm text-gray-500">ماكو عرض مطابق. ابحث برقم العرض.</p>
              ) : rows.map(r => (
                <div key={r.id} className="border rounded-lg p-3 flex items-center gap-3">
                  <div className="flex-1 min-w-0 text-sm">
                    <div className="font-bold">{r.quotationNumber} {r.match && <span className="text-xs bg-cyan-100 text-cyan-800 rounded px-1.5">نفس الزبون</span>}</div>
                    <div className="text-gray-600 truncate">{r.customerName}{r.customerPhone ? ` · ${r.customerPhone}` : ''}</div>
                    <div className="text-xs text-gray-400">{Math.round(r.netTotal).toLocaleString('en-US')} · {new Date(r.createdAt).toLocaleDateString('ar-IQ')}</div>
                  </div>
                  <button disabled={busy} onClick={() => void link(r.id)}
                    className="text-sm px-3 py-1.5 rounded-lg bg-cyan-600 text-white font-bold disabled:opacity-50">ربط</button>
                </div>
              ))}
            </>
          )}
        </div>
      </div>
    </div>
  )
}
