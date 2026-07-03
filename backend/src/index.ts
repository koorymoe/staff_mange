import express from 'express'
import cors from 'cors'
import helmet from 'helmet'
import rateLimit from 'express-rate-limit'
import dotenv from 'dotenv'
import employeesRouter from './routes/employees'
import servicesRouter from './routes/services'
import customersRouter from './routes/customers'
import bookingsRouter from './routes/bookings'
import authRouter from './routes/auth'
import statsRouter from './routes/stats'
import expensesRouter from './routes/expenses'
import permissionsRouter from './routes/permissions'
import kpiRouter from './routes/kpi'
import cartRouter from './routes/cart'
import inventoryRouter from './routes/inventory'
import complaintsRouter from './routes/complaints'
import gpsRouter from './routes/gps'
import quotationsRouter from './routes/quotations'
import productsRouter from './routes/products'
import suppliersRouter from './routes/suppliers'
import projectsRouter from './routes/projects'
import missionsRouter from './routes/missions'
import procurementRouter from './routes/procurement'
import smartKpiRouter from './routes/smart-kpi'
import trainingRouter from './routes/training'
import { requireAuth } from './middleware/requireAuth'

dotenv.config()

const app = express()

// رؤوس أمان قياسية (Helmet) — تمنع هجمات clickjacking / MIME-sniffing / إلخ
app.use(helmet())

// CORS مقيّد بالأصل المسموح فقط (الفرونت إند المحلي أو دومين الإنتاج)، مو مفتوح لأي موقع
const allowedOrigins = (process.env.CORS_ORIGIN || 'http://localhost:5173')
  .split(',')
  .map(o => o.trim())
  .filter(Boolean)
app.use(cors({
  origin: allowedOrigins,
  credentials: true,
}))

app.use(express.json({ limit: '10mb' }))

app.get('/api/health', (_req, res) => {
  res.json({ status: 'ok' })
})

// حماية من هجمات تخمين كلمة المرور: حد أقصى 10 محاولات كل 15 دقيقة لكل IP على مسار الدخول فقط
const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  limit: 10,
  standardHeaders: true,
  legacyHeaders: false,
  message: { error: 'محاولات دخول كثيرة جداً، حاول مرة أخرى بعد قليل' },
})
app.use('/api/auth/login', loginLimiter)

// تسجيل الدخول متاح بدون توثيق مسبق (هو نفسه المصدر الوحيد للتوكن)
app.use('/api/auth', authRouter)

// كل شي بعد هذا السطر يتطلب توكن دخول صالح
app.use('/api', requireAuth)

app.use('/api/employees', employeesRouter)
app.use('/api/services', servicesRouter)
app.use('/api/customers', customersRouter)
app.use('/api/bookings', bookingsRouter)
app.use('/api/stats', statsRouter)
app.use('/api/expenses', expensesRouter)
app.use('/api/permissions', permissionsRouter)
app.use('/api/kpi', kpiRouter)
app.use('/api/cart', cartRouter)
app.use('/api/inventory', inventoryRouter)
app.use('/api/complaints', complaintsRouter)
app.use('/api/gps', gpsRouter)
app.use('/api/quotations', quotationsRouter)
app.use('/api/products', productsRouter)
app.use('/api/suppliers', suppliersRouter)
app.use('/api/projects', projectsRouter)
app.use('/api/missions', missionsRouter)
app.use('/api/procurement', procurementRouter)
app.use('/api/smart-kpi', smartKpiRouter)
app.use('/api/training', trainingRouter)

const PORT = process.env.PORT || 4000
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`)
})
