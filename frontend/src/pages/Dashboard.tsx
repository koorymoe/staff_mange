export default function Dashboard() {
  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">لوحة التحكم</h2>
      <p className="mt-2 text-slate-500">
        مرحباً بك في نظام إدارة شركة الأماني المتكامل. اختر وحدة من القائمة الجانبية للبدء.
      </p>

      <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h3 className="font-bold text-brand-800">إدارة الكوادر</h3>
          <p className="mt-1 text-sm text-slate-500">
            إدارة الموظفين، الشهادات، والمهارات الفنية لكل خدمة.
          </p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h3 className="font-bold text-brand-800">الزبائن</h3>
          <p className="mt-1 text-sm text-slate-500">
            قاعدة بيانات الزبائن مع كود ثابت لكل زبون.
          </p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h3 className="font-bold text-brand-800">الخدمات</h3>
          <p className="mt-1 text-sm text-slate-500">
            قائمة الخدمات التي تقدمها الشركة وربطها بمهارات الموظفين.
          </p>
        </div>
      </div>
    </div>
  )
}
