import MyExtraTasks from '../components/MyExtraTasks'

// ═══ مهامي الإضافية — لكل موظف ═══
//
// المهام الإضافية كانت تنعرض جوّا شاشة «مهامي»، وهاي محصورة بالفنيين
// والتقنيين. يعني المدير يوجّه مهمة لإداري أو محاسب أو موظف مبيعات
// — **وما توصله** إلا كإشعار يضيع بين باقي الإشعارات.
//
// المهمة الي ما إلها مكان ثابت تنعرض بيه، تنتنسى. والمدير يظن إنه
// وجّه شغل، والموظف ما يدري إن عليه شغل.
//
// فصارت شاشة مستقلة يشوفها **كل** موظف — بلا صلاحية، لأن هاي مهامه
// هو مو مهام غيره. أما **توجيه** المهام لغيره فيحتاج صلاحية
// `extra_tasks_assign`.
export default function MyExtraTasksPage() {
  return (
    <div dir="rtl" className="mx-auto max-w-4xl space-y-4 p-1">
      <div className="flex items-center gap-3">
        <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-violet-50 text-xl">📋</span>
        <div>
          <h1 className="text-2xl font-black text-[#0f2040]">مهامي الإضافية</h1>
          <p className="text-xs text-slate-500">الشغل الي وجّهه إلك المدير خارج الحجوزات</p>
        </div>
      </div>

      <MyExtraTasks />
    </div>
  )
}
