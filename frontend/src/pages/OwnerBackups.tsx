import { useEffect, useState } from 'react'
import { api, type BackupOverview, type BackupRun } from '../api'
import { useSession } from '../session'

// ═══ النسخ الاحتياطية — شاشة المالك وحده ═══
//
// شرط صريح: هذا الملف ما يشوفه غير المالك. حتى مدير النظام (ADMIN) لا
// يشوف الرابط ولا الصفحة، والمسار بالباك إند يرجّعله 404 مو 403 — يعني
// ما يعرف إنها موجودة أصلاً.
//
// ⚠️ الواجهة تطبّع دور المالك إلى 'ADMIN' (شوف Layout.tsx) والدور
// الحقيقي يبقى بـactualRole. لهذا الفحص هنا لازم يكون actualRole —
// لو كتبته role === 'OWNER' ما يشوفها المالك نفسه، ولو كتبته
// role === 'ADMIN' يشوفها كل مدير.
export default function OwnerBackups() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'

  const [data, setData] = useState<BackupOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // ⚠️ غير المالك ما ننادي له أصلاً — ومو محتاجين نطفّي مؤشر التحميل
  // إله، لأن الرندر يرجّع «الصفحة غير موجودة» قبل ما يوصل لفحص
  // التحميل. تصفير الحالة هنا چان يسبّب دورة رسم زايدة بلا فايدة.
  useEffect(() => {
    if (!isOwner) return
    let alive = true
    void (async () => {
      try {
        const overview = await api.getOwnerBackups(30)
        if (alive) setData(overview)
      } catch (e) {
        if (alive) setError(e instanceof Error ? e.message : 'تعذر جلب حالة النسخ')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [isOwner])

  if (!isOwner) {
    return (
      <div className="rounded-2xl bg-white p-8 text-center shadow">
        <p className="text-lg font-bold text-slate-700">الصفحة غير موجودة</p>
      </div>
    )
  }
  if (loading) return <div className="p-8 text-center text-slate-500">جاري التحميل…</div>
  if (error) return <div className="rounded-2xl bg-red-50 p-6 text-red-700">{error}</div>
  if (!data) return null

  const h = data.health
  const tone: Record<string, string> = {
    OK: 'bg-emerald-50 border-emerald-200 text-emerald-800',
    WARN: 'bg-amber-50 border-amber-200 text-amber-800',
    CRITICAL: 'bg-red-50 border-red-200 text-red-800',
    UNKNOWN: 'bg-slate-50 border-slate-200 text-slate-700',
  }
  const icon: Record<string, string> = { OK: '✅', WARN: '⚠️', CRITICAL: '🔴', UNKNOWN: '❔' }

  const fmt = (iso: string | null) =>
    iso ? new Date(iso).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' }) : '—'
  const mb = (b: number) => (b > 0 ? `${(b / 1024 / 1024).toFixed(1)} م.ب` : '—')
  const since = (hours: number | null) => {
    if (hours == null) return '—'
    if (hours < 24) return `قبل ${Math.round(hours)} ساعة`
    return `قبل ${Math.floor(hours / 24)} يوم`
  }

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <h1 className="flex items-center gap-2 text-xl font-extrabold text-[#0f2040]">
          🔐 النسخ الاحتياطية
        </h1>
        <p className="mt-1 text-sm text-slate-500">
          هذي الشاشة تخصك وحدك — ما يشوفها أي حساب ثاني بالنظام، ولا حتى مدير النظام.
        </p>
      </div>

      {/* الحكم */}
      <div className={`rounded-2xl border p-5 ${tone[h.status]}`}>
        <div className="flex items-start gap-3">
          <span className="text-2xl">{icon[h.status]}</span>
          <div className="flex-1">
            <p className="text-lg font-extrabold">{h.message}</p>
            <p className="mt-1 text-sm opacity-80">
              آخر نسخة ناجحة: {fmt(h.lastSuccess?.startedAt ?? null)} ({since(h.hoursSinceSuccess)})
            </p>
          </div>
        </div>
        {h.recommendations.length > 0 && (
          <ul className="mt-4 space-y-1.5 text-sm">
            {h.recommendations.map((r, i) => (
              <li key={i} className="flex gap-2"><span>←</span><span>{r}</span></li>
            ))}
          </ul>
        )}
      </div>

      {/* أرقام سريعة */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <Stat label="نجحت (٧ أيام)" value={String(h.success7d)} />
        <Stat label="فشلت (٧ أيام)" value={String(h.failed7d)} tone={h.failed7d > 0 ? 'bad' : undefined} />
        <Stat label="خارج السيرفر" value={h.offsiteOK ? 'نعم' : 'لا'} tone={h.offsiteOK ? undefined : 'bad'} />
        <Stat label="مشفّرة" value={h.encrypted ? 'نعم' : 'لا'} tone={h.encrypted ? undefined : 'bad'} />
      </div>

      {/* السجل */}
      <div className="overflow-hidden rounded-2xl bg-white shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <div className="border-b border-slate-100 px-5 py-3 text-sm font-bold text-[#0f2040]">
          سجل التشغيلات ({data.runs.length})
        </div>
        {data.runs.length === 0 ? (
          <p className="p-6 text-center text-sm text-slate-500">
            ماكو تشغيلات مسجّلة. شغّل <code className="rounded bg-slate-100 px-1">./setup-backups.sh</code> على السيرفر.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-right text-sm">
              <thead className="bg-slate-50 text-xs text-slate-500">
                <tr>
                  <th className="p-3">الوقت</th>
                  <th className="p-3">الحالة</th>
                  <th className="p-3">الملف</th>
                  <th className="p-3">الحجم</th>
                  <th className="p-3">الجداول</th>
                  <th className="p-3">خارج السيرفر</th>
                  <th className="p-3">ملاحظات</th>
                </tr>
              </thead>
              <tbody>
                {data.runs.map((run: BackupRun) => (
                  <tr key={run.id} className="border-t border-slate-100 align-top">
                    <td className="whitespace-nowrap p-3 text-slate-600">{fmt(run.startedAt)}</td>
                    <td className="p-3">
                      {run.ok
                        ? <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-bold text-emerald-700">نجحت</span>
                        : <span className="rounded-full bg-red-50 px-2 py-0.5 text-xs font-bold text-red-700">فشلت</span>}
                    </td>
                    <td className="p-3 text-xs text-slate-500">
                      {run.fileName || '—'}
                      {run.encrypted && <span className="mr-1">🔒</span>}
                    </td>
                    <td className="whitespace-nowrap p-3 text-slate-600">{mb(run.sizeBytes)}</td>
                    <td className="p-3 text-slate-600">{run.tableCount || '—'}</td>
                    <td className="p-3 text-xs text-slate-600">{run.offsite ? (run.offsiteTarget || 'نعم') : '—'}</td>
                    <td className="p-3 text-xs text-slate-500">
                      {run.error && <div className="font-bold text-red-600">{run.error}</div>}
                      {run.warnings}
                      {!run.error && !run.warnings && '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* تذكير الاسترجاع — النسخة الي ما جرّبت ترجّعها ما تعرف إذا تشتغل */}
      <div className="rounded-2xl bg-white p-5 text-sm text-slate-600 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <p className="mb-2 font-bold text-[#0f2040]">وقت الحاجة (من السيرفر عبر SSH)</p>
        <pre className="overflow-x-auto rounded-xl bg-slate-900 p-3 text-left text-xs text-slate-100" dir="ltr">
{`cd ~/staff_mange
./backup-status.sh      # فحص سريع
./backup-db.sh          # نسخة يدوية الآن
./restore-backup.sh     # استرجاع من آخر نسخة`}
        </pre>
        <p className="mt-2 text-xs text-slate-500">
          الأدلة الكاملة بالمستودع: <b>EMERGENCY.md</b> وقت الكارثة، و<b>RESTORE.md</b> للتفاصيل.
          واحتفظ بـ<b>BACKUP_PASSPHRASE</b> بمكان آمن برّا السيرفر — بدونها النسخ المشفّرة ما تنفتح.
        </p>
      </div>
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: string; tone?: 'bad' }) {
  return (
    <div className="rounded-2xl bg-white p-4 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-xl font-extrabold ${tone === 'bad' ? 'text-red-600' : 'text-[#0f2040]'}`}>{value}</p>
    </div>
  )
}
