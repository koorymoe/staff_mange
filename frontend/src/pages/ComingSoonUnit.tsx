export default function ComingSoonUnit({ title }: { title: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 text-5xl">🚧</div>
      <h2 className="text-2xl font-bold text-brand-900">{title}</h2>
      <p className="mt-2 max-w-md text-slate-500">
        هذي الوحدة مضافة بالقائمة بس محتواها لسا ما انبنى — بانتظار تحديد آلية العمل والصلاحيات المطلوبة.
      </p>
    </div>
  )
}
