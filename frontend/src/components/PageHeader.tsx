/**
 * ترويسة الصفحة المتدرّجة — نفس شكل «تتبع المهام»
 * (MissionsPage.tsx) الي اختاره صاحب النظام.
 *
 * التدرّج غامق بالحالتين فالنص أبيض دائماً، فما نحتاج رموز هنا:
 * البطاقة نفسها هي الخلفية مو خلفية الصفحة.
 */
import { BackLink } from '../ui/BackLink'

export default function PageHeader({
  title, subtitle, aside, children, back,
}: {
  title: string
  /** (دازاي 11-15) إذا انطى مسار: يطلع «→ رجوع» فوگ العنوان، وهذا المسار إذا ماكو تاريخ. */
  back?: string
  subtitle?: string
  /** رقم بارز أو شارات على الجنب الثاني. */
  aside?: React.ReactNode
  /** أدوات تحت العنوان (أزرار مثلاً). */
  children?: React.ReactNode
}) {
  return (
    <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1a3a6c] p-6 text-white">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          {back && <BackLink to={back} className="mb-2" />}
          <h1 className="text-2xl font-bold">{title}</h1>
          {subtitle && <p className="mt-1 text-sm text-blue-200/80">{subtitle}</p>}
        </div>
        {aside}
      </div>
      {children && <div className="mt-4">{children}</div>}
    </div>
  )
}
