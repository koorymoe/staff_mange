import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
  /** داخل الصفحة بس (مو الشاشة كلها): القائمة والهيدر يبقون شغّالين،
   *  والانتقال لشاشة ثانية يصفّر الخطأ (Layout يمرّر key بالمسار). */
  inline?: boolean
}

interface State {
  hasError: boolean
}

// أي خطأ برمجي غير متوقع بأي صفحة كان يخلي الشاشة كلها تطلع بيضاء بدون أي تفسير
// (لأنه ماكان أكو ErrorBoundary). هسه أي خطأ ينلقط هنا ويطلع للمستخدم رسالة
// واضحة + زر "إعادة تحميل" بدل الشاشة البيضاء الصامتة.
export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info)
  }

  render() {
    if (this.state.hasError && this.props.inline) {
      return (
        <div dir="rtl" className="mx-auto mt-10 flex max-w-md flex-col items-center gap-3 rounded-2xl border border-red-200 bg-red-50 p-8 text-center">
          <span className="text-4xl">⚠️</span>
          <h2 className="text-lg font-extrabold text-red-800">صار خطأ بهاي الشاشة</h2>
          <p className="text-sm text-red-700">
            بقية النظام شغّال — تكدر تروح لأي شاشة ثانية من القائمة. وإذا تكرر، دز صورة لإدارة النظام.
          </p>
          <button onClick={() => window.location.reload()}
            className="rounded-xl bg-red-600 px-5 py-2 text-sm font-bold text-white hover:bg-red-700">
            تحديث الصفحة
          </button>
        </div>
      )
    }
    if (this.state.hasError) {
      return (
        <div dir="rtl" className="flex min-h-screen flex-col items-center justify-center gap-4 bg-[#f0f4f9] p-6 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#dc2626" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 9v4M12 17h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
            </svg>
          </div>
          <h1 className="text-lg font-extrabold text-slate-800">صار خطأ غير متوقع</h1>
          <p className="max-w-sm text-sm text-slate-500">
            حدث خطأ بهذي الصفحة. جرب تحديث الصفحة، وإذا تكررت المشكلة راجع إدارة النظام.
          </p>
          <button
            onClick={() => window.location.reload()}
            className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2.5 font-medium text-white shadow-md"
          >
            تحديث الصفحة
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
