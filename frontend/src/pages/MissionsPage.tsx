import { useState, useEffect, useCallback } from 'react'
import { useSession } from '../session'

const API = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'
async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, { headers: { 'Content-Type': 'application/json' }, ...opts })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

interface Mission {
  id: string; code: string; stage: string; bookingId: string
  leaderId: string; memberIds: string[]
  customerLat: number | null; customerLng: number | null; customerAddress: string | null
  assignedAt: string; materialsReadyAt: string | null; departedAt: string | null
  arrivedAt: string | null; workStartedAt: string | null; completedAt: string | null; stoppedAt: string | null
  estimatedMinutes: number | null; actualMinutes: number | null; distanceKm: number | null
  stopReason: string | null; notes: string | null
  leader: { id: string; name: string } | null
  members: { id: string; name: string }[]
  booking: {
    id: string; code: string
    customer: { name: string; phone: string; location: string | null }
    service: { name: string } | null
  }
  events: { id: string; action: string; employeeId: string; createdAt: string; note: string | null }[]
}

const STAGES: Record<string, { label: string; color: string; icon: string }> = {
  ASSIGNED:        { label: 'تم الإسناد', color: '#6b7280', icon: '📋' },
  MATERIALS_PREP:  { label: 'تجهيز المواد', color: '#d97706', icon: '📦' },
  MATERIALS_READY: { label: 'المواد جاهزة', color: '#2563eb', icon: '✅' },
  EN_ROUTE:        { label: 'بالطريق', color: '#7c3aed', icon: '🚗' },
  ARRIVED:         { label: 'وصل الموقع', color: '#0891b2', icon: '📍' },
  WORK_STARTED:    { label: 'جاري العمل', color: '#ea580c', icon: '⚡' },
  COMPLETED:       { label: 'مكتمل', color: '#16a34a', icon: '🏆' },
  STOPPED:         { label: 'متوقف', color: '#dc2626', icon: '⏸️' },
}

const stageOrder = ['ASSIGNED', 'MATERIALS_PREP', 'MATERIALS_READY', 'EN_ROUTE', 'ARRIVED', 'WORK_STARTED', 'COMPLETED']

function timeSince(dateStr: string) {
  const mins = Math.floor((Date.now() - new Date(dateStr).getTime()) / 60000)
  if (mins < 1) return 'الآن'
  if (mins < 60) return `${mins} دقيقة`
  const hrs = Math.floor(mins / 60)
  return `${hrs} ساعة ${mins % 60} دقيقة`
}

function timeStr(dateStr: string | null) {
  if (!dateStr) return '---'
  return new Date(dateStr).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
}

export default function MissionsPage() {
  const { employee, permissions } = useSession()
  const [missions, setMissions] = useState<Mission[]>([])
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'active' | 'completed' | 'monitor'>('active')
  const [selectedMission, setSelectedMission] = useState<Mission | null>(null)

  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'HR_COORDINATOR' || employee?.role === 'MONITOR' || permissions.includes('monitoring')

  const load = useCallback(async () => {
    try {
      let data: Mission[]
      if (isAdmin) {
        data = await request<Mission[]>('/missions')
      } else if (employee) {
        data = await request<Mission[]>(`/missions/my/${employee.id}`)
      } else {
        data = []
      }
      setMissions(data)
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [employee, isAdmin])

  useEffect(() => { load() }, [load])

  // Auto-refresh every 30s
  useEffect(() => {
    const iv = setInterval(load, 30000)
    return () => clearInterval(iv)
  }, [load])

  const advanceStage = async (mission: Mission, newStage: string, extra?: Record<string, unknown>) => {
    if (!employee) return
    let lat: number | undefined, lng: number | undefined
    if (['EN_ROUTE', 'ARRIVED'].includes(newStage)) {
      try {
        const pos = await new Promise<GeolocationPosition>((resolve, reject) =>
          navigator.geolocation.getCurrentPosition(resolve, reject, { enableHighAccuracy: true, timeout: 10000 })
        )
        lat = pos.coords.latitude
        lng = pos.coords.longitude
      } catch { /* GPS not available */ }
    }

    await request(`/missions/${mission.id}/stage`, {
      method: 'PUT',
      body: JSON.stringify({ stage: newStage, employeeId: employee.id, lat, lng, ...extra }),
    })
    load()
    setSelectedMission(null)
  }

  const activeMissions = missions.filter(m => !['COMPLETED', 'STOPPED'].includes(m.stage))
  const completedMissions = missions.filter(m => ['COMPLETED', 'STOPPED'].includes(m.stage))

  const getDelayStatus = (m: Mission) => {
    if (!m.estimatedMinutes) return 'normal'
    if (m.stage === 'EN_ROUTE' && m.departedAt) {
      const elapsed = Math.floor((Date.now() - new Date(m.departedAt).getTime()) / 60000)
      if (elapsed > m.estimatedMinutes) return 'late'
      if (elapsed > m.estimatedMinutes * 0.8) return 'warning'
    }
    if (m.actualMinutes && m.actualMinutes > m.estimatedMinutes) return 'late'
    return 'normal'
  }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" /></div>

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      {/* Header */}
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1a3a6c] p-6 text-white">
        <h1 className="text-2xl font-bold">🎯 نظام تتبع المهام الميدانية</h1>
        <p className="mt-1 text-blue-200/80">تتبع حي لجميع المهام والفرق الميدانية</p>
      </div>

      {/* Stats */}
      {isAdmin && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
          {Object.entries(STAGES).filter(([k]) => k !== 'STOPPED').map(([key, s]) => {
            const count = missions.filter(m => m.stage === key).length
            return (
              <div key={key} className="rounded-xl bg-white p-4 shadow-sm text-center">
                <span className="text-2xl">{s.icon}</span>
                <p className="mt-1 text-2xl font-bold" style={{ color: s.color }}>{count}</p>
                <p className="text-xs text-slate-500">{s.label}</p>
              </div>
            )
          })}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-2">
        <button onClick={() => setTab('active')} className={`rounded-xl px-5 py-2.5 text-sm font-bold transition ${tab === 'active' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>
          المهام النشطة ({activeMissions.length})
        </button>
        <button onClick={() => setTab('completed')} className={`rounded-xl px-5 py-2.5 text-sm font-bold transition ${tab === 'completed' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>
          المنجزة ({completedMissions.length})
        </button>
        {isAdmin && (
          <button onClick={() => setTab('monitor')} className={`rounded-xl px-5 py-2.5 text-sm font-bold transition ${tab === 'monitor' ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}>
            🖥️ لوحة المراقبة
          </button>
        )}
      </div>

      {/* Active Missions */}
      {tab === 'active' && (
        <div className="grid gap-4 lg:grid-cols-2">
          {activeMissions.length === 0 && <p className="col-span-2 text-center text-slate-400 py-8">لا توجد مهام نشطة</p>}
          {activeMissions.map(m => {
            const s = STAGES[m.stage] || STAGES.ASSIGNED
            const delay = getDelayStatus(m)
            const isMyMission = m.leaderId === employee?.id || m.memberIds.includes(employee?.id || '')
            const isLeader = m.leaderId === employee?.id
            const stageIdx = stageOrder.indexOf(m.stage)

            return (
              <div key={m.id} className={`rounded-2xl bg-white p-5 shadow-sm border-r-4 ${
                delay === 'late' ? 'border-red-500' : delay === 'warning' ? 'border-yellow-400' : 'border-transparent'
              }`}>
                {/* Top */}
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-lg">{s.icon}</span>
                    <span className="rounded-full px-3 py-1 text-xs font-bold text-white" style={{ backgroundColor: s.color }}>{s.label}</span>
                    <span className="text-xs text-slate-400">{m.code}</span>
                  </div>
                  {delay === 'late' && <span className="text-xs font-bold text-red-500 animate-pulse">❗ متأخر</span>}
                </div>

                {/* Customer info */}
                <div className="mb-3 space-y-1">
                  <p className="font-bold text-slate-800">{m.booking.customer.name}</p>
                  <p className="text-sm text-slate-500">{m.booking.service?.name || '---'} • {m.customerAddress || m.booking.customer.location || '---'}</p>
                  <p className="text-sm text-slate-400">📞 {m.booking.customer.phone}</p>
                </div>

                {/* Team */}
                <div className="mb-3 flex flex-wrap gap-1">
                  <span className="rounded-lg bg-blue-50 px-2 py-1 text-xs font-bold text-blue-700">👤 {m.leader?.name || '---'} (ليدر)</span>
                  {m.members.map(mem => (
                    <span key={mem.id} className="rounded-lg bg-slate-50 px-2 py-1 text-xs text-slate-600">{mem.name}</span>
                  ))}
                </div>

                {/* Timeline */}
                <div className="mb-3 flex items-center gap-1">
                  {stageOrder.map((st, i) => {
                    const done = i <= stageIdx
                    const stg = STAGES[st]
                    return (
                      <div key={st} className="flex items-center gap-1">
                        <div className={`h-2.5 w-2.5 rounded-full ${done ? '' : 'bg-slate-200'}`} style={done ? { backgroundColor: stg.color } : {}} title={stg.label} />
                        {i < stageOrder.length - 1 && <div className={`h-0.5 w-4 ${done ? 'bg-slate-400' : 'bg-slate-200'}`} />}
                      </div>
                    )
                  })}
                </div>

                {/* Times */}
                <div className="mb-3 grid grid-cols-3 gap-2 text-center text-[11px]">
                  <div><p className="text-slate-400">الإسناد</p><p className="font-bold">{timeStr(m.assignedAt)}</p></div>
                  <div><p className="text-slate-400">الانطلاق</p><p className="font-bold">{timeStr(m.departedAt)}</p></div>
                  <div><p className="text-slate-400">الوصول</p><p className="font-bold">{timeStr(m.arrivedAt)}</p></div>
                </div>

                {/* ETA info */}
                {m.estimatedMinutes && (
                  <div className="mb-3 rounded-lg bg-slate-50 p-2 text-center text-sm">
                    <span className="text-slate-500">الوقت المتوقع: </span>
                    <span className="font-bold text-blue-700">{m.estimatedMinutes} دقيقة</span>
                    {m.distanceKm && <span className="text-slate-400"> ({m.distanceKm.toFixed(1)} كم)</span>}
                    {m.actualMinutes && (
                      <>
                        <span className="text-slate-500"> • الفعلي: </span>
                        <span className={`font-bold ${m.actualMinutes > m.estimatedMinutes ? 'text-red-600' : 'text-green-600'}`}>
                          {m.actualMinutes} دقيقة
                        </span>
                      </>
                    )}
                  </div>
                )}

                {/* Timer */}
                {m.stage === 'MATERIALS_READY' && m.materialsReadyAt && (
                  <div className="mb-3 rounded-lg bg-amber-50 border border-amber-200 p-3 text-center">
                    <p className="text-xs text-amber-600">⏱ وقت الانتظار منذ جهوزية المواد</p>
                    <p className="text-lg font-bold text-amber-700">{timeSince(m.materialsReadyAt)}</p>
                  </div>
                )}
                {m.stage === 'EN_ROUTE' && m.departedAt && (
                  <div className="mb-3 rounded-lg bg-purple-50 border border-purple-200 p-3 text-center">
                    <p className="text-xs text-purple-600">🚗 بالطريق منذ</p>
                    <p className="text-lg font-bold text-purple-700">{timeSince(m.departedAt)}</p>
                  </div>
                )}
                {m.stage === 'WORK_STARTED' && m.workStartedAt && (
                  <div className="mb-3 rounded-lg bg-orange-50 border border-orange-200 p-3 text-center">
                    <p className="text-xs text-orange-600">⚡ يعمل منذ</p>
                    <p className="text-lg font-bold text-orange-700">{timeSince(m.workStartedAt)}</p>
                  </div>
                )}

                {/* Actions */}
                {(isMyMission || isAdmin) && (
                  <div className="flex flex-wrap gap-2">
                    {m.stage === 'ASSIGNED' && isLeader && (
                      <button onClick={() => advanceStage(m, 'MATERIALS_PREP')} className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-bold text-white hover:bg-amber-600">📦 بدأ تجهيز المواد</button>
                    )}
                    {m.stage === 'MATERIALS_PREP' && isLeader && (
                      <button onClick={() => advanceStage(m, 'MATERIALS_READY')} className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-bold text-white hover:bg-blue-700">✅ المواد جاهزة</button>
                    )}
                    {m.stage === 'MATERIALS_READY' && (
                      <button onClick={() => {
                        const est = prompt('كم دقيقة الوقت المتوقع للوصول؟')
                        const dist = prompt('كم كيلومتر المسافة؟ (اختياري)')
                        advanceStage(m, 'EN_ROUTE', { estimatedMinutes: est, distanceKm: dist })
                      }} className="rounded-lg bg-purple-600 px-4 py-2 text-sm font-bold text-white hover:bg-purple-700">🚗 انطلقنا</button>
                    )}
                    {m.stage === 'EN_ROUTE' && isLeader && (
                      <button onClick={() => advanceStage(m, 'ARRIVED')} className="rounded-lg bg-cyan-600 px-4 py-2 text-sm font-bold text-white hover:bg-cyan-700">📍 وصلنا</button>
                    )}
                    {m.stage === 'ARRIVED' && (
                      <button onClick={() => advanceStage(m, 'WORK_STARTED')} className="rounded-lg bg-orange-500 px-4 py-2 text-sm font-bold text-white hover:bg-orange-600">⚡ بدأ العمل</button>
                    )}
                    {m.stage === 'WORK_STARTED' && (
                      <>
                        <button onClick={() => advanceStage(m, 'COMPLETED')} className="rounded-lg bg-green-600 px-4 py-2 text-sm font-bold text-white hover:bg-green-700">🏆 تم الإنجاز</button>
                        <button onClick={() => {
                          const reason = prompt('سبب التوقف؟')
                          if (reason) advanceStage(m, 'STOPPED', { stopReason: reason })
                        }} className="rounded-lg bg-red-500 px-4 py-2 text-sm font-bold text-white hover:bg-red-600">⏸️ توقف</button>
                      </>
                    )}
                    <button onClick={() => setSelectedMission(m)} className="rounded-lg bg-slate-100 px-3 py-2 text-sm text-slate-600 hover:bg-slate-200">📊 التفاصيل</button>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* Completed Missions */}
      {tab === 'completed' && (
        <div className="space-y-3">
          {completedMissions.length === 0 && <p className="text-center text-slate-400 py-8">لا توجد مهام منجزة</p>}
          {completedMissions.map(m => {
            const s = STAGES[m.stage]
            return (
              <div key={m.id} className="flex items-center justify-between rounded-xl bg-white p-4 shadow-sm">
                <div className="flex items-center gap-3">
                  <span className="text-lg">{s.icon}</span>
                  <div>
                    <p className="font-bold text-slate-700">{m.booking.customer.name} — {m.code}</p>
                    <p className="text-xs text-slate-400">{m.booking.service?.name} • {m.leader?.name}</p>
                  </div>
                </div>
                <div className="text-left">
                  {m.actualMinutes && m.estimatedMinutes && (
                    <p className={`text-sm font-bold ${m.actualMinutes > m.estimatedMinutes ? 'text-red-500' : 'text-green-500'}`}>
                      {m.actualMinutes > m.estimatedMinutes ? `+${m.actualMinutes - m.estimatedMinutes} دقيقة تأخير` : 'بالوقت ✓'}
                    </p>
                  )}
                  {m.stage === 'STOPPED' && <p className="text-xs text-red-400">{m.stopReason}</p>}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Monitor Dashboard */}
      {tab === 'monitor' && isAdmin && (
        <div className="space-y-4">
          <div className="grid gap-4 lg:grid-cols-2">
            {activeMissions.map(m => {
              const s = STAGES[m.stage]
              const delay = getDelayStatus(m)
              return (
                <div key={m.id} className={`rounded-xl p-4 shadow-sm ${
                  delay === 'late' ? 'bg-red-50 border border-red-200' : delay === 'warning' ? 'bg-yellow-50 border border-yellow-200' : 'bg-white'
                }`}>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="rounded-full px-3 py-1 text-xs font-bold text-white" style={{ backgroundColor: s.color }}>{s.icon} {s.label}</span>
                      <span className="text-xs text-slate-400">{m.code}</span>
                    </div>
                    <span className={`text-xs font-bold ${delay === 'late' ? 'text-red-500' : delay === 'warning' ? 'text-yellow-600' : 'text-green-500'}`}>
                      {delay === 'late' ? '🔴 متأخر' : delay === 'warning' ? '🟡 قارب' : '🟢 بالوقت'}
                    </span>
                  </div>
                  <p className="mt-2 font-bold text-slate-700">{m.leader?.name} ← {m.booking.customer.name}</p>
                  <p className="text-xs text-slate-400">{m.customerAddress || m.booking.customer.location || '---'}</p>
                  {m.estimatedMinutes && (
                    <p className="mt-1 text-xs">⏱ المتوقع: {m.estimatedMinutes} دقيقة {m.distanceKm ? `(${m.distanceKm.toFixed(1)} كم)` : ''}</p>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {selectedMission && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setSelectedMission(null)}>
          <div className="mx-4 max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl" onClick={e => e.stopPropagation()}>
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-bold">📊 تفاصيل المهمة {selectedMission.code}</h3>
              <button onClick={() => setSelectedMission(null)} className="text-slate-400 hover:text-slate-600">✕</button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="rounded-lg bg-slate-50 p-3">
                <p className="font-bold">{selectedMission.booking.customer.name}</p>
                <p className="text-slate-500">{selectedMission.booking.service?.name} • {selectedMission.customerAddress}</p>
              </div>
              <h4 className="font-bold text-slate-700">سجل الأحداث</h4>
              <div className="space-y-2">
                {selectedMission.events.map(ev => (
                  <div key={ev.id} className="flex items-center gap-3 rounded-lg bg-slate-50 p-2">
                    <span className="text-lg">{STAGES[ev.action]?.icon || '•'}</span>
                    <div>
                      <p className="font-bold text-slate-700">{STAGES[ev.action]?.label || ev.action}</p>
                      <p className="text-xs text-slate-400">{new Date(ev.createdAt).toLocaleString('ar-IQ')}</p>
                      {ev.note && <p className="text-xs text-slate-500">{ev.note}</p>}
                    </div>
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-2 gap-2 text-center">
                <div className="rounded-lg bg-blue-50 p-2">
                  <p className="text-xs text-blue-500">الوقت المتوقع</p>
                  <p className="font-bold text-blue-700">{selectedMission.estimatedMinutes || '---'} دقيقة</p>
                </div>
                <div className="rounded-lg bg-green-50 p-2">
                  <p className="text-xs text-green-500">الوقت الفعلي</p>
                  <p className="font-bold text-green-700">{selectedMission.actualMinutes || '---'} دقيقة</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
