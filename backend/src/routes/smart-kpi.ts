import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

function getMonthRange(month?: string) {
  const now = new Date()
  const [year, mon] = month
    ? month.split('-').map(Number)
    : [now.getFullYear(), now.getMonth() + 1]
  const start = new Date(year, mon - 1, 1)
  const end = new Date(year, mon, 1)
  return { start, end, year, mon }
}

function getWorkingDays(year: number, month: number): number {
  // month is 1-based
  const daysInMonth = new Date(year, month, 0).getDate()
  let workDays = 0
  for (let d = 1; d <= daysInMonth; d++) {
    const day = new Date(year, month - 1, d).getDay()
    if (day !== 5) workDays++ // Friday off
  }
  return workDays
}

async function calculateTechnicianKpi(employeeId: string, month?: string) {
  const { start, end, year, mon } = getMonthRange(month)

  const employee = await prisma.employee.findUnique({
    where: { id: employeeId },
    select: { id: true, name: true, role: true },
  })
  if (!employee) throw new Error('Employee not found')
  if (employee.role !== 'TECHNICIAN') throw new Error('Employee is not a technician')

  // 1. Completed bookings via assignments
  const assignments = await prisma.bookingAssignment.findMany({
    where: {
      employeeId,
      booking: {
        status: 'COMPLETED',
        completedAt: { gte: start, lt: end },
      },
    },
    include: {
      booking: {
        select: { id: true, createdAt: true, completedAt: true },
      },
    },
  })

  const completedCount = assignments.length
  const completedBookingsPoints = completedCount * 10

  // 2. Completion speed
  let totalMinutes = 0
  let speedEntries = 0
  // Try missions first for more accurate timing
  const missions = await prisma.mission.findMany({
    where: {
      leaderId: employeeId,
      stage: 'COMPLETED',
      completedAt: { gte: start, lt: end },
    },
    select: { assignedAt: true, completedAt: true },
  })

  if (missions.length > 0) {
    for (const m of missions) {
      if (m.completedAt) {
        totalMinutes += (m.completedAt.getTime() - m.assignedAt.getTime()) / 60000
        speedEntries++
      }
    }
  } else {
    // Fallback to booking times
    for (const a of assignments) {
      if (a.booking.completedAt) {
        totalMinutes += (new Date(a.booking.completedAt).getTime() - new Date(a.booking.createdAt).getTime()) / 60000
        speedEntries++
      }
    }
  }

  const avgMinutes = speedEntries > 0 ? Math.round(totalMinutes / speedEntries) : 0
  // Speed bonus: under 120 min avg = 20pts, under 240 = 10pts, else 5pts
  let speedPoints = 0
  if (speedEntries > 0) {
    if (avgMinutes <= 120) speedPoints = 20
    else if (avgMinutes <= 240) speedPoints = 10
    else speedPoints = 5
  }

  // 3. Work reports
  const workReports = await prisma.workReport.findMany({
    where: {
      employeeId,
      createdAt: { gte: start, lt: end },
    },
    select: { cleanedSite: true, gaveInfo: true, tookPhotos: true },
  })

  const fullReports = workReports.filter(
    (r) => r.cleanedSite && r.gaveInfo && r.tookPhotos,
  ).length
  const partialReports = workReports.length - fullReports
  const workReportPoints = fullReports * 10 + partialReports * 5

  // 4. Attendance
  const attendanceRecords = await prisma.attendance.findMany({
    where: {
      employeeId,
      date: { gte: start, lt: end },
    },
  })

  const daysPresent = attendanceRecords.length
  const totalDays = getWorkingDays(year, mon)
  const attendanceRatio = totalDays > 0 ? daysPresent / totalDays : 0
  const attendancePoints = attendanceRatio >= 1 ? 20 : Math.round(attendanceRatio * 20)

  // 5. Complaints
  const complaintsCount = await prisma.complaint.count({
    where: {
      assignedToEmployeeId: employeeId,
      createdAt: { gte: start, lt: end },
    },
  })

  const complaintsPoints = complaintsCount === 0 ? 15 : -(complaintsCount * 10)

  // 6. Manual KPI deductions
  const manualDeductions = await prisma.kpiEvaluation.findMany({
    where: {
      employeeId,
      createdAt: { gte: start, lt: end },
    },
  })

  const manualDeductionPoints = manualDeductions.reduce((sum, d) => sum + d.points, 0)

  const totalPoints =
    completedBookingsPoints +
    speedPoints +
    workReportPoints +
    attendancePoints +
    complaintsPoints -
    manualDeductionPoints

  return {
    employeeId: employee.id,
    employeeName: employee.name,
    period: `${year}-${String(mon).padStart(2, '0')}`,
    breakdown: {
      completedBookings: { count: completedCount, points: completedBookingsPoints },
      completionSpeed: { avgMinutes, points: speedPoints },
      workReports: {
        count: workReports.length,
        fullReports,
        points: workReportPoints,
      },
      attendance: { daysPresent, totalDays, points: attendancePoints },
      complaints: { count: complaintsCount, points: complaintsPoints },
      manualDeductions: { count: manualDeductions.length, points: manualDeductionPoints },
    },
    totalPoints,
  }
}

// GET /technician/:employeeId?month=YYYY-MM
router.get('/technician/:employeeId', async (req, res) => {
  try {
    const result = await calculateTechnicianKpi(
      req.params.employeeId,
      req.query.month as string | undefined,
    )
    res.json(result)
  } catch (err: any) {
    res.status(400).json({ error: err.message })
  }
})

// GET /leaderboard?month=YYYY-MM
router.get('/leaderboard', async (req, res) => {
  try {
    const technicians = await prisma.employee.findMany({
      where: { role: 'TECHNICIAN', status: 'ACTIVE' },
      select: { id: true },
    })

    const month = req.query.month as string | undefined
    const results = await Promise.all(
      technicians.map((t) =>
        calculateTechnicianKpi(t.id, month).catch(() => null),
      ),
    )

    const leaderboard = results
      .filter((r): r is NonNullable<typeof r> => r !== null)
      .sort((a, b) => b.totalPoints - a.totalPoints)

    res.json(leaderboard)
  } catch (err: any) {
    res.status(400).json({ error: err.message })
  }
})

export default router
