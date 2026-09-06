import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './fonts.css'
import './index.css'
import App from './App.tsx'
import ErrorBoundary from './components/ErrorBoundary.tsx'
import { installNetworkErrorTrap } from './netErrors.ts'
import { applyTheme, prefersDark } from './utils/theme.ts'

// قبل الرندر: أي طلب يفشل بلا معالجة بالشاشة يوصل الموظف برسالة
// مفهومة بدل ما ينبلع بصمت.
installNetworkErrorTrap()

// ⚠️ الوضع الليلي ينطبّق **قبل** أول رسم: لو انتظرنا React، الموظف
// يشوف ومضة بيضا بكل فتحة صفحة — وهاي بالليل تضرب العين فعلاً.
applyTheme(prefersDark())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <BrowserRouter basename={import.meta.env.BASE_URL}>
        <App />
      </BrowserRouter>
    </ErrorBoundary>
  </StrictMode>,
)
