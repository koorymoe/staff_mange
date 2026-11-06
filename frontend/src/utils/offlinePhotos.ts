import { api } from '../api'
import { isNetworkError } from './offlineStages'

// ═══ صور الشغل بلا نت (دازاي 11-06، فكرة ٣) ═══
// الصورة الي ما تنرفع لأن ماكو نت تنحفظ بـIndexedDB (localStorage ما يشيل صور) وترفع وحدها
// من يرجع النت — وتنرفع **قبل** ضغطات المراحل حتى «تم» ما ينرفض على شرط الصور.

type Queued = { id: string; bookingId: string; phase: 'BEFORE' | 'AFTER'; blob: Blob; at: string; tries: number }

const DB = 'amani_offline_v1'
const STORE = 'photos'

function open(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const r = indexedDB.open(DB, 1)
    r.onupgradeneeded = () => r.result.createObjectStore(STORE, { keyPath: 'id' })
    r.onsuccess = () => resolve(r.result)
    r.onerror = () => reject(r.error)
  })
}

async function tx<T>(mode: IDBTransactionMode, fn: (s: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  const db = await open()
  return new Promise((resolve, reject) => {
    const req = fn(db.transaction(STORE, mode).objectStore(STORE))
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

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
    await tx('readwrite', (s) => s.put({ id: crypto.randomUUID?.() ?? String(Date.now()), bookingId, phase, blob, at: new Date().toISOString(), tries: 0 }))
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
    const items = (await pendingPhotos()).sort((a, b) => a.at.localeCompare(b.at))
    for (const it of items) {
      try {
        await api.uploadWorkPhoto(it.bookingId, it.phase, it.blob)
        await tx('readwrite', (s) => s.delete(it.id))
        sent++
      } catch (e) {
        if (isNetworkError(e)) break
        if (it.tries + 1 >= 3) await tx('readwrite', (s) => s.delete(it.id))
        else await tx('readwrite', (s) => s.put({ ...it, tries: it.tries + 1 }))
      }
    }
  } finally {
    flushing = false
    window.dispatchEvent(new Event('photo-queue'))
  }
  return sent
}
