import { useEffect, useState } from 'react'
import { api, type PrivacyPolicyPoint } from '../api'
import { useSession } from '../session'

/**
 * صفحة سياسة الخصوصية — يقراها أي موظف من الإعدادات بأي وقت.
 * صاحب صلاحية privacy_policy_manage يقدر يضيف/يعدّل/يحذف النقاط.
 * المالك ومدير النظام يشوفون منو أضاف كل نقطة (السيرفر يحجبها عن غيرهم).
 */
export default function PrivacyPolicyPage() {
  const { employee, permissions } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  const canManage = isAdmin || permissions.includes('privacy_policy_manage')

  const [points, setPoints] = useState<PrivacyPolicyPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [newPoint, setNewPoint] = useState('')
  const [saving, setSaving] = useState(false)
  const [editId, setEditId] = useState<string | null>(null)
  const [editText, setEditText] = useState('')
  const [acceptedAt, setAcceptedAt] = useState<string | null>(null)

  const load = () => {
    api.getPrivacyPolicy(canManage)
      .then(setPoints)
      .catch(() => setPoints([]))
      .finally(() => setLoading(false))
    api.getPrivacyPolicyStatus().then((s) => setAcceptedAt(s.acceptedAt)).catch(() => {})
  }

  useEffect(load, [canManage])

  const addPoint = async () => {
    if (!newPoint.trim()) return
    setSaving(true)
    try {
      await api.createPrivacyPolicyPoint(newPoint.trim())
      setNewPoint('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر إضافة النقطة')
    } finally {
      setSaving(false)
    }
  }

  const saveEdit = async (id: string) => {
    if (!editText.trim()) return
    try {
      await api.updatePrivacyPolicyPoint(id, { content: editText.trim() })
      setEditId(null)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التعديل')
    }
  }

  const toggleActive = async (p: PrivacyPolicyPoint) => {
    try {
      await api.updatePrivacyPolicyPoint(p.id, { isActive: !p.isActive })
      load()
    } catch { /* ignore */ }
  }

  const removePoint = async (p: PrivacyPolicyPoint) => {
    if (!confirm('حذف هذي النقطة نهائياً؟')) return
    try {
      await api.deletePrivacyPolicyPoint(p.id)
      load()
    } catch { /* ignore */ }
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">🔒 سياسة الخصوصية</h2>
      <p className="mt-1 text-slate-500">
        سياسة الخصوصية الخاصة بشركة الأماني — نقاط يقراها كل موظف ويوافق عليها.
      </p>
      {acceptedAt && (
        <p className="mt-2 inline-block rounded-lg bg-emerald-50 px-3 py-1.5 text-sm font-bold text-emerald-700">
          ✓ وافقت عليها بتاريخ {new Date(acceptedAt).toLocaleDateString('ar-IQ')}
        </p>
      )}

      {canManage && (
        <div className="mt-6 rounded-xl border border-brand-100 bg-brand-50/50 p-5">
          <label className="mb-2 block text-sm font-bold text-brand-800">إضافة نقطة جديدة</label>
          <textarea
            value={newPoint}
            onChange={(e) => setNewPoint(e.target.value)}
            rows={2}
            placeholder="اكتب نص النقطة..."
            className="w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
          />
          <button
            onClick={addPoint}
            disabled={saving || !newPoint.trim()}
            className="mt-2 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2.5 font-bold text-white disabled:opacity-50"
          >
            {saving ? 'جاري الإضافة...' : '+ إضافة النقطة'}
          </button>
          <p className="mt-2 text-xs text-slate-500">
            أي نقطة جديدة تنطلب موافقة جديدة من كل الموظفين تلقائياً أول ما يفتحون النظام.
          </p>
        </div>
      )}

      {loading ? (
        <p className="mt-6 text-slate-400">جاري التحميل...</p>
      ) : points.length === 0 ? (
        <div className="mt-6 rounded-xl bg-white p-10 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <p className="text-slate-400">ما اكو نقاط مضافة لسياسة الخصوصية بعد</p>
        </div>
      ) : (
        <ol className="mt-6 space-y-3">
          {points.map((p, i) => (
            <li
              key={p.id}
              className={`rounded-xl border bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] ${
                p.isActive ? 'border-white' : 'border-slate-200 opacity-60'
              }`}
            >
              <div className="flex items-start gap-3">
                <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-100 text-sm font-bold text-brand-700">
                  {i + 1}
                </span>
                <div className="flex-1">
                  {editId === p.id ? (
                    <>
                      <textarea
                        value={editText}
                        onChange={(e) => setEditText(e.target.value)}
                        rows={2}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                      />
                      <div className="mt-2 flex gap-2">
                        <button onClick={() => saveEdit(p.id)} className="rounded-lg bg-brand-600 px-4 py-1.5 text-xs font-bold text-white">حفظ</button>
                        <button onClick={() => setEditId(null)} className="rounded-lg border border-slate-300 px-4 py-1.5 text-xs font-bold text-slate-600">إلغاء</button>
                      </div>
                    </>
                  ) : (
                    <p className="text-slate-800">{p.content}</p>
                  )}

                  <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-400">
                    {/* منو أضاف النقطة — السيرفر ما يرجعه إلا للمالك ومدير النظام */}
                    {p.createdByName && <span>أضافها: <span className="font-bold text-slate-600">{p.createdByName}</span></span>}
                    <span>{new Date(p.createdAt).toLocaleDateString('ar-IQ')}</span>
                    {!p.isActive && <span className="font-bold text-amber-600">معطّلة</span>}
                  </div>
                </div>

                {canManage && editId !== p.id && (
                  <div className="flex shrink-0 flex-wrap gap-1">
                    <button onClick={() => { setEditId(p.id); setEditText(p.content) }}
                      className="rounded-lg bg-brand-50 px-2.5 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">✎</button>
                    <button onClick={() => toggleActive(p)}
                      className="rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">
                      {p.isActive ? 'تعطيل' : 'تفعيل'}
                    </button>
                    <button onClick={() => removePoint(p)}
                      className="rounded-lg bg-red-50 px-2.5 py-1 text-xs font-bold text-red-600 hover:bg-red-100">🗑</button>
                  </div>
                )}
              </div>
            </li>
          ))}
        </ol>
      )}
    </div>
  )
}
