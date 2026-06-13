import express from 'express'
import cors from 'cors'
import dotenv from 'dotenv'
import employeesRouter from './routes/employees'
import servicesRouter from './routes/services'
import customersRouter from './routes/customers'
import bookingsRouter from './routes/bookings'
import authRouter from './routes/auth'
import statsRouter from './routes/stats'

dotenv.config()

const app = express()
app.use(cors())
app.use(express.json())

app.get('/api/health', (_req, res) => {
  res.json({ status: 'ok' })
})

app.use('/api/employees', employeesRouter)
app.use('/api/services', servicesRouter)
app.use('/api/customers', customersRouter)
app.use('/api/bookings', bookingsRouter)
app.use('/api/auth', authRouter)
app.use('/api/stats', statsRouter)

const PORT = process.env.PORT || 4000
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`)
})
