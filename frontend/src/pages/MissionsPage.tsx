import { useState, useEffect, useCallback, useRef } from 'react'
import { useSession } from '../session'

const API = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'
async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const token = localStorage.getItem('authToken')
  const res = await fetch(`${API}${path}`, {
    ...opts,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(opts?.headers || {}),
    },
  })
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

// ═══ ألوان المراحل ═══
// ⚠️ لونان لكل مرحلة مو واحد: `ink` للنص (ينقلب بالوضع الليلي حتى
// ينقرا على البطاقة الغامقة)، و`fill` لخلفية الشارة الي عليها نص
// أبيض (يبقى غامقاً بالوضعين — قلبه يخلّي أبيض على فاتح).
// نفس فخّ `--color-white` الي انكوينا بيه مرتين قبل.
//
// `word` كلمة حالة البطاقة، و`empty` الي تنعرض لمن يكون العدد صفر
// («لا توجد الآن» أوضح من «نشطة: 0»). و`tint` خلفية دائرة الأيقونة.
// كلهن بيانات بالخريطة مو شروط بالكود — إضافة مرحلة جديدة ما
// تحتاج تلمس العرض.
const STAGES: Record<string, {
  label: string; ink: string; fill: string; icon: string
  word: string; empty: string; tint: string
}> = {
  ASSIGNED:        { label: 'تم الإسناد',   ink: 'var(--t-muted)',   fill: '#475569', icon: '📋',
                     word: 'موزعة',          empty: 'ماكو إسناد',     tint: 'var(--sf-sunken)' },
  MATERIALS_PREP:  { label: 'تجهيز المواد', ink: 'var(--t-warning)', fill: '#b45309', icon: '📦',
                     word: 'بانتظار التجهيز', empty: 'ماكو تجهيز',    tint: 'var(--tint-warning)' },
  MATERIALS_READY: { label: 'المواد جاهزة', ink: 'var(--t-info)',    fill: '#1d4ed8', icon: '✅',
                     word: 'جاهزة',          empty: 'ماكو جاهز',      tint: 'var(--sf-info)' },
  EN_ROUTE:        { label: 'بالطريق',      ink: 'var(--t-violet)',  fill: '#6d28d9', icon: '🚗',
                     word: 'متحرّكة',        empty: 'لا توجد الآن',   tint: 'var(--sf-violet)' },
  ARRIVED:         { label: 'وصل الموقع',   ink: 'var(--t-cyan)',    fill: '#0e7490', icon: '📍',
                     word: 'فعال',           empty: 'ماكو وصول',      tint: 'var(--sf-info)' },
  WORK_STARTED:    { label: 'جاري العمل',   ink: 'var(--t-warning)', fill: '#b45309', icon: '⚡',
                     word: 'نشطة',           empty: 'ماكو شغل جارٍ',  tint: 'var(--tint-warning)' },
  COMPLETED:       { label: 'مكتمل',        ink: 'var(--t-success)', fill: '#15803d', icon: '🏆',
                     word: 'منجز',           empty: 'ماكو إنجاز',     tint: 'var(--sf-success)' },
  STOPPED:         { label: 'متوقف',        ink: 'var(--t-danger)',  fill: '#b91c1c', icon: '⏸️',
                     word: 'متوقفة',         empty: 'ماكو متوقف',     tint: 'var(--sf-danger)' },
}

// ترتيب بطاقات الحالات: من الشغل الجاري للمنجز — الأهم أول.
// ⚠️ غير `stageOrder` تحت: هذاك مسار المهمة بالخط الزمني.
const TILE_ORDER = ['WORK_STARTED', 'ARRIVED', 'EN_ROUTE', 'MATERIALS_READY',
  'MATERIALS_PREP', 'ASSIGNED', 'COMPLETED']

const stageOrder = ['ASSIGNED', 'MATERIALS_PREP', 'MATERIALS_READY', 'EN_ROUTE', 'ARRIVED', 'WORK_STARTED', 'COMPLETED']

// حلقة النسبة — SVG خالص. دائرتان: مسار باهت وقوس فوقه بطول
// `stroke-dasharray` يساوي النسبة من المحيط.
// شريحة صغيرة بالترويسة — شفافة على المتدرّج الغامق، فالنص أبيض
// دائماً وما تحتاج رموز الوضع الليلي.
function HeroChip({ icon, label, value }: { icon: string; label: string; value: string }) {
  return (
    <span className="flex items-center gap-2 rounded-xl bg-white/10 px-3 py-2">
      <span className="text-base">{icon}</span>
      <span className="leading-tight">
        <span className="block text-[10px] text-blue-200/80">{label}</span>
        <span className="block text-sm font-bold text-white">{value}</span>
      </span>
    </span>
  )
}

function ProgressRing({ pct }: { pct: number }) {
  const R = 34
  const C = 2 * Math.PI * R
  const safe = Math.max(0, Math.min(100, pct))
  return (
    <div className="relative shrink-0" style={{ width: 84, height: 84 }}>
      <svg width="84" height="84" viewBox="0 0 84 84" style={{ transform: 'rotate(-90deg)' }}>
        <circle cx="42" cy="42" r={R} fill="none" stroke="rgba(255,255,255,0.18)" strokeWidth="8" />
        <circle
          cx="42" cy="42" r={R} fill="none" stroke="#4ade80" strokeWidth="8"
          strokeLinecap="round"
          strokeDasharray={`${(safe / 100) * C} ${C}`}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-xl font-extrabold text-white">{Math.round(safe)}%</span>
        <span className="text-[9px] text-blue-200/80">نسبة الإنجاز</span>
      </div>
    </div>
  )
}

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
  const [pathMission, setPathMission] = useState<Mission | null>(null)
  // Ticking clock so delay-status calculations stay pure during render (no Date.now() in render body)
  // while still updating live; 30s cadence is enough for a minutes-based delay indicator.
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const iv = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(iv)
  }, [])

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

  // `load` is async and only calls setState after its awaits resolve (never
  // synchronously in the effect body), so this isn't the cascading-render pattern the
  // rule targets; it's a false positive for async fetch-on-mount effects.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load() }, [load])

  // Auto-refresh every 30s
  useEffect(() => {
    const iv = setInterval(load, 30000)
    return () => clearInterval(iv)
  }, [load])

  // تتبع موقع حي: أي مهمة أنا قائدها أو عضو فيها وحالتها "بالطريق" — نرسل موقعنا كل 20 ثانية
  useEffect(() => {
    if (!employee || !('geolocation' in navigator)) return
    const myEnRoute = missions.find(m => m.stage === 'EN_ROUTE' && (m.leaderId === employee.id || m.memberIds.includes(employee.id)))
    if (!myEnRoute) return

    const sendPing = () => {
      navigator.geolocation.getCurrentPosition(
        pos => {
          request('/location-pings', {
            method: 'POST',
            body: JSON.stringify({ latitude: pos.coords.latitude, longitude: pos.coords.longitude, bookingId: myEnRoute.bookingId }),
          }).catch(() => {})
        },
        () => {},
        { enableHighAccuracy: true, timeout: 10000 }
      )
    }
    sendPing()
    const iv = setInterval(sendPing, 20000)
    return () => clearInterval(iv)
  }, [missions, employee])

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

  // ⚠️ «نسبة الإنجاز» **مو تعريفاً جديداً**: نفس الي تستعمله شاشة
  // الإحصائيات (StatsPage.tsx: المنجز ÷ المسند). كلمة وحدة بمعنى
  // واحد بكل الشاشات — وإلا يصير رقمان بنفس الاسم، وهاي بالضبط
  // العلّة الي دا نصلّحها بشاشات ثانية.
  const activeMissions = missions.filter(m => !['COMPLETED', 'STOPPED'].includes(m.stage))
  const completedMissions = missions.filter(m => ['COMPLETED', 'STOPPED'].includes(m.stage))

  const donePct = missions.length
    ? (missions.filter(m => m.stage === 'COMPLETED').length / missions.length) * 100
    : 0

  const getDelayStatus = (m: Mission) => {
    if (!m.estimatedMinutes) return 'normal'
    if (m.stage === 'EN_ROUTE' && m.departedAt) {
      const elapsed = Math.floor((now - new Date(m.departedAt).getTime()) / 60000)
      if (elapsed > m.estimatedMinutes) return 'late'
      if (elapsed > m.estimatedMinutes * 0.8) return 'warning'
    }
    if (m.actualMinutes && m.actualMinutes > m.estimatedMinutes) return 'late'
    return 'normal'
  }

  // ⚠️ «حالة اليوم» حكم على الشغل، فلازم عتباته تكون مكتوبة صريحة
  // مو مخبّاة: تنبني على المهام **المتأخرة فعلاً** (getDelayStatus
  // الي تقارن الوقت المستغرق بـestimatedMinutes)، مو على مزاج.
  const lateCount = activeMissions.filter(m => getDelayStatus(m) === 'late').length
  const dayHealth = lateCount === 0
    ? { word: 'ممتاز', tone: 'var(--t-success)' }
    : lateCount <= 2
      ? { word: 'جيد', tone: 'var(--t-warning)' }
      : { word: 'يحتاج متابعة', tone: 'var(--t-danger)' }

  if (loading) return <div className="flex items-center justify-center p-12"><div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" /></div>

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      {/* Header */}
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1a3a6c] p-6 text-white">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">🎯 نظام تتبع المهام الميدانية</h1>
            <p className="mt-1 text-blue-200/80">
              تتبع جميع المهام وفرق العمل في الميدان لحظة بلحظة
            </p>
            <div className="mt-4 flex flex-wrap gap-2">
              <HeroChip icon="🗓️" label="حالة اليوم" value={dayHealth.word} />
              <HeroChip icon="👥" label="المهام النشطة" value={String(activeMissions.length)} />
              <HeroChip icon="📈" label="نسبة الإنجاز" value={`${Math.round(donePct)}%`} />
            </div>
          </div>
          <ProgressRing pct={donePct} />
        </div>
      </div>

      {/* Stats */}
      {isAdmin && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-7">
          {TILE_ORDER.map((key) => {
            const st = STAGES[key]
            const count = missions.filter(m => m.stage === key).length
            return (
              <div key={key}
                className="rounded-xl border p-4 text-center shadow-sm"
                style={{
                  backgroundColor: 'var(--sf-card)',
                  borderColor: key === 'COMPLETED' ? st.ink : 'var(--bd-line)',
                }}>
                <span
                  className="mx-auto flex h-10 w-10 items-center justify-center rounded-full text-xl"
                  style={{ backgroundColor: st.tint }}>
                  {st.icon}
                </span>
                <p className="mt-2 text-2xl font-extrabold" style={{ color: st.ink }}>{count}</p>
                <p className="text-xs" style={{ color: 'var(--t-muted)' }}>{st.label}</p>
                {/* ⚠️ «نشطة: 0» تضليل — لمن يكون صفراً نكول ماكو شي. */}
                <p className="mt-0.5 text-[10px] font-bold"
                  style={{ color: count > 0 ? st.ink : 'var(--t-faint)' }}>
                  {count > 0 ? st.word : st.empty}
                </p>
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
                    <span className="rounded-full px-3 py-1 text-xs font-bold text-white" style={{ backgroundColor: s.fill }}>{s.label}</span>
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
                        <div className={`h-2.5 w-2.5 rounded-full ${done ? '' : 'bg-slate-200'}`} style={done ? { backgroundColor: stg.fill } : {}} title={stg.label} />
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
                    {isAdmin && (
                      <button onClick={() => setPathMission(m)} className="mt-2 rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-purple-700">
                        🗺️ شوف المسار الحي
                      </button>
                    )}
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
                        <button onClick={() => advanceStage(m, 'COMPLETED')} className="rounded-lg bg-green-700 px-4 py-2 text-sm font-bold text-white hover:bg-green-800">🏆 تم الإنجاز</button>
                        <button onClick={() => {
                          const reason = prompt('سبب التوقف؟')
                          if (reason) advanceStage(m, 'STOPPED', { stopReason: reason })
                        }} className="rounded-lg bg-red-600 px-4 py-2 text-sm font-bold text-white hover:bg-red-700">⏸️ توقف</button>
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
                      <span className="rounded-full px-3 py-1 text-xs font-bold text-white" style={{ backgroundColor: s.fill }}>{s.icon} {s.label}</span>
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

      {pathMission && <LivePathModal mission={pathMission} onClose={() => setPathMission(null)} />}
    </div>
  )
}

// ═══ خريطة المسار الحي — تجيب نقاط الموقع المرسلة من متصفح الفني وترسمها كخط متحرك ═══
function LivePathModal({ mission, onClose }: { mission: Mission; onClose: () => void }) {
  const [points, setPoints] = useState<{ latitude: number; longitude: number; createdAt: string }[]>([])
  const mapRef = useRef<HTMLDivElement | null>(null)
  const leafletRef = useRef<{ map: import('leaflet').Map | null; line: import('leaflet').Polyline | null; marker: import('leaflet').Marker | null }>({ map: null, line: null, marker: null })

  const load = useCallback(async () => {
    try {
      const memberId = mission.leaderId
      const data = await request<{ latitude: number; longitude: number; createdAt: string }[]>(
        `/location-pings/path?employeeId=${memberId}&bookingId=${mission.bookingId}`
      )
      setPoints(data)
    } catch { /* ignore */ }
  }, [mission])

  // Same as above: `load` is async and only sets state after awaiting the request.
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load(); const iv = setInterval(load, 15000); return () => clearInterval(iv) }, [load])

  useEffect(() => {
    let disposed = false
    const leaflet = leafletRef.current
    import('leaflet').then(L => {
      if (disposed || !mapRef.current || leaflet.map) return
      const center = points[points.length - 1] || { latitude: 33.3152, longitude: 44.3661 }
      const map = L.map(mapRef.current).setView([center.latitude, center.longitude], 13)
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '© OpenStreetMap' }).addTo(map)
      leaflet.map = map
    })
    return () => { disposed = true; if (leaflet.map) { leaflet.map.remove(); leaflet.map = null } }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!leafletRef.current.map || points.length === 0) return
    import('leaflet').then(L => {
      const ref = leafletRef.current
      const map = ref.map
      if (!map) return
      const latlngs = points.map(p => [p.latitude, p.longitude]) as [number, number][]
      if (ref.line) map.removeLayer(ref.line)
      if (ref.marker) map.removeLayer(ref.marker)
      ref.line = L.polyline(latlngs, { color: 'var(--t-violet)', weight: 4 }).addTo(map)
      ref.marker = L.marker(latlngs[latlngs.length - 1]).addTo(map)
      map.fitBounds(ref.line.getBounds(), { padding: [30, 30] })
    })
  }, [points])

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div className="w-full max-w-2xl overflow-hidden rounded-2xl bg-white shadow-xl" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-slate-100 p-4">
          <h3 className="text-lg font-bold text-purple-700">🗺️ المسار الحي — {mission.leader?.name || 'الفريق'}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600">✕</button>
        </div>
        <div ref={el => { mapRef.current = el }} style={{ height: 420 }} />
        <div className="p-3 text-center text-xs text-slate-400">
          {points.length > 0 ? `${points.length} نقطة موقع مسجلة — يتحدث تلقائياً كل 15 ثانية` : 'ما وصلت أي نقطة موقع بعد — تأكد إن الفني فاتح النظام بموبايله'}
        </div>
      </div>
    </div>
  )
}
