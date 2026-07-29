import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

const ICON_CHOICES = ['😀', '😎', '🦁', '🐺', '🦅', '🐉', '⚡', '🔥', '⭐', '🚀', '🛠️', '🎯']

export default function SettingsPanel({ onClose }: { onClose: () => void }) {
  const { employee } = useSession()
  const [tab, setTab] = useState<'password' | 'volume' | 'icon' | 'approvals'>('password')

  // تغيير كلمة المرور
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [pwSaving, setPwSaving] = useState(false)
  const [pwMsg, setPwMsg] = useState<{ text: string; ok: boolean } | null>(null)

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setPwMsg(null)
    setPwSaving(true)
    try {
      await api.changePassword(currentPassword, newPassword)
      setPwMsg({ text: 'تم تغيير كلمة المرور بنجاح', ok: true })
      setCurrentPassword(''); setNewPassword('')
    } catch (err) {
      setPwMsg({ text: err instanceof Error ? err.message : 'حدث خطأ', ok: false })
    } finally {
      setPwSaving(false)
    }
  }

  // مستوى الصوت (يُخزّن محلياً — يستخدمه لاحقاً نظام الفيديوهات التدريبية)
  const [volume, setVolume] = useState(() => {
    const saved = localStorage.getItem('appVolume')
    return saved ? Number(saved) : 70
  })
  const handleVolumeChange = (v: number) => {
    setVolume(v)
    localStorage.setItem('appVolume', String(v))
  }

  // طلب تعديل رمز الحضور
  const [selectedIcon, setSelectedIcon] = useState('')
  const [iconSubmitting, setIconSubmitting] = useState(false)
  const [iconMsg, setIconMsg] = useState<string | null>(null)

  const handleRequestIcon = async () => {
    if (!selectedIcon) return
    setIconSubmitting(true)
    setIconMsg(null)
    try {
      await api.createAttendanceIconRequest(selectedIcon)
      setIconMsg('تم إرسال طلبك — بانتظار موافقة مدير النظام')
      setSelectedIcon('')
    } catch (err) {
      setIconMsg(err instanceof Error ? err.message : 'تعذر إرسال الطلب')
    } finally {
      setIconSubmitting(false)
    }
  }

  // موافقات مدير النظام على طلبات تعديل الرمز
  const [pendingRequests, setPendingRequests] = useState<Awaited<ReturnType<typeof api.getPendingAttendanceIconRequests>>>([])
  const isAdmin = employee?.role === 'ADMIN'
  useEffect(() => {
    if (isAdmin) api.getPendingAttendanceIconRequests().then(setPendingRequests).catch(() => {})
  }, [isAdmin])

  const handleResolve = async (id: string, approve: boolean) => {
    try {
      if (approve) await api.approveAttendanceIconRequest(id)
      else await api.rejectAttendanceIconRequest(id)
      setPendingRequests((prev) => prev.filter((r) => r.id !== id))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تنفيذ الإجراء')
    }
  }

  const tabs: { key: typeof tab; label: string }[] = [
    { key: 'password', label: 'كلمة المرور' },
    { key: 'volume', label: 'الصوت' },
    { key: 'icon', label: 'رمز الحضور' },
    ...(isAdmin ? [{ key: 'approvals' as const, label: `طلبات الرمز${pendingRequests.length ? ` (${pendingRequests.length})` : ''}` }] : []),
  ]

  return (
    <>
      <div className="fixed inset-0 z-40" onClick={onClose} />
      <div className="absolute left-0 top-12 z-50 w-80 max-w-[90vw] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <span className="text-sm font-bold text-slate-800">⚙️ الإعدادات</span>
          <button onClick={onClose} className="text-xl text-slate-400 hover:text-slate-600">×</button>
        </div>
        <div className="flex flex-wrap gap-1 border-b border-slate-100 p-2">
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`rounded-lg px-2.5 py-1.5 text-xs font-bold ${tab === t.key ? 'bg-brand-500 text-white' : 'text-slate-500 hover:bg-slate-100'}`}
            >
              {t.label}
            </button>
          ))}
        </div>

        <div className="max-h-96 overflow-y-auto p-4">
          {tab === 'password' && (
            <form onSubmit={handleChangePassword} className="flex flex-col gap-3">
              <input
                type="password" required placeholder="كلمة المرور الحالية"
                value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)}
                className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
              />
              <input
                type="password" required placeholder="كلمة المرور الجديدة" minLength={6}
                value={newPassword} onChange={(e) => setNewPassword(e.target.value)}
                className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
              />
              {pwMsg && <p className={`text-xs ${pwMsg.ok ? 'text-emerald-600' : 'text-red-600'}`}>{pwMsg.text}</p>}
              <button type="submit" disabled={pwSaving} className="rounded-lg bg-brand-500 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {pwSaving ? 'جاري الحفظ...' : 'تغيير كلمة المرور'}
              </button>
            </form>
          )}

          {tab === 'volume' && (
            <div>
              <label className="mb-2 block text-sm font-medium text-slate-600">مستوى الصوت: {volume}%</label>
              <input
                type="range" min={0} max={100} value={volume}
                onChange={(e) => handleVolumeChange(Number(e.target.value))}
                className="w-full"
              />
              <p className="mt-2 text-xs text-slate-400">راح يُستخدم هذا الإعداد لاحقاً مع الفيديوهات التدريبية.</p>
            </div>
          )}

          {tab === 'icon' && (
            <div>
              <p className="mb-3 text-sm text-slate-500">اختر رمزك المفضل — يحتاج موافقة مدير النظام قبل ما ينفعّل.</p>
              <div className="grid grid-cols-6 gap-2">
                {ICON_CHOICES.map((icon) => (
                  <button
                    key={icon}
                    onClick={() => setSelectedIcon(icon)}
                    className={`rounded-lg border-2 p-2 text-xl ${selectedIcon === icon ? 'border-brand-500 bg-brand-50' : 'border-transparent hover:bg-slate-50'}`}
                  >
                    {icon}
                  </button>
                ))}
              </div>
              <button
                onClick={handleRequestIcon}
                disabled={!selectedIcon || iconSubmitting}
                className="mt-3 w-full rounded-lg bg-brand-500 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50"
              >
                {iconSubmitting ? 'جاري الإرسال...' : 'إرسال طلب التعديل'}
              </button>
              {iconMsg && <p className="mt-2 text-xs text-slate-600">{iconMsg}</p>}
            </div>
          )}

          {tab === 'approvals' && isAdmin && (
            <div className="flex flex-col gap-2">
              {pendingRequests.length === 0 && <p className="py-6 text-center text-sm text-slate-400">ماكو طلبات معلّقة</p>}
              {pendingRequests.map((r) => (
                <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-200 p-2">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">{r.requestedIcon}</span>
                    <span className="text-sm font-bold text-slate-700">{r.employee?.name}</span>
                  </div>
                  <div className="flex gap-1">
                    <button onClick={() => handleResolve(r.id, true)} className="rounded-md bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">✔ موافقة</button>
                    <button onClick={() => handleResolve(r.id, false)} className="rounded-md bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">✘ رفض</button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  )
}
