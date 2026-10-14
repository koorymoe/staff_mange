// عميل API تطبيق الزبائن — كل المسارات تحت /api/c/
const TOKEN_KEY = 'amani_customer_token'

export function getToken(): string | null {
  try { return localStorage.getItem(TOKEN_KEY) } catch { return null }
}
export function setToken(t: string | null) {
  try { if (t) localStorage.setItem(TOKEN_KEY, t); else localStorage.removeItem(TOKEN_KEY) } catch { /* وضع خاص */ }
}

export class ApiError extends Error {
  status: number
  data: Record<string, unknown>
  constructor(status: number, data: Record<string, unknown>) {
    super(String(data.error || 'صار خطأ — حاول مرة ثانية'))
    this.status = status
    this.data = data
  }
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {}
  const t = getToken()
  if (t) headers.Authorization = `Bearer ${t}`
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  let res: Response
  try {
    res = await fetch(`/api/c${path}`, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) })
  } catch {
    throw new ApiError(0, { error: 'ماكو اتصال بالإنترنت' })
  }
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    if (res.status === 401 && t) { setToken(null); window.dispatchEvent(new Event('amani-logout')) }
    throw new ApiError(res.status, data)
  }
  return data as T
}

export interface Account { id: string; username: string; email: string; name: string; phone: string; createdAt: string }
export interface Session { token: string; account: Account }
export interface Promo { id: string; kind: 'offer' | 'work' | 'service'; title: string; body: string | null; hasImage: boolean; serviceId: string | null; createdAt: string; isNew: boolean }
export interface Service { id: string; name: string; category: string | null }
export interface Home { items: Promo[]; services: Service[]; newCount: number; completedJobs: number; matrixOn: boolean }
export interface CBooking { id: string; code: string; status: string; statusLabel: string; services: string; address?: string; scheduledAt?: string; completedAt?: string; createdAt: string; fromApp: boolean }
export interface ChatMsg { id: string; role: 'user' | 'matrix'; content: string; bookingCode?: string; createdAt: string }

export const imgUrl = (id: string) => `/api/c/img/${id}`

export const api = {
  signup: (b: { name: string; username: string; email: string; phone: string; password: string }) => req<{ username: string; email: string }>('POST', '/signup', b),
  verify: (username: string, code: string) => req<Session>('POST', '/verify', { username, code }),
  resend: (username: string, purpose: 'signup' | 'reset') => req<{ email: string }>('POST', '/resend', { username, purpose }),
  login: (username: string, password: string) => req<Session>('POST', '/login', { username, password }),
  forgot: (username: string) => req<{ email: string }>('POST', '/forgot', { username }),
  reset: (username: string, code: string, password: string) => req<Session>('POST', '/reset', { username, code, password }),
  me: () => req<Account>('GET', '/me'),
  updateMe: (b: { name: string; phone: string }) => req<Account>('PUT', '/me', b),
  home: () => req<Home>('GET', '/home'),
  seen: () => req<{ ok: boolean }>('POST', '/home/seen'),
  bookings: () => req<CBooking[]>('GET', '/bookings'),
  chat: () => req<{ messages: ChatMsg[]; matrixOn: boolean }>('GET', '/chat'),
  send: (text: string) => req<ChatMsg>('POST', '/chat', { text }),
}
