import { useEffect, useRef, useState } from 'react'
import { speakText } from '../utils/speak'
import { notify } from '../ui/notify'

// ═══ 🔊 كمال يحچي (دازاي 11-17) ═══
// زر يقرا رد كمال بصوت عراقي. وإذا «يقرا الردود تلقائياً» شغّال، الرد الجديد ينقرا لحاله.

export default function SpeakButton({ text, auto }: { text: string; auto?: boolean }) {
  const [busy, setBusy] = useState(false)
  const done = useRef(false)
  const play = async () => {
    if (busy || !text.trim()) return
    setBusy(true)
    try { await speakText(text) } catch (e) { notify(e instanceof Error ? e.message : 'تعذّر الصوت') } finally { setBusy(false) }
  }
  useEffect(() => {
    if (!auto || done.current || !text.trim()) return
    done.current = true
    speakText(text).catch(() => { /* التشغيل التلقائي ممكن المتصفح يمنعه — الزر يبقى */ })
  }, [auto, text])
  return (
    <button type="button" onClick={() => void play()} disabled={busy} title="اسمع الرد"
      className="mt-1 rounded-lg px-2 py-0.5 text-xs text-slate-500 hover:bg-slate-100 disabled:opacity-50">
      {busy ? '⏳' : '🔊'} اسمع
    </button>
  )
}
