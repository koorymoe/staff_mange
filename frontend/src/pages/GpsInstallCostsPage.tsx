import { useEffect, useState } from 'react'
import { api, type GpsInstallCostSummary } from '../api'

/**
 * حساب تكاليف الشد — تفصيلي لكل الكوادر، ضمن خانة الحسابات.
 *
 * كان بشيت إكسل منفصل يتحدّث يدوياً؛ صار يتحسب حي من بيانات الشد نفسها،
 * فيتحدّث لحاله مع كل عملية شد جديدة.
 */
const money = (n: number) => n.toLocaleString('en-US') + ' د.ع'

export default function GpsInstallCostsPage() {
  const [data, setData] = useState<GpsInstallCostSummary | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getGpsInstallCosts()
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر حساب تكاليف الشد'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div dir="rtl" className="p-8 text-center text-slate-500">جاري الحساب...</div>
  if (error) return <div dir="rtl" className="rounded-xl bg-red-50 p-6 text-center text-red-700">{error}</div>
  if (!data) return null

  // نجمّع الصفوف بجدول: صف لكل شهر، عمود لكل موظف — نفس شكل الإكسل القديم.
  // القوائم ممكن ترجع فاضية أو null، فنحصّن قبل ما نمر عليها.
  const rows = data.rows ?? []
  const byEmployee = data.byEmployee ?? []
  const months = [...new Set(rows.map((r) => r.month))]
  const people = byEmployee.map((e) => e.employeeName)
  const cell = (month: string, name: string) =>
    rows.find((r) => r.month === month && r.employeeName === name)

  // ماكو بيانات شد بعد — نوضح السبب بدل شاشة فاضية أو كراش
  if (rows.length === 0) {
    return (
      <div dir="rtl" className="space-y-6">
        <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
          <h1 className="text-2xl font-bold text-white">🔧 حساب تكاليف الشد</h1>
          <p className="mt-1 text-sm text-blue-200">تكاليف شد أجهزة الجي بي اس — تفصيلي لكل الكوادر</p>
        </div>
        <div className="rounded-2xl bg-white p-8 text-center shadow-sm">
          <p className="text-lg font-bold text-slate-700">ماكو بيانات شد بعد</p>
          <p className="mx-auto mt-3 max-w-xl text-sm leading-7 text-slate-500">
            هذي الشاشة تحسب تكاليف <b>شد أجهزة الجي بي اس</b> من سجلات الشد نفسها (منفّذ الشد
            وتكلفته). تبقى فاضية لحد ما تنستورد بيانات الجي بي اس القديمة أو تنسجّل عمليات شد
            جديدة بالنظام.
            <br /><br />
            هي <b>غير</b> «حساب تكلفة التنصيب للتنفيذ» — ذيچ حاسبة تحسبلك كلفة تنفيذ شغلة قبل
            ما تسويها، وهذي تقرير يجمعلك شكد كلّفتنا عمليات الشد الي خلصت فعلاً.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">🔧 حساب تكاليف الشد</h1>
        <p className="mt-1 text-sm text-blue-200">تفصيلي لكل الكوادر — يتحدّث تلقائياً مع كل عملية شد</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <p className="text-sm text-slate-500">إجمالي التكاليف</p>
          <p className="mt-1 text-2xl font-bold" style={{ color: 'var(--t-title)' }}>{money(data.grandTotal)}</p>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <p className="text-sm text-slate-500">عدد عمليات الشد</p>
          <p className="mt-1 text-2xl font-bold text-emerald-600">{data.totalInstalls}</p>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <p className="text-sm text-slate-500">عدد الأشهر</p>
          <p className="mt-1 text-2xl font-bold text-blue-600">{data.monthCount}</p>
        </div>
      </div>

      <section className="rounded-2xl bg-white p-6 shadow-sm">
        <h2 className="mb-4 text-lg font-bold" style={{ color: 'var(--t-title)' }}>الإجمالي لكل كادر</h2>
        <div className="space-y-2">
          {data.byEmployee.slice().sort((a, b) => b.total - a.total).map((e) => (
            <div key={e.employeeName} className="flex items-center justify-between rounded-xl bg-slate-50 p-3">
              <span className="font-medium text-slate-800">{e.employeeName}</span>
              <span className="font-bold" style={{ color: 'var(--t-title)' }}>{money(e.total)}</span>
            </div>
          ))}
          {data.byEmployee.length === 0 && <p className="p-6 text-center text-slate-400">ماكو بيانات شد</p>}
        </div>
      </section>

      <section className="overflow-hidden rounded-2xl bg-white shadow-sm">
        <h2 className="p-6 pb-3 text-lg font-bold" style={{ color: 'var(--t-title)' }}>التفصيل الشهري</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-slate-200 text-right">
            <thead style={{ backgroundColor: '#1a3a5c' }} className="text-white">
              <tr>
                <th className="whitespace-nowrap px-4 py-3 text-sm font-semibold">الشهر</th>
                {people.map((p) => <th key={p} className="whitespace-nowrap px-4 py-3 text-sm font-semibold">{p}</th>)}
                <th className="whitespace-nowrap px-4 py-3 text-sm font-semibold">المجموع</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {months.map((m) => {
                const monthTotal = data.rows.filter((r) => r.month === m).reduce((s, r) => s + r.total, 0)
                return (
                  <tr key={m} className="hover:bg-slate-50">
                    <td className="whitespace-nowrap px-4 py-3 font-medium">{m}</td>
                    {people.map((p) => {
                      const c = cell(m, p)
                      return (
                        <td key={p} className="whitespace-nowrap px-4 py-3 text-sm text-slate-600">
                          {c ? <>{money(c.total)}<span className="block text-xs text-slate-400">{c.installs} شد</span></> : '—'}
                        </td>
                      )
                    })}
                    <td className="whitespace-nowrap px-4 py-3 font-bold" style={{ color: 'var(--t-title)' }}>{money(monthTotal)}</td>
                  </tr>
                )
              })}
              {months.length === 0 && (
                <tr><td colSpan={people.length + 2} className="p-8 text-center text-slate-400">ماكو بيانات</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  )
}
