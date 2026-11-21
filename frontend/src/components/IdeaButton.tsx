import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import Modal from '../ui/Modal'
import { apiRequest } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'
import { IDEA_CATEGORIES, IDEA_STATUS, type Idea } from './ideaTypes'

// 💡 عندي فكرة (دازاي 11-21 — كايزن): الموظف يقترح تحسين لشغله، ومدير إدارته يرد خلال ٤٨ ساعة.
// الفكرة الي «نطبقها» تنحسب إله شغل إضافي بتقييمه.
export default function IdeaButton() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <button type="button" onClick={() => setOpen(true)} title="عندي فكرة تحسّن الشغل"
        className="flex h-10 items-center gap-1 rounded-full px-3 text-sm font-bold text-amber-600 transition-all hover:bg-amber-50">
        💡 <span className="hidden sm:inline">فكرة</span>
      </button>
      {open && createPortal(<IdeaModal onClose={() => setOpen(false)} />, document.body)}
    </>
  )
}

function IdeaModal({ onClose }: { onClose: () => void }) {
  const [text, setText] = useState('')
  const [cat, setCat] = useState('OTHER')
  const [busy, setBusy] = useState(false)
  const [mine, setMine] = useState<Idea[]>([])
  const load = () => { apiRequest<Idea[]>('/me/ideas').then(setMine).catch(() => setMine([])) }
  useEffect(load, [])
  const send = async () => {
    setBusy(true)
    try {
      await apiRequest('/me/ideas', { method: 'POST', body: JSON.stringify({ text, category: cat }) })
      notify('انرسلت فكرتك لمدير إدارتك — يرد خلال ٤٨ ساعة', 'success')
      setText(''); load()
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر الإرسال') } finally { setBusy(false) }
  }
  return (
    <Modal title={<>💡 عندي فكرة</>} onClose={onClose} size="md" busy={busy}>
      <div className="space-y-3">
        <p className="text-xs text-slate-500">شي صغير يخلّي شغلك أسرع، أرخص، أحسن، أو أأمن. مدير إدارتك يرد خلال ٤٨ ساعة، والفكرة الي تنطبق تنحسبلك بتقييمك.</p>
        <div className="flex flex-wrap gap-1">
          {Object.entries(IDEA_CATEGORIES).map(([k, v]) => (
            <button key={k} type="button" onClick={() => setCat(k)}
              className={`rounded-full px-3 py-1 text-xs font-bold ${cat === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>{v}</button>
          ))}
        </div>
        <textarea value={text} onChange={(e) => setText(e.target.value)} rows={3} maxLength={2000}
          placeholder="مثلاً: نخلي عدّة الكاميرات بصندوق واحد بكل سيارة حتى ما نرجع للمخزن"
          className="w-full rounded-xl border border-slate-200 p-2 text-sm" />
        <button type="button" disabled={busy || [...text.trim()].length < 10} onClick={() => void send()}
          className="w-full rounded-xl bg-amber-500 py-2 text-sm font-bold text-white disabled:opacity-50">أرسل الفكرة</button>
        {mine.length > 0 && (
          <div className="space-y-1">
            <p className="text-xs font-extrabold text-slate-700">أفكارك</p>
            {mine.map((i) => (
              <div key={i.id} className="rounded-lg bg-slate-50 px-2 py-1.5 text-xs">
                <p>{i.text}</p>
                <p className="mt-0.5 text-slate-500">{IDEA_STATUS[i.status] ?? i.status}{i.reply && <> — «{i.reply}»</>}{i.decidedBy && <> · {i.decidedBy}</>} · {fmtDateTime(i.createdAt)}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </Modal>
  )
}
