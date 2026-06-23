import express from 'express'
import cors from 'cors'
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

dotenv.config()

const app = express()
app.use(cors())
app.use(express.json({ limit: '10mb' }))

app.get('/api/health', (_req, res) => {
  res.json({ status: 'ok' })
})

app.use('/api/employees', employeesRouter)
app.use('/api/services', servicesRouter)
app.use('/api/customers', customersRouter)
app.use('/api/bookings', bookingsRouter)
app.use('/api/auth', authRouter)
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

const PORT = process.env.PORT || 4000
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`)
})
