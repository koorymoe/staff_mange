import { useEffect, useRef, useState } from 'react'
import { api } from '../api'

// ═══ فويس مسج للإنجاز — طلب (ع) 10-05 ═══
// يسجّل من المايك (لحد ٣ دقايق) ويبقى بسيرفرنا بس — ما يطلع لأي مزوّد.

const MAX_SECONDS = 180

export function VoiceRecorder({ value, onChange }: { value: { blob: Blob; seconds: number } | null; onChange: (v: { blob: Blob; seconds: number } | null) => void }) {
  const [rec, setRec] = useState<MediaRecorder | null>(null)
  const [secs, setSecs] = useState(0)
  const [err, setErr] = useState('')
  const [url, setUrl] = useState('')
  const startRef = useRef(0)
  const timer = useRef(0)

  useEffect(() => {
    if (!value) return
    const u = URL.createObjectURL(value.blob)
    queueMicrotask(() => setUrl(u))
    return () => URL.revokeObjectURL(u)
  }, [value])

  const stop = () => { rec?.stop(); window.clearInterval(timer.current) }

  const start = async () => {
    setErr('')
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const r = new MediaRecorder(stream)
      const chunks: Blob[] = []
      r.ondataavailable = (e) => { if (e.data.size) chunks.push(e.data) }
      r.onstop = () => {
        stream.getTracks().forEach((t) => t.stop())
        const seconds = (Date.now() - startRef.current) / 1000
        onChange({ blob: new Blob(chunks, { type: r.mimeType }), seconds })
        setRec(null)
      }
      startRef.current = Date.now()
      setSecs(0)
      r.start()
      setRec(r)
      timer.current = window.setInterval(() => {
        const s = Math.floor((Date.now() - startRef.current) / 1000)
        setSecs(s)
        if (s >= MAX_SECONDS) { r.stop(); window.clearInterval(timer.current) }
      }, 500)
    } catch {
      setErr('ما گدرنا نفتح المايك — اسمح للمتصفح يستعمله.')
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      {rec ? (
        <button type="button" onClick={stop} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-bold text-white">
          ⏹️ وقّف ({secs} ث)
        </button>
      ) : (
        <button type="button" onClick={() => void start()} className="rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-bold text-slate-700 hover:bg-slate-50">
          🎙️ {value ? 'سجّل من جديد' : 'سجّل فويس'}
        </button>
      )}
      {value && !rec && (
        <>
          {url && <audio controls src={url} className="h-9 max-w-full" />}
          <button type="button" onClick={() => onChange(null)} className="text-xs text-red-600">✖ احذف</button>
        </>
      )}
      {err && <span className="text-xs text-red-600">{err}</span>}
      <span className="w-full text-[11px] text-slate-400">الفويس يبقى بسيرفر الشركة ويسمعه المدير بس (لحد ٣ دقايق).</span>
    </div>
  )
}

export function VoicePlayer({ achievementId, seconds }: { achievementId: string; seconds?: number | null }) {
  const [url, setUrl] = useState('')
  const [busy, setBusy] = useState(false)
  useEffect(() => () => { if (url) URL.revokeObjectURL(url) }, [url])
  if (url) return <audio controls autoPlay src={url} className="mt-2 h-9 max-w-full" />
  return (
    <button type="button" disabled={busy} onClick={async () => {
      setBusy(true)
      try { setUrl(URL.createObjectURL(await api.getAchievementVoice(achievementId))) } catch { /* ماكو */ } finally { setBusy(false) }
    }} className="mt-2 rounded-lg border border-slate-200 px-3 py-1 text-xs font-bold text-slate-700 hover:bg-slate-50">
      {busy ? '…' : `▶️ اسمع الفويس${seconds ? ` (${seconds} ث)` : ''}`}
    </button>
  )
}
