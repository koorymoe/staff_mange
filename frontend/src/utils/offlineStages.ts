import { api, type Booking } from '../api'

// ═══ مراحل الحجز بلا نت — قرار (ع) 10-14 ═══
//
// الليدر يدوس «استلام / المواد / انطلاق / وصلنا / بدأنا / تم الإنجاز» حتى لو ماكو نت عند الزبون:
// الضغطة تنحفظ بالموبايل **بوقتها الحقيقي وموقعها** (GPS ما يحتاج نت)،
// وترسل وحدها من يرجع النت. السيرفر يصحّح وقت المرحلة لوقت الضغطة.
// فـ«ماعندي نت» ما يصير عذر للتسجيل بالشركة آخر اليوم.

export type Stage = 'ACCEPTED' | 'MATERIALS' | 'DEPARTED' | 'ARRIVED' | 'STARTED' | 'COMPLETED'

interface QueuedStage {
  id: string
  bookingId: string
  code: string
  stage: Stage
  pressedAt: string
  lat?: number
  lng?: number
  accuracy?: number
  payload?: { missingToolIds?: string[]; completionNotes?: string; amountCollected?: number; advancePaid?: number }
}

const KEY = 'amani_stage_queue_v1'
const CACHE = 'amani_my_tasks_cache_v1'

function read(): QueuedStage[] {
  try { return JSON.parse(localStorage.getItem(KEY) || '[]') as QueuedStage[] } catch { return [] }
}
function write(q: QueuedStage[]) {
  try { localStorage.setItem(KEY, JSON.stringify(q)) } catch { /* ممتلئ/خاص */ }
  window.dispatchEvent(new Event('stage-queue'))
}
export const pendingStages = () => read()

// الموقع من الأقمار الصناعية — يشتغل بلا نت. مهلة ٨ ثواني وبعدها نكمل بلاه.
export function currentPosition(): Promise<{ lat: number; lng: number; accuracy: number } | null> {
  return new Promise((resolve) => {
    if (!('geolocation' in navigator)) return resolve(null)
    navigator.geolocation.getCurrentPosition(
      (p) => resolve({ lat: p.coords.latitude, lng: p.coords.longitude, accuracy: p.coords.accuracy }),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 60000 },
    )
  })
}

// خطأ شبكة (مو رفض من السيرفر) → نحفظ الضغطة بدل ما نضيّعها
export const isNetworkError = (e: unknown) => !navigator.onLine || e instanceof TypeError ||
  (e instanceof Error && /fetch|network|تعذر الاتصال|الاتصال بالخادم/i.test(e.message))

function call(item: QueuedStage): Promise<unknown> {
  if (item.stage === 'ACCEPTED') return api.acceptBooking(item.bookingId)
  if (item.stage === 'MATERIALS') return api.setMaterialsReady(item.bookingId)
  if (item.stage === 'DEPARTED') return api.departBooking(item.bookingId, { lat: item.lat, lng: item.lng })
  if (item.stage === 'ARRIVED') return api.markArrived(item.bookingId)
  if (item.stage === 'STARTED') return api.startBooking(item.bookingId, item.payload?.missingToolIds)
  return api.completeBooking(item.bookingId, item.payload ?? {})
}

/**
 * runStage: ينفّذ المرحلة. بالنت → ينرسل فوراً ويّا دليل الموقع.
 * بلا نت → ينحفظ بالطابور ويرجّع null (الواجهة تعرض «محفوظ — ينرسل من يرجع النت»).
 */
export type StagePos = { lat: number; lng: number; accuracy: number } | null
export async function runStage<T>(booking: Booking, stage: Stage, fn: (pos: StagePos) => Promise<T>,
  payload?: QueuedStage['payload']): Promise<{ result: T | null; queued: boolean }> {
  const pressedAt = new Date().toISOString()
  const pos = await currentPosition()
  const queue = () => {
    write([...read(), { id: crypto.randomUUID?.() ?? String(Date.now()), bookingId: booking.id, code: booking.code, stage, pressedAt, ...(pos ?? {}), payload }])
    return { result: null, queued: true }
  }
  if (!navigator.onLine) return queue()
  try {
    const result = await fn(pos)
    api.stageProof(booking.id, { stage, pressedAt, offline: false, ...(pos ?? {}) }).catch(() => {})
    return { result, queued: false }
  } catch (e) {
    if (isNetworkError(e)) return queue()
    throw e
  }
}

let flushing = false
/** يرسل الضغطات المحفوظة بالترتيب. يشتغل وحده من يرجع النت. */
export async function flushStages(): Promise<number> {
  if (flushing || !navigator.onLine) return 0
  flushing = true
  let sent = 0
  try {
    for (const item of read()) {
      try {
        await call(item)
      } catch (e) {
        if (isNetworkError(e)) break
        // السيرفر رفض (مثلاً المرحلة انسجّلت أصلاً) — نكمل بالدليل حتى الوقت يتصحّح
      }
      try {
        await api.stageProof(item.bookingId, { stage: item.stage, pressedAt: item.pressedAt, offline: true, lat: item.lat, lng: item.lng, accuracy: item.accuracy })
      } catch (e) {
        // بلا نت أو الجلسة منتهية → نوقف ونخلي الضغطة بالطابور (ما تضيع)
        if (isNetworkError(e) || (e instanceof Error && /تسجيل الدخول|الجلسة|صلاحية/.test(e.message))) break
      }
      write(read().filter((x) => x.id !== item.id))
      sent++
    }
  } finally {
    flushing = false
  }
  return sent
}

export function startStageSync() {
  window.addEventListener('online', () => { void flushStages() })
  setInterval(() => { void flushStages() }, 60_000)
  void flushStages()
}

// كاش «مهامي» حتى الشاشة تفتح بلا نت
export function cacheMyTasks(rows: Booking[]) {
  try { localStorage.setItem(CACHE, JSON.stringify({ at: Date.now(), rows })) } catch { /* */ }
}
export function cachedMyTasks(): Booking[] | null {
  try {
    const c = JSON.parse(localStorage.getItem(CACHE) || 'null') as { at: number; rows: Booking[] } | null
    return c && Date.now() - c.at < 36 * 3600_000 ? c.rows : null
  } catch { return null }
}
