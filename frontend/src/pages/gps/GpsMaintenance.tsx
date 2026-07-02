import { useState, useEffect } from 'react'
import { api } from '../../api'

const statusConfig: Record<string, { label: string; color: string; bg: string }> = {
  PENDING: { label: 'معلق', color: '#d97706', bg: '#fffbeb' },
  IN_PROGRESS: { label: 'قيد المعالجة', color: '#7c3aed', bg: '#f5f3ff' },
  COMPLETED: { label: 'مكتمل', color: '#16a34a', bg: '#f0fdf4' },
}

export default function GpsMaintenance() {
  const [requests, setRequests] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedRequest, setSelectedRequest] = useState<any>(null)
  const [adminNotes, setAdminNotes] = useState('')

  const load = () => {
    setLoading(true)
    api.getGpsMaintenance()
      .then(setRequests)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleUpdateStatus = async (id: string, status: 'PENDING' | 'IN_PROGRESS' | 'COMPLETED') => {
    try {
      await api.updateGpsMaintenance(id, { status, adminNotes })
      setSelectedRequest(null)
      setAdminNotes('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>
  if (error) return <div className="rounded-2xl bg-red-50 p-6 text-red-600 shadow-sm">{error}</div>

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">🔧 طلبات الصيانة</h1>
        <p className="mt-1 text-blue-200 text-sm">إدارة طلبات صيانة أجهزة التتبع</p>
      </div>

      {/* Request Cards */}
      <div className="space-y-4">
        {requests.map((r) => {
          const sc = statusConfig[r.status] || { label: r.status, color: '#64748b', bg: '#f8fafc' }
          return (
            <div key={r.id} className="rounded-2xl bg-white p-5 shadow-sm">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">🔧</span>
                  <div>
                    <h3 className="font-bold" style={{ color: '#1a3a5c' }}>{r.customer?.fullName || 'غير معروف'}</h3>
                    <p className="text-sm text-slate-500">📞 {r.customer?.phone || '-'}</p>
                  </div>
                </div>
                <span
                  className="rounded-full px-4 py-1 text-xs font-bold"
                  style={{ backgroundColor: sc.bg, color: sc.color }}
                >
                  {sc.label}
                </span>
              </div>

              <div className="mb-3 rounded-xl bg-slate-50 p-3">
                <span className="text-xs text-slate-500">وصف المشكلة</span>
                <p className="text-sm font-medium">{r.problemDescription || '-'}</p>
              </div>

              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 mb-3">
                <div>
                  <span className="text-xs text-slate-500">الموظف</span>
                  <p className="text-sm font-medium">{r.employee?.name || '-'}</p>
                </div>
                <div>
                  <span className="text-xs text-slate-500">التاريخ</span>
                  <p className="text-sm font-medium">{r.createdAt ? new Date(r.createdAt).toLocaleDateString('ar-IQ') : '-'}</p>
                </div>
                {r.adminNotes && (
                  <div>
                    <span className="text-xs text-slate-500">ملاحظات الإدارة</span>
                    <p className="text-sm font-medium">{r.adminNotes}</p>
                  </div>
                )}
              </div>

              <div className="border-t border-slate-100 pt-3">
                <button
                  onClick={() => { setSelectedRequest(r); setAdminNotes(r.adminNotes || '') }}
                  className="rounded-2xl px-5 py-2 text-sm font-bold text-white"
                  style={{ backgroundColor: '#1a3a5c' }}
                >
                  📋 تفاصيل
                </button>
              </div>
            </div>
          )
        })}
        {requests.length === 0 && (
          <div className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">لا يوجد طلبات صيانة</div>
        )}
      </div>

      {/* Detail Modal */}
      {selectedRequest && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-lg">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-xl font-bold" style={{ color: '#1a3a5c' }}>🔧 تفاصيل طلب الصيانة</h3>
              <button onClick={() => setSelectedRequest(null)} className="text-2xl text-slate-400 hover:text-slate-600">&times;</button>
            </div>

            {/* Customer Info Grid */}
            <div className="grid grid-cols-2 gap-4 mb-4">
              <div className="rounded-xl bg-slate-50 p-3">
                <span className="text-xs text-slate-500">الزبون</span>
                <p className="font-bold">{selectedRequest.customer?.fullName}</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <span className="text-xs text-slate-500">الهاتف</span>
                <p className="font-bold">{selectedRequest.customer?.phone || '-'}</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <span className="text-xs text-slate-500">الموظف</span>
                <p className="font-bold">{selectedRequest.employee?.name || '-'}</p>
              </div>
              <div className="rounded-xl bg-slate-50 p-3">
                <span className="text-xs text-slate-500">التاريخ</span>
                <p className="font-bold">{selectedRequest.createdAt ? new Date(selectedRequest.createdAt).toLocaleDateString('ar-IQ') : '-'}</p>
              </div>
            </div>

            {/* Problem Description */}
            <div className="mb-4 rounded-xl bg-slate-50 p-4">
              <span className="text-xs text-slate-500">وصف المشكلة</span>
              <p className="mt-1 font-medium">{selectedRequest.problemDescription}</p>
            </div>

            {/* Admin Notes */}
            <div className="mb-4">
              <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظات المسؤول</label>
              <textarea
                value={adminNotes}
                onChange={(e) => setAdminNotes(e.target.value)}
                rows={3}
                className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400"
                placeholder="أضف ملاحظاتك هنا..."
              />
            </div>

            {/* Action Buttons */}
            <div className="flex gap-3">
              {selectedRequest.status !== 'COMPLETED' && (
                <>
                  {selectedRequest.status === 'PENDING' && (
                    <button
                      onClick={() => handleUpdateStatus(selectedRequest.id, 'IN_PROGRESS')}
                      className="flex-1 rounded-2xl px-5 py-3 text-sm font-bold text-white"
                      style={{ backgroundColor: '#7c3aed' }}
                    >
                      ⚙️ قيد المعالجة
                    </button>
                  )}
                  {(selectedRequest.status === 'PENDING' || selectedRequest.status === 'IN_PROGRESS') && (
                    <button
                      onClick={() => handleUpdateStatus(selectedRequest.id, 'COMPLETED')}
                      className="flex-1 rounded-2xl px-5 py-3 text-sm font-bold text-white"
                      style={{ backgroundColor: '#16a34a' }}
                    >
                      ✅ مكتمل
                    </button>
                  )}
                </>
              )}
              <button
                onClick={() => setSelectedRequest(null)}
                className="rounded-2xl bg-slate-200 px-5 py-3 text-sm font-bold text-slate-600"
              >
                إغلاق
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
