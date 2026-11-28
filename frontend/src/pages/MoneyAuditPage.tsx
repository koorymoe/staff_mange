import { useEffect, useState } from 'react'
import { api } from '../api'
import { card } from '../ui/classes'

// ═══ 🔎 فحص الأرقام (دازاي 11-26) ═══
// «هل كل الفواتير بليدر؟ هل كل المبالغ حسبها النظام؟ الصافي صحيح؟ شكد إدخال الشركة؟ التقييم صحيح؟»
// قراءة بس من بيانات النظام الحقيقية، الفترة من 26 لـ25.

type Item = { ref: string; label: string; amount: number; date: string }
type Check = { key: string; title: string; why: string; ok: boolean; warn?: boolean; count: number; value?: string; items?: Item[] }
type Section = { title: string; checks: Check[] }
type Line = { label: string; amount: number; note?: string; bold?: boolean }
type Audit = { period: string; from: string; to: string; problems: number; money: Line[]; sections: Section[] }

const num = (v: number) => Math.round(v).toLocaleString('en-US')

function shiftPeriod(p: string, k: number) {
  const [y, m] = p.split('-').map(Number)
  const d = new Date(Date.UTC(y, m - 1 + k, 1))
  return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`
}

function CheckRow({ c }: { c: Check }) {
  const [open, setOpen] = useState(false)
  const icon = !c.ok ? '❌' : c.warn ? '⚠️' : '✅'
  const tone = !c.ok ? 'border-red-200 bg-red-50' : c.warn ? 'border-amber-200 bg-amber-50' : 'border-emerald-200 bg-emerald-50'
  const hasItems = (c.items?.length ?? 0) > 0
  return (
    <div className={`rounded-xl border p-3 ${tone}`}>
      <button type="button" className="flex w-full items-start gap-2 text-right" onClick={() => hasItems && setOpen(!open)}>
        <span className="text-lg leading-none">{icon}</span>
        <span className="flex-1">
          <span className="block font-bold text-slate-800">
            {c.title}{c.count > 0 && <span className="mr-2 rounded-full bg-white px-2 text-sm">{c.count}</span>}
          </span>
          <span className="block text-xs text-slate-600">{c.why}</span>
          {c.value && <span className="mt-1 block text-sm font-semibold text-slate-800">{c.value}</span>}
        </span>
        {hasItems && <span className="text-xs text-slate-500">{open ? 'إخفاء ▲' : 'التفاصيل ▼'}</span>}
      </button>
      {open && hasItems && (
        <div className="mt-2 overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="text-xs text-slate-500"><tr><th className="text-right">الحجز / الموظف</th><th className="text-right">شنو</th><th>المبلغ</th><th>التاريخ</th></tr></thead>
            <tbody>
              {c.items!.map((x, i) => (
                <tr key={i} className="border-t border-white/70">
                  <td className="py-1 font-mono">{x.ref}</td><td>{x.label}</td>
                  <td className="text-center" dir="ltr">{x.amount ? num(x.amount) : ''}</td><td className="text-center" dir="ltr">{x.date}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {c.count > (c.items?.length ?? 0) && <p className="mt-1 text-xs text-slate-500">يطلع أول {c.items!.length} من {c.count}.</p>}
        </div>
      )}
    </div>
  )
}

export default function MoneyAuditPage() {
  const [period, setPeriod] = useState('')
  const [a, setA] = useState<Audit | null>(null)
  const [err, setErr] = useState('')
  const [doneFor, setDoneFor] = useState<string | null>(null)
  const busy = doneFor !== period
  useEffect(() => {
    let alive = true
    api.getMoneyAudit<Audit>(period)
      .then((x) => { if (alive) { setA(x); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الفحص') })
      .finally(() => { if (alive) setDoneFor(period) })
    return () => { alive = false }
  }, [period])
  return (
    <div className="space-y-4" dir="rtl">
      <div className={card}>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h1 className="text-xl font-extrabold text-[#0f2040]">🔎 فحص الأرقام</h1>
            <p className="text-sm text-slate-500">يفحص الفواتير والمبالغ والصافي والتقييم من بيانات النظام الحقيقية. قراءة بس، ما يغيّر شي.</p>
          </div>
          {a && (
            <div className="flex items-center gap-2 text-sm">
              <button type="button" className="rounded-lg border px-2 py-1" onClick={() => setPeriod(shiftPeriod(a.period, -1))}>→ السابقة</button>
              <span className="font-bold" dir="ltr">{a.from} … {a.to}</span>
              <button type="button" className="rounded-lg border px-2 py-1" onClick={() => setPeriod(shiftPeriod(a.period, 1))}>اللاحقة ←</button>
            </div>
          )}
        </div>
        {a && (
          <p className={`mt-3 rounded-lg p-2 text-center font-bold ${a.problems ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'}`}>
            {a.problems ? `❌ ${a.problems} فحص بيه مشكلة — افتح التفاصيل` : '✅ كل الفحوصات سليمة'}
          </p>
        )}
      </div>
      {busy && <p className="text-center text-slate-400">جاري الفحص…</p>}
      {err && <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>}
      {a && !busy && (
        <>
          <section className={card}>
            <h2 className="mb-2 text-lg font-extrabold text-[#0f2040]">💰 إدخال الشركة والصافي</h2>
            <table className="w-full text-sm">
              <tbody>
                {a.money.map((l, i) => (
                  <tr key={i} className={`border-t ${l.bold ? 'bg-slate-50 font-extrabold' : ''}`}>
                    <td className="py-1.5">{l.label}{l.note && <span className="mr-2 text-xs font-normal text-slate-500">({l.note})</span>}</td>
                    <td className={`text-left ${l.amount < 0 ? 'text-red-600' : ''}`} dir="ltr">{num(l.amount)} د.ع</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p className="mt-2 text-xs text-slate-500">ما داخل بالحساب هسه: أي فلوس برّا الفواتير ودفعات المشاريع وتجديد الـGPS (مثل بيع جهاز GPS بلا فاتورة).</p>
          </section>
          {a.sections.map((s) => (
            <section key={s.title} className={card}>
              <h2 className="mb-2 text-lg font-extrabold text-[#0f2040]">{s.title}</h2>
              <div className="space-y-2">{s.checks.map((c) => <CheckRow key={c.key} c={c} />)}</div>
            </section>
          ))}
        </>
      )}
    </div>
  )
}
