import FormDialog, { type Field } from '../components/FormDialog'
import { createRoot } from 'react-dom/client'

// ═══ نوافذ داخل النظام بدل نوافذ المتصفح (قرار (ع) 10-07) ═══
// «هاي الخيارات… المفروض تكون من ضمن النظام» — prompt/alert الخام تطلع
// فوگ الصفحة بنافذة المتصفح («staffmanage.cc says») ويا «اكتب رقم الخيار».
// هنا نفس الأسئلة بنافذة من النظام: أزرار بدل أرقام، ونص عربي مرتب.

function mount<T>(render: (done: (v: T) => void) => React.ReactNode): Promise<T> {
  return new Promise((resolve) => {
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    const done = (v: T) => { root.unmount(); host.remove(); resolve(v) }
    root.render(render(done))
  })
}

/** نافذة أسئلة من النظام. ترجع القيم، أو null إذا انلغت. */
export function askForm(title: string, fields: Field[], ok = 'تمام'): Promise<Record<string, string> | null> {
  return mount((done) => <FormDialog title={title} fields={fields} ok={ok} onDone={done} />)
}

/** اختيار واحد من قائمة — بديل promptChoice. */
export async function askChoice<T extends string>(title: string, options: [T, string][]): Promise<T | null> {
  const v = await askForm(title, [{ kind: 'choice', key: 'c', label: '', options }])
  return (v?.c as T) ?? null
}

// ═══ تنبيه داخل النظام بدل alert() ═══
// alert ما ينتظر جواب، فنبدّله بشريط يطلع جوّا الصفحة ويختفي لحاله.
function showNotice(msg: string) {
  const el = document.createElement('div')
  el.dir = 'rtl'
  el.setAttribute('role', 'status')
  el.className = 'fixed left-1/2 top-4 z-[110] max-w-[90vw] -translate-x-1/2 rounded-xl bg-[#0f2040] px-5 py-3 text-sm font-bold text-white shadow-2xl'
  el.style.whiteSpace = 'pre-line'
  el.textContent = msg
  el.onclick = () => el.remove()
  document.body.appendChild(el)
  window.setTimeout(() => el.remove(), Math.min(9000, 3000 + msg.length * 60))
}

/** يبدّل alert() بكل النظام بتنبيه من النظام نفسه. */
export function installInAppAlert() {
  window.alert = (msg?: unknown) => showNotice(String(msg ?? ''))
}
