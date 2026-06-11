import express from 'express'
import cors from 'cors'
import dotenv from 'dotenv'
import employeesRouter from './routes/employees'
import servicesRouter from './routes/services'

dotenv.config()

const app = express()
app.use(cors())
app.use(express.json())

app.get('/api/health', (_req, res) => {
  res.json({ status: 'ok' })
})

app.use('/api/employees', employeesRouter)
app.use('/api/services', servicesRouter)

const PORT = process.env.PORT || 4000
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`)
})
