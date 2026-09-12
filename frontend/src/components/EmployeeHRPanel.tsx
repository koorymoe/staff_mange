import { useState } from 'react'
import { api, DEPARTMENTS, REVIEW_GRADES, type Employee } from '../api'

// ═══ ملف الموارد البشرية ═══
//
// منقول من نظام الطاقة الشمسية: القسم وتاريخ التعيين وسنوات الخبرة
// والمستوى الوظيفي وآخر تقييم.
//
// والحالة الوظيفية (مستقر / يحتاج ترقية / يحتاج تدريب) **ما تنكتب
// بالإيد** — تنحسب بالسيرفر من الخبرة والمستوى والتقييم. بالنظام
// القديم جانت تنحسب بالمتصفح، يعني كل واحد يقدر يغيّرها من أدوات
// المطوّر، والقاعدة نفسها تنطبق أو ما تنطبق حسب أي شاشة انفتحت.
//
// القاعدة:
//   خبرة ≥٣ سنوات و مستوى ≥٧ و تقييم ممتاز/جيد جداً  →  يحتاج ترقية
//   تقييم «يحتاج تحسين» أو مستوى <٤                    →  يحتاج تدريب
//   غيرها                                              →  مستقر
export default function EmployeeHRPanel({
  employee,
  onUpdated,
}: {
  employee: Employee
  onUpdated: (e: Employee) => void
}) {
  const [saving, setSaving] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const save = async (patch: Parameters<typeof api.updateEmployee>[1], key: string) => {
    setSaving(key)
    setErr(null)
    try {
      onUpdated(await api.updateEmployee(employee.id, patch))
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setSaving(null)
    }
  }

  const status = employee.careerStatus || 'مستقر'
  const statusStyle =
    status === 'يحتاج ترقية' ? 'border-amber-400 bg-amber-50 text-amber-800'
      : status === 'يحتاج تدريب' ? 'border-red-400 bg-red-50 text-red-800'
        : status === 'تحت المراقبة' ? 'border-slate-400 bg-slate-50 text-slate-700'
          : 'border-emerald-400 bg-emerald-50 text-emerald-800'
  const statusIcon =
    status === 'يحتاج ترقية' ? '⬆️' : status === 'يحتاج تدريب' ? '📚' : status === 'تحت المراقبة' ? '👁️' : '✅'

  return (
    <div className="mt-5 rounded-2xl border-2 border-cyan-200 bg-cyan-50/40 p-4">
      <div className="mb-3 flex items-center gap-2">
        <div className="h-1.5 w-1.5 rounded-full bg-cyan-600" />
        <h4 className="text-sm font-bold text-[#0f2040]">ملف الموارد البشرية</h4>
      </div>

      <div className={`mb-3 rounded-xl border-2 p-3 ${statusStyle}`}>
        <div className="text-sm font-black">{statusIcon} {status}</div>
        <div className="mt-1 text-[11px] opacity-80">
          تنحسب تلقائياً من الخبرة والمستوى والتقييم — غيّر أي وحدة منهن وتتحدّث لحالها
        </div>
        {status === 'يحتاج ترقية' && (
          <div className="mt-2">
            <label className="mb-1 block text-[11px] font-bold">الوظيفة المقترحة للترقية</label>
            <input
              defaultValue={employee.nextRole ?? ''}
              onBlur={(e) => save({ nextRole: e.target.value }, 'nextRole')}
              placeholder="مثال: رئيس فريق التركيب"
              className="w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm"
            />
          </div>
        )}
        {status === 'يحتاج تدريب' && (
          <div className="mt-2">
            <label className="mb-1 block text-[11px] font-bold">الاحتياجات التدريبية</label>
            <textarea
              defaultValue={employee.trainingNeeds ?? ''}
              onBlur={(e) => save({ trainingNeeds: e.target.value }, 'trainingNeeds')}
              rows={2}
              placeholder="دورة سلامة كهربائية، تدريب على أحدث تقنيات الإنفيرترات..."
              className="w-full rounded-lg border border-red-300 bg-white px-3 py-2 text-sm"
            />
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">القسم</label>
          <select
            defaultValue={employee.department ?? ''}
            onChange={(e) => save({ department: e.target.value || null }, 'department')}
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
          >
            <option value="">غير محدد</option>
            {DEPARTMENTS.map((d) => <option key={d}>{d}</option>)}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">تاريخ التعيين</label>
          <input
            type="date"
            defaultValue={employee.hireDate ? employee.hireDate.slice(0, 10) : ''}
            onBlur={(e) => save({ hireDate: e.target.value }, 'hireDate')}
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">سنوات الخبرة</label>
          <input
            type="number"
            min={0}
            step={0.5}
            defaultValue={employee.experienceYears ?? ''}
            onBlur={(e) => save({ experienceYears: e.target.value === '' ? null : Number(e.target.value) }, 'exp')}
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">المستوى الوظيفي (١-١٠)</label>
          <input
            type="number"
            min={1}
            max={10}
            defaultValue={employee.jobLevel ?? 5}
            onBlur={(e) => {
              const v = Number(e.target.value)
              if (v >= 1 && v <= 10) save({ jobLevel: v }, 'level')
            }}
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">آخر تقييم أداء</label>
          <select
            defaultValue={employee.lastReview ?? ''}
            onChange={(e) => save({ lastReview: e.target.value || null }, 'review')}
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm"
          >
            <option value="">غير مقيّم</option>
            {REVIEW_GRADES.map((g) => <option key={g}>{g}</option>)}
          </select>
        </div>

        <div className="flex items-end">
          {saving && <span className="text-xs font-bold text-cyan-700">جاري الحفظ...</span>}
          {err && <span className="text-xs font-bold text-red-600">{err}</span>}
        </div>
      </div>
    </div>
  )
}
