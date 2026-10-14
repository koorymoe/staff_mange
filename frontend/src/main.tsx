import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './fonts.css'
import './index.css'
import App from './App.tsx'
import ErrorBoundary from './components/ErrorBoundary.tsx'
import { installNetworkErrorTrap } from './netErrors.ts'
import { applyTheme, getTheme } from './utils/theme.ts'
import { installInAppAlert } from './utils/dialog.tsx'
import { startStageSync } from './utils/offlineStages.ts'

// قبل الرندر: أي طلب يفشل بلا معالجة بالشاشة يوصل الموظف برسالة
// مفهومة بدل ما ينبلع بصمت.
installNetworkErrorTrap()
// alert() يطلع بشريط من النظام بدل نافذة المتصفح «staffmanage.cc says».
installInAppAlert()

// ⚠️ الوضع الليلي ينطبّق **قبل** أول رسم: لو انتظرنا React، الموظف
// يشوف ومضة بيضا بكل فتحة صفحة — وهاي بالليل تضرب العين فعلاً.
applyTheme(getTheme())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <BrowserRouter basename={import.meta.env.BASE_URL}>
        <App />
      </BrowserRouter>
    </ErrorBoundary>
  </StrictMode>,
)

// (ع) 10-14: النظام يفتح بلا نت (مهامي)، والضغطات المحفوظة ترسل من يرجع النت
if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => { navigator.serviceWorker.register(`${import.meta.env.BASE_URL}sw.js`).catch(() => {}) })
}
startStageSync()
