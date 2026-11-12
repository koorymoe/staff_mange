import { Link } from 'react-router-dom'
import type { MatrixWatch } from '../api'
import { GROUP_LABEL, type EyeGroup } from './matrixEyeColors'

// ═══ إثبات الأرقام — (ع) 10-13: «من اضغط عليهن يوديني لشي يثبت صحتهن» ═══
export type ProofKind = 'stability' | 'attention' | 'people' | 'tasks'
type WatchGroup = { group: string; employees: MatrixWatch[]; red: number; alert: number }

export default function MatrixProofModal({ kind, groups, onClose }: { kind: ProofKind; groups: WatchGroup[]; onClose: () => void }) {
  const all = groups.flatMap((g) => g.employees.map((e) => ({ ...e, group: g.group })))
  const people = all.length
  const red = all.filter((e) => e.level === 'RED')
  const alert = all.filter((e) => e.level === 'ALERT')
  const name = (e: MatrixWatch) => e.name || '—'
  const glabel = (g: string) => GROUP_LABEL[g as EyeGroup] ?? g
  const emp = (e: MatrixWatch, extra?: React.ReactNode) => (
    <li key={e.id} className="rounded-lg bg-[var(--mx-sunken)] px-3 py-2">
      <div className="flex items-center justify-between gap-2">
        {e.id ? <Link to={`/matrix/employee/${e.id}`} className="font-bold text-[var(--mx-title)] hover:underline">{name(e)} ←</Link> : <b>{name(e)}</b>}
        <span className="text-[11px] text-[var(--mx-muted)]">{glabel(e.group)} · {e.level === 'RED' ? '🔴 أحمر' : e.level === 'ALERT' ? '🟡 منتبه' : '🟢 مستقر'}</span>
      </div>
      {extra}
    </li>
  )
  const reasons = (e: MatrixWatch) => e.items.length > 0 && (
    <ul className="mt-1 space-y-0.5 text-[12px] text-[var(--mx-muted)]">
      {e.items.map((it, i) => <li key={i}>• {it.escalated ? '⬆️ ' : ''}{it.summary} <span className="text-[10px]">({it.label} · من {new Date(it.since).toLocaleDateString('ar-IQ')})</span></li>)}
    </ul>
  )
  const work = (e: MatrixWatch) => (
    <ul className="mt-1 space-y-0.5 text-[12px]">
      {e.workload.map((w) => (
        <li key={w.key} className="flex items-center justify-between gap-2">
          <Link to={w.route} className="text-[var(--mx-accent)] hover:underline">{w.label} ←</Link>
          <span className="tabular-nums text-[var(--mx-muted)]">{w.verb} {w.done} · باقي {w.left} = <b className="text-[var(--mx-title)]">{w.done + w.left}</b></span>
        </li>
      ))}
    </ul>
  )
  let title: string, how: string, body: React.ReactNode
  if (kind === 'stability') {
    // (ل) 11-12 دازاي: «كل نسبة بالنظام لازم الها سبب واضح ومفهوم للقارئ» — الحسبة بالأرقام،
    // معنى كل لون، توزيعها على المجموعات، وكل أحمر ليش صار أحمر.
    const pct = people ? Math.round(((people - red.length) / people) * 100) : 100
    const each = people ? Math.round((100 / people) * 10) / 10 : 0
    const whyRed = (e: MatrixWatch) => {
      const esc = e.items.some((it) => it.escalated)
      const open = e.items.filter((it) => !it.info).length
      return esc ? `تذكير ما انحل وصعد للمدير${open > 1 ? ` + ${open - 1} تذكير مفتوح ثاني` : ''}` : `${open} تذكيرات مفتوحة (٣ أو أكثر = أحمر)`
    }
    title = `〰️ استقرار النظام ${pct}%`
    how = `${people - red.length} موظف مو بالأحمر من أصل ${people} موظف تحت المراقبة (${people} ناقص ${red.length} أحمر) = ${pct}%`
    body = (
      <div className="space-y-4">
        <div className="space-y-1 rounded-lg border border-[var(--mx-border)] p-3 text-[13px]">
          <p className="font-extrabold text-[var(--mx-title)]">شنو تعني النسبة؟</p>
          <p>من كل الموظفين الي ماتركس يراقب شغلهم، شكد منهم <b>مو بالأحمر</b>. كل موظف بالأحمر ينزّلها تقريباً {each}%.</p>
          <p>🔴 <b>أحمر</b> = عنده تذكير من ماتركس ما انحل وصعد للمدير، أو ٣ تذكيرات مفتوحة أو أكثر بنفس الوقت — <b>هذا الي ينزّل النسبة</b>.</p>
          <p>🟡 <b>منتبه</b> = عنده تذكير أو تذكيرين مفتوحة — ما ينزّل النسبة، بس إذا وصل ٣ يصير أحمر.</p>
          <p>🟢 <b>مستقر</b> = ماكو عليه تذكير مفتوح.</p>
          <p className="text-[var(--mx-muted)]">التذكير ينفتح لمن ماتركس يلگى شغل متأخر أو ناقص بدور الموظف (مثلاً حجز ما انقفل، تقييم ما انسوّى، جرد ما انعمل)، وينسد لحاله أول ما الموظف يخلّصه. تنبيهات «للعلم» ما تنحسب. المالك والمدير ما ينحسبون.</p>
          {red.length > 0 && <p className="font-bold text-[var(--mx-ok)]">إذا انحلّت تذكيرات الـ{red.length} الحمر، النسبة تصير ١٠٠%.</p>}
        </div>
        <div>
          <p className="mb-1 text-sm font-extrabold text-[var(--mx-title)]">حسب المجموعة</p>
          <table className="w-full text-[13px]">
            <thead><tr className="text-[var(--mx-muted)]"><th className="text-right">المجموعة</th><th>الكل</th><th>🔴</th><th>🟡</th><th>🟢</th><th>النسبة</th></tr></thead>
            <tbody>
              {groups.filter((g) => g.employees.length).map((g) => {
                const r = g.employees.filter((e) => e.level === 'RED').length
                const a = g.employees.filter((e) => e.level === 'ALERT').length
                const n = g.employees.length
                return (
                  <tr key={g.group} className="border-t border-[var(--mx-border)] text-center tabular-nums">
                    <td className="py-1 text-right font-bold">{glabel(g.group)}</td><td>{n}</td><td>{r}</td><td>{a}</td><td>{n - r - a}</td>
                    <td className="font-bold">{Math.round(((n - r) / n) * 100)}%</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
        <div>
          <p className="mb-1 text-sm font-extrabold text-[var(--mx-title)]">الحمر ({red.length}) وسبب كل واحد</p>
          {red.length ? <ul className="space-y-1.5">{red.map((e) => emp(e, <><p className="mt-1 text-[12px] font-bold text-[var(--mx-bad)]">ليش أحمر: {whyRed(e)}</p>{reasons(e)}</>))}</ul> : <p>ماكو موظف أحمر ✅</p>}
        </div>
      </div>
    )
  } else if (kind === 'attention') {
    title = `⚠️ ${red.length + alert.length} حالة تحتاج انتباه`
    how = `= الحمر (${red.length}) + المنتبه عليهم (${alert.length}). تحت كل اسم التذكيرات المفتوحة الي خلّته هيچ.`
    body = <ul className="space-y-1.5">{[...red, ...alert].map((e) => emp(e, reasons(e)))}</ul>
  } else if (kind === 'people') {
    title = `👥 ${people} موظف تحت المراقبة`
    how = 'كل موظف فعّال عنده دور يراقبه ماتركس (المالك والمدير ما ينحسبون). مجمّعين حسب المجموعة.'
    body = (
      <div className="space-y-3">
        {groups.filter((g) => g.employees.length).map((g) => (
          <div key={g.group}>
            <p className="mb-1 text-sm font-extrabold text-[var(--mx-title)]">{glabel(g.group)} ({g.employees.length})</p>
            <ul className="grid gap-1 sm:grid-cols-2">{g.employees.map((e) => emp({ ...e, group: g.group }))}</ul>
          </div>
        ))}
      </div>
    )
  } else {
    const withWork = all.filter((e) => e.workload.length > 0)
    const total = all.reduce((n, e) => n + e.workload.reduce((k, w) => k + w.done + w.left, 0), 0)
    title = `📦 ${total} شغلة اليوم`
    how = 'مو «مهام» بالمعنى العادي: هو مجموع شغل اليوم لكل موظف حسب دوره (مثلاً حجوزات يدققها المراقب، فواتير يعتمدها المحاسب، ورق الليدر) = الي خلّصه + الي باقي عليه. كل سطر يودّيك لشاشته.'
    body = withWork.length ? <ul className="space-y-1.5">{withWork.map((e) => emp(e, work(e)))}</ul> : <p>ماكو شغل محسوب اليوم.</p>
  }
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4" onClick={onClose}>
      <div dir="rtl" className="w-full max-w-3xl rounded-2xl bg-[var(--mx-bg)] p-5 text-[var(--mx-text)] shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-2 flex items-center justify-between">
          <h3 className="text-lg font-extrabold text-[var(--mx-title)]">{title}</h3>
          <button onClick={onClose} className="text-2xl text-[var(--mx-muted)]">×</button>
        </div>
        <p className="mb-3 rounded-lg bg-[var(--mx-sunken)] px-3 py-2 text-[12px]">🧮 {how}</p>
        <div className="max-h-[70vh] overflow-y-auto">{body}</div>
      </div>
    </div>
  )
}

