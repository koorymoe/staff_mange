import { api } from '../api'
import { isNetworkError } from './offlineStages'

// ═══ تتبع المسار — قرار (ع) 10-14 ═══
//
// من «انطلاق» لحد «تم الإنجاز» الموبايل يسجّل نقطة كل ٣٠ ثانية أو كل ٥٠ متر،
// يحفظها محلياً (حتى بلا نت) ويرسلها دفعات. السيرفر منها يحسب الانطلاق والوصول
// (توقف ٣٠ دقيقة قرب الزبون) والإنجاز والتوقفات بالطريق — للمقارنة ويّا الضغطات.
//
// ⚠️ حد المتصفح: إذا الصفحة تسكّرت أو الموبايل انقفل، المتصفح يوقف الـGPS.
// فنطلب Wake Lock حتى الشاشة ما تطفى، والضغطات اليدوية تبقى هي الأساس.

interface TrackPoint { lat: number; lng: number; accuracy?: number; at: string; bookingId?: string }

const QUEUE = 'amani_track_queue_v1'
const ACTIVE = 'amani_track_active_v1'
const MIN_GAP_MS = 30_000
const MIN_MOVE_M = 50

const read = (): TrackPoint[] => { try { return JSON.parse(localStorage.getItem(QUEUE) || '[]') as TrackPoint[] } catch { return [] } }
const write = (q: TrackPoint[]) => { try { localStorage.setItem(QUEUE, JSON.stringify(q.slice(-5000))) } catch { /* ممتلئ */ } }

function distM(a: TrackPoint, b: { lat: number; lng: number }) {
  const r = 6371000, rad = (d: number) => d * Math.PI / 180
  const dLat = rad(b.lat - a.lat), dLng = rad(b.lng - a.lng)
  const x = Math.sin(dLat / 2) ** 2 + Math.cos(rad(a.lat)) * Math.cos(rad(b.lat)) * Math.sin(dLng / 2) ** 2
  return 2 * r * Math.asin(Math.sqrt(x))
}

let watchId: number | null = null
let activeBooking: string | null = null
let last: TrackPoint | null = null
let wakeLock: { release: () => Promise<void> } | null = null

async function keepAwake() {
  try {
    const wl = (navigator as Navigator & { wakeLock?: { request: (t: 'screen') => Promise<{ release: () => Promise<void> }> } }).wakeLock
    if (wl && !wakeLock) wakeLock = await wl.request('screen')
  } catch { /* مو مدعوم أو الصفحة مخفية */ }
}

function onPos(p: GeolocationPosition) {
  if (p.coords.accuracy > 1000) return
  const pt: TrackPoint = { lat: p.coords.latitude, lng: p.coords.longitude, accuracy: Math.round(p.coords.accuracy), at: new Date(p.timestamp).toISOString(), bookingId: activeBooking ?? undefined }
  if (last && Date.now() - new Date(last.at).getTime() < MIN_GAP_MS && distM(last, pt) < MIN_MOVE_M) return
  last = pt
  write([...read(), pt])
  window.dispatchEvent(new Event('track-queue'))
}

let stopTimer: ReturnType<typeof setTimeout> | null = null

/**
 * setTrackingBooking: الحجز الي الموظف بطريقه/عنده هسه (null = وقف التتبع).
 * بعد الإنجاز نكمل ١٥ دقيقة حتى النظام يشوف الطلعة من عند الزبون (الإنجاز التلقائي).
 */
export function setTrackingBooking(bookingId: string | null) {
  if (bookingId === null && activeBooking !== null && watchId !== null) {
    if (!stopTimer) stopTimer = setTimeout(() => { stopTimer = null; stopNow() }, 15 * 60_000)
    return
  }
  if (stopTimer) { clearTimeout(stopTimer); stopTimer = null }
  if (bookingId === null) { stopNow(); return }
  startOn(bookingId)
}

function stopNow() {
  activeBooking = null
  try { localStorage.removeItem(ACTIVE) } catch { /* */ }
  if (watchId !== null) navigator.geolocation.clearWatch(watchId)
  watchId = null
  void wakeLock?.release().catch(() => {}); wakeLock = null
  void flushTrack()
}

function startOn(bookingId: string) {
  if (bookingId === activeBooking && watchId !== null) return
  activeBooking = bookingId
  try { localStorage.setItem(ACTIVE, bookingId) } catch { /* */ }
  if (!('geolocation' in navigator)) return
  if (watchId === null) {
    watchId = navigator.geolocation.watchPosition(onPos, () => {}, { enableHighAccuracy: true, maximumAge: 10_000, timeout: 60_000 })
  }
  void keepAwake()
}

export const isTracking = () => watchId !== null
export const pendingTrackPoints = () => read().length

let flushing = false
export async function flushTrack(): Promise<void> {
  if (flushing || !navigator.onLine) return
  flushing = true
  try {
    for (;;) {
      const q = read()
      if (q.length === 0) break
      const chunk = q.slice(0, 500)
      try {
        await api.pingBatch(chunk)
      } catch (e) {
        if (isNetworkError(e)) break
      }
      write(read().slice(chunk.length))
    }
  } finally {
    flushing = false
  }
}

export function startFieldTracker() {
  try { const id = localStorage.getItem(ACTIVE); if (id) setTrackingBooking(id) } catch { /* */ }
  window.addEventListener('online', () => { void flushTrack() })
  // المتصفح يفك قفل الشاشة لما الصفحة تختفي — نرجع نطلبه من ترجع
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState !== 'visible') { wakeLock = null; return }
    if (activeBooking) void keepAwake()
  })
  setInterval(() => { void flushTrack() }, 60_000)
}
