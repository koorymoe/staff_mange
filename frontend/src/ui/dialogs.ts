// ═══ 🪟 نوافذ التأكيد والإدخال الموحّدة (تعديلات ٧ب — 11-04) ═══
// بدل `confirm()` و`prompt()` مال المتصفح (ممنوعات — قاعدة دازاي): نفس الاستعمال
// بس بـ await، وبتصميم النظام وبالعربي. الشاشة تنرسم بـ <DialogHost/> (main.tsx).
//   if (!(await askConfirm('تحذف؟'))) return
//   const reason = await askInput('اكتب السبب:'); if (reason === null) return

export type DialogReq =
  | { id: number; kind: 'confirm'; message: string; okLabel?: string; danger?: boolean; resolve: (v: boolean) => void }
  | { id: number; kind: 'input'; message: string; defaultValue?: string; multiline?: boolean; resolve: (v: string | null) => void }

let seq = 0
let listener: ((r: DialogReq) => void) | null = null
export const pendingDialogs: DialogReq[] = []

export function onDialog(fn: ((r: DialogReq) => void) | null) { listener = fn }

function send(r: DialogReq) {
  if (listener) listener(r)
  else pendingDialogs.push(r)
}

export function askConfirm(message: string, opts: { okLabel?: string; danger?: boolean } = {}): Promise<boolean> {
  return new Promise((resolve) => send({ id: ++seq, kind: 'confirm', message, ...opts, resolve }))
}

export function askInput(message: string, defaultValue = '', opts: { multiline?: boolean } = {}): Promise<string | null> {
  return new Promise((resolve) => send({ id: ++seq, kind: 'input', message, defaultValue, ...opts, resolve }))
}
