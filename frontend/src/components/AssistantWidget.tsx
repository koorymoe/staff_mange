import { useState } from 'react'
import { api } from '../api'

type Msg = { role: 'user' | 'assistant'; text: string }

export default function AssistantWidget() {
  const [open, setOpen] = useState(false)
  const [messages, setMessages] = useState<Msg[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)

  const send = async () => {
    const text = input.trim()
    if (!text || loading) return
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', text }])
    setLoading(true)
    try {
      const res = await api.askAssistant(text)
      setMessages((prev) => [...prev, { role: 'assistant', text: res.reply }])
    } catch (err) {
      setMessages((prev) => [...prev, { role: 'assistant', text: err instanceof Error ? err.message : 'صار خطأ، جرب مرة ثانية' }])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed bottom-5 left-5 z-[65]">
      {open && (
        <div className="mb-3 flex h-[480px] w-80 flex-col overflow-hidden rounded-2xl border border-white bg-white shadow-2xl">
          <div className="flex items-center justify-between bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 text-white">
            <span className="font-bold">🤖 المساعد الذكي</span>
            <button onClick={() => setOpen(false)} className="text-white/80 hover:text-white">✕</button>
          </div>
          <div className="flex-1 space-y-2 overflow-y-auto p-3">
            {messages.length === 0 && (
              <p className="p-3 text-center text-xs text-slate-400">
                اسألني عن راتبك، مهاراتك المتبقية، أو أي شي يخص ملفك الشخصي.
              </p>
            )}
            {messages.map((m, i) => (
              <div
                key={i}
                className={`max-w-[85%] rounded-xl px-3 py-2 text-sm ${
                  m.role === 'user' ? 'mr-auto bg-brand-500 text-white' : 'ml-auto bg-slate-100 text-slate-700'
                }`}
              >
                {m.text}
              </div>
            ))}
            {loading && <div className="ml-auto max-w-[85%] rounded-xl bg-slate-100 px-3 py-2 text-sm text-slate-400">...</div>}
          </div>
          <div className="flex gap-2 border-t border-slate-100 p-3">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && send()}
              placeholder="اكتب سؤالك..."
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            <button
              onClick={send}
              disabled={loading}
              className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50"
            >
              إرسال
            </button>
          </div>
        </div>
      )}
      <button
        onClick={() => setOpen(!open)}
        className="flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-brand-800 text-2xl text-white shadow-xl transition-transform hover:scale-105"
      >
        🤖
      </button>
    </div>
  )
}
