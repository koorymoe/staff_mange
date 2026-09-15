import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'
import { getSwitches, forgetSwitches } from '../systemSwitches'

// ═══ زر إطفاء ميزة — للمالك حصراً ═══
//
// ⚠️ **مكوّن واحد مو نسختان**: مفتاحان اليوم (الكائن والإعلانات)
// والطلبات جايّة. ونسختان تتفرّقان بأول تصحيح — فيصير مفتاح يعرض
// «انطفى» وهو ما انطفى.
//
// 🔴 **والزر يختفي عن غير المالك بالكامل**: عرضه معطَّلاً يقول لكل
// موظف «أكو مفتاح يطفّي هذا الشي» — ومحد يحتاج يعرف. والخادم يرد
// ٤٠٣ لغير المالك أصلاً، فهذا **راحة عرض مو حماية**.

interface Props {
  /** مفتاح النظام — من `systemSwitches.ts`. */
  switchKey: string
  /** الاسم الي يشوفه (ع): «شخصية الكائن». */
  label: string
  /** جملة وحدة تشرح شنو يصير لمّا ينطفي. */
  hint: string
}

export default function OwnerSwitch({ switchKey, label, hint }: Props) {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [on, setOn] = useState<boolean | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!isOwner) return
    let alive = true
    getSwitches().then((all) => { if (alive) setOn(all[switchKey] !== false) })
    return () => { alive = false }
  }, [isOwner, switchKey])

  if (!isOwner) return null

  const toggle = async () => {
    if (on === null || busy) return
    setBusy(true)
    setError(null)
    try {
      const next = await api.setSystemSwitch(switchKey, !on)
      // ⚠️ **نقرا الجواب مو نفترضه**: الخادم هو صاحب القرار، ولو
      // رفض لأي سبب لازم الزر يبين الواقع مو نيّتنا.
      setOn(next[switchKey] !== false)
      // والمخزون يُنسى حتى بقية الشاشات تلقط التبديل بلا إعادة تحميل.
      forgetSwitches()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر حفظ المفتاح')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="rounded-xl border p-3" style={{ background: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-bold" style={{ color: 'var(--t-title)' }}>
            {label} — {on === null ? '…' : on ? 'شغّالة' : '⛔ مطفية'}
          </p>
          <p className="mt-0.5 text-xs" style={{ color: 'var(--t-muted)' }}>{hint}</p>
        </div>
        <button onClick={toggle} disabled={busy || on === null}
          className={`shrink-0 rounded-xl px-4 py-2 text-sm font-bold text-white transition-colors disabled:opacity-50 ${
            on ? 'bg-rose-600 hover:bg-rose-700' : 'bg-emerald-600 hover:bg-emerald-700'
          }`}>
          {busy ? '…' : on ? 'أطفيها لكل الموظفين' : 'شغّلها لكل الموظفين'}
        </button>
      </div>
      {/* ⚠️ الإطفاء يسري على **كل** الموظفين بأول تحميل صفحة عندهم —
          وهاي تنكتب صراحةً حتى (ع) ما يتفاجأ إن موظفاً بعده يشوفها
          بتبويب مفتوح من قبل التبديل. */}
      <p className="mt-1.5 text-[11px]" style={{ color: 'var(--t-faint)' }}>
        القرار يسري على كل الموظفين — والي عنده النظام مفتوح الآن يشوف
        التغيير بأول تحديث للصفحة.
      </p>
      {error && <p className="mt-1 text-xs font-bold text-rose-600">{error}</p>}
    </div>
  )
}
