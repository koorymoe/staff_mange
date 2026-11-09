import { api } from '../api'
import { currentEmployeeId, isNetworkError } from './offlineStages'

// ═══ صور الشغل بلا نت (دازاي 11-06، فكرة ٣) ═══
// الصورة الي ما تنرفع لأن ماكو نت تنحفظ بـIndexedDB (localStorage ما يشيل صور) وترفع وحدها
// من يرجع النت — وتنرفع **قبل** ضغطات المراحل حتى «تم» ما ينرفض على شرط الصور.

// employeeId (11-07): صاحب الصورة — بموبايل مشترك ما ترتفع بجلسة زميله
type Queued = { id: string; employeeId?: string; bookingId: string; phase: 'BEFORE' | 'AFTER'; blob: Blob; at: string; tries: number }

const DB = 'amani_offline_v1'
const STORE = 'photos'

// (مراجعة 11-06) اتصال واحد مخزّن — يتسكّر إذا نسخة ثانية من التطبيق طلبت ترقية القاعدة
let dbp: Promise<IDBDatabase> | null = null
function open(): Promise<IDBDatabase> {
  if (dbp) return dbp
  dbp = new Promise((resolve, reject) => {
    const r = indexedDB.open(DB, 1)
    r.onupgradeneeded = () => r.result.createObjectStore(STORE, { keyPath: 'id' })
    r.onsuccess = () => {
      const db = r.result
      db.onversionchange = () => { db.close(); dbp = null }
      db.onclose = () => { dbp = null }
      resolve(db)
    }
    r.onerror = () => { dbp = null; reject(r.error) }
  })
  return dbp
}

// ينحسب «انحفظ» بس من تكمل المعاملة (oncomplete) — الطلب ممكن ينجح والمعاملة تنلغي (ذاكرة ممتلئة)
async function tx<T>(mode: IDBTransactionMode, fn: (s: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  const db = await open()
  return new Promise((resolve, reject) => {
    const t = db.transaction(STORE, mode)
    const req = fn(t.objectStore(STORE))
    let result: T
    req.onsuccess = () => { result = req.result }
    t.oncomplete = () => resolve(result)
    t.onabort = () => reject(t.error ?? new Error('ذاكرة الموبايل ممتلئة — ما انحفظت الصورة'))
    t.onerror = () => reject(t.error)
  })
}

const isAuthError = (e: unknown) => e instanceof Error && /401|تسجيل الدخول|الجلسة|انتهت/.test(e.message)

export async function pendingPhotos(bookingId?: string): Promise<Queued[]> {
  try {
    const all = await tx<Queued[]>('readonly', (s) => s.getAll() as IDBRequest<Queued[]>)
    return bookingId ? all.filter((q) => q.bookingId === bookingId) : all
  } catch {
    return []
  }
}

/** يرفع الصورة؛ بلا نت ينحفظ بالموبايل ويرجّع queued=true. */
export async function uploadOrQueue(bookingId: string, phase: 'BEFORE' | 'AFTER', blob: Blob): Promise<{ queued: boolean }> {
  const queue = async () => {
    try {
      await tx('readwrite', (s) => s.put({ id: crypto.randomUUID?.() ?? String(Date.now()), employeeId: currentEmployeeId(), bookingId, phase, blob, at: new Date().toISOString(), tries: 0 }))
    } catch {
      throw new Error('ماكو نت وذاكرة الموبايل ما تشيل الصورة — فرّغ مساحة أو صوّرها من يرجع النت')
    }
    window.dispatchEvent(new Event('photo-queue'))
    return { queued: true }
  }
  if (!navigator.onLine) return queue()
  try {
    await api.uploadWorkPhoto(bookingId, phase, blob)
    return { queued: false }
  } catch (e) {
    if (isNetworkError(e)) return queue()
    throw e
  }
}

let flushing = false
/** يرفع الصور المحفوظة بالترتيب. رفض السيرفر ٣ مرات = تنشال (حتى ما تعلق للأبد). */
export async function flushPhotos(): Promise<number> {
  if (flushing || !navigator.onLine) return 0
  flushing = true
  let sent = 0
  try {
    const me = currentEmployeeId()
    const items = (await pendingPhotos()).sort((a, b) => a.at.localeCompare(b.at))
    for (const it of items) {
      // صورة زميل (موبايل مشترك): تبقى لحد ما صاحبها يدخل — ما تنرفع باسم غيره ولا تنحسب رفض
      if (it.employeeId && it.employeeId !== me) continue
      try {
        await api.uploadWorkPhoto(it.bookingId, it.phase, it.blob)
        await tx('readwrite', (s) => s.delete(it.id))
        sent++
      } catch (e) {
        // بلا نت أو الجلسة منتهية: نوقف ونخليها — ما تنحسب رفض
        if (isNetworkError(e) || isAuthError(e)) break
        if (it.tries + 1 >= 3) {
          await tx('readwrite', (s) => s.delete(it.id))
          window.dispatchEvent(new CustomEvent('photo-dropped', { detail: { bookingId: it.bookingId, reason: e instanceof Error ? e.message : '' } }))
        } else await tx('readwrite', (s) => s.put({ ...it, tries: it.tries + 1 }))
      }
    }
  } finally {
    flushing = false
    window.dispatchEvent(new Event('photo-queue'))
  }
  return sent
}
