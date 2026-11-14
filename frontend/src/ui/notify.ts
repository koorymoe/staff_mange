// ═══ 🔔 تنبيه موحّد بدل alert() (دازاي 11-15) ═══
// شريط صغير فوگ بالنص، يختفي لحاله بعد ~٥ ثواني، وإذا اجه أكثر من واحد يترتبون فوگ بعض.
// الشاشة تنرسم بـ <NotifyHost/> (main.tsx) — قبل ما تنرسم الرسائل تنتظر بالطابور.
//   notify('ما انحفظ')            ← خطأ (الافتراضي)
//   notify('تم الحفظ ✅', 'success')

export type NotifyKind = 'error' | 'success' | 'info'
export type Toast = { id: number; message: string; kind: NotifyKind }

export const NOTIFY_MS = 5000

let seq = 0
let listener: ((t: Toast) => void) | null = null
export const pendingToasts: Toast[] = []

export function onNotify(fn: ((t: Toast) => void) | null) { listener = fn }

export function notify(message: string, kind: NotifyKind = 'error'): void {
  const t: Toast = { id: ++seq, message: String(message ?? ''), kind }
  if (listener) listener(t)
  else pendingToasts.push(t)
}
