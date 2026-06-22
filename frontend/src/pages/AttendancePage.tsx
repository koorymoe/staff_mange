import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '../session'

/* ───── localStorage helpers ───── */

interface AttendanceRecord {
  employeeId: string
  employeeName: string
  date: string
  checkInTime: string | null
  checkOutTime: string | null
}

const todayKey = () => new Date().toISOString().slice(0, 10)

function getRecords(): AttendanceRecord[] {
  try {
    return JSON.parse(localStorage.getItem('attendance_records') || '[]')
  } catch {
    return []
  }
}

function saveRecords(records: AttendanceRecord[]) {
  localStorage.setItem('attendance_records', JSON.stringify(records))
}

function getMyRecord(employeeId: string): AttendanceRecord | undefined {
  return getRecords().find(
    (r) => r.employeeId === employeeId && r.date === todayKey(),
  )
}

function checkIn(employeeId: string, employeeName: string): AttendanceRecord {
  const records = getRecords()
  const now = new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
  const rec: AttendanceRecord = {
    employeeId,
    employeeName,
    date: todayKey(),
    checkInTime: now,
    checkOutTime: null,
  }
  records.push(rec)
  saveRecords(records)
  return rec
}

function checkOut(employeeId: string): AttendanceRecord | undefined {
  const records = getRecords()
  const rec = records.find(
    (r) => r.employeeId === employeeId && r.date === todayKey(),
  )
  if (rec) {
    rec.checkOutTime = new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
    saveRecords(records)
  }
  return rec
}

function getTodayRecords(): AttendanceRecord[] {
  return getRecords().filter((r) => r.date === todayKey())
}

/* ───── elapsed time helper ───── */

function elapsedSince(timeStr: string): string {
  // parse HH:MM from Arabic locale string — extract digits
  const digits = timeStr.replace(/[^\d:]/g, '')
  const [h, m] = digits.split(':').map(Number)
  if (isNaN(h) || isNaN(m)) return '--'
  const now = new Date()
  const checkInDate = new Date()
  checkInDate.setHours(h, m, 0, 0)
  const diffMs = now.getTime() - checkInDate.getTime()
  if (diffMs < 0) return '٠ دقائق'
  const diffMin = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMin / 60)
  const mins = diffMin % 60
  if (hours === 0) return `${mins} دقيقة`
  return `${hours} ساعة و ${mins} دقيقة`
}

/* ───── Component ───── */

export default function AttendancePage() {
  const { employee } = useSession()
  const navigate = useNavigate()

  const [record, setRecord] = useState<AttendanceRecord | undefined>(undefined)
  const [justCheckedIn, setJustCheckedIn] = useState(false)
  const [justCheckedOut, setJustCheckedOut] = useState(false)
  const [elapsed, setElapsed] = useState('')
  const [todayAll, setTodayAll] = useState<AttendanceRecord[]>([])

  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'MONITOR'

  // Load existing record
  useEffect(() => {
    if (!employee) return
    const existing = getMyRecord(employee.id)
    setRecord(existing)
    if (isAdmin) setTodayAll(getTodayRecords())
  }, [employee, isAdmin])

  // Elapsed time ticker
  useEffect(() => {
    if (!record?.checkInTime || record.checkOutTime) return
    setElapsed(elapsedSince(record.checkInTime))
    const interval = setInterval(() => {
      setElapsed(elapsedSince(record.checkInTime!))
    }, 60000)
    return () => clearInterval(interval)
  }, [record])

  const handleCheckIn = useCallback(() => {
    if (!employee) return
    const rec = checkIn(employee.id, employee.name)
    setRecord(rec)
    setJustCheckedIn(true)
    if (isAdmin) setTodayAll(getTodayRecords())
    setTimeout(() => navigate('/'), 2000)
  }, [employee, navigate, isAdmin])

  const handleCheckOut = useCallback(() => {
    if (!employee) return
    const rec = checkOut(employee.id)
    setRecord(rec)
    setJustCheckedOut(true)
    if (isAdmin) setTodayAll(getTodayRecords())
  }, [employee, isAdmin])

  if (!employee) return null

  /* ── Just checked in: success screen ── */
  if (justCheckedIn && record) {
    return (
      <div dir="rtl" className="flex min-h-[80vh] items-center justify-center p-4">
        <div className="w-full max-w-md rounded-2xl bg-white p-10 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          {/* Check icon */}
          <div className="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-full bg-green-100">
            <svg className="h-14 w-14 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="mb-2 text-2xl font-bold text-green-700">شكراً! تم تسجيل حضورك</h2>
          <p className="text-lg text-gray-600">وقت الحضور: {record.checkInTime}</p>
          <p className="mt-4 text-sm text-gray-400">جارٍ التحويل إلى لوحة التحكم...</p>
        </div>
      </div>
    )
  }

  /* ── Already checked in today ── */
  if (record && record.checkInTime) {
    return (
      <div dir="rtl" className="mx-auto max-w-3xl space-y-8 p-4">
        {/* Status card */}
        <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-6 text-white">
            <h1 className="text-2xl font-bold">حالة الحضور</h1>
            <p className="mt-1 text-sm text-blue-200">{employee.name}</p>
          </div>

          <div className="p-8 text-center">
            {!record.checkOutTime ? (
              <>
                {/* Present indicator */}
                <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-green-100">
                  <svg className="h-10 w-10 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <p className="mb-1 text-xl font-bold text-gray-800">
                  أنت متواجد منذ الساعة {record.checkInTime}
                </p>
                <p className="mb-8 text-lg text-gray-500">المدة: {elapsed}</p>

                {/* Check-out button */}
                <button
                  onClick={handleCheckOut}
                  className="inline-flex items-center gap-3 rounded-full bg-gradient-to-l from-red-600 to-amber-500 px-10 py-4 text-lg font-bold text-white shadow-lg transition hover:scale-105 hover:shadow-xl"
                >
                  <svg className="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h6a2 2 0 012 2v1" />
                  </svg>
                  انصرفت
                </button>
              </>
            ) : (
              <>
                {/* Checked out */}
                <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-amber-100">
                  <svg className="h-10 w-10 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <p className="text-xl font-bold text-gray-800">تم تسجيل الانصراف</p>
                <div className="mt-4 inline-block rounded-xl bg-gray-50 px-8 py-4">
                  <p className="text-gray-600">الحضور: <span className="font-bold text-[#2c5aad]">{record.checkInTime}</span></p>
                  <p className="text-gray-600">الانصراف: <span className="font-bold text-[#2c5aad]">{record.checkOutTime}</span></p>
                </div>
                {justCheckedOut && (
                  <p className="mt-4 text-lg font-semibold text-green-600">شكراً لك، أحسنت العمل اليوم!</p>
                )}
              </>
            )}
          </div>
        </div>

        {/* Admin table */}
        {isAdmin && <AdminTable records={todayAll} />}
      </div>
    )
  }

  /* ── Default: not checked in yet ── */
  return (
    <div dir="rtl" className="mx-auto max-w-3xl space-y-8 p-4">
      {/* Welcome card */}
      <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-10 text-center text-white">
          <h1 className="text-3xl font-bold">مرحباً، {employee.name}</h1>
          <p className="mt-2 text-blue-200">يرجى تسجيل حضورك للبدء</p>
        </div>

        <div className="flex flex-col items-center gap-6 p-10">
          {/* Big pulsing check-in button */}
          <button
            onClick={handleCheckIn}
            className="group relative flex h-40 w-40 items-center justify-center rounded-full bg-gradient-to-br from-green-500 to-emerald-600 shadow-xl transition hover:scale-105 hover:shadow-2xl"
          >
            {/* Pulse ring */}
            <span className="absolute inset-0 animate-ping rounded-full bg-green-400 opacity-20" />
            <span className="absolute inset-0 animate-pulse rounded-full bg-green-400 opacity-10" />

            {/* Icon */}
            <svg className="relative z-10 h-16 w-16 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </button>

          <span className="text-2xl font-bold text-gray-800">متواجد بالشركة</span>
          <p className="text-gray-500">اضغط لتسجيل حضورك</p>
        </div>
      </div>

      {/* Admin table */}
      {isAdmin && <AdminTable records={todayAll} />}
    </div>
  )
}

/* ───── Admin attendance table ───── */

function AdminTable({ records }: { records: AttendanceRecord[] }) {
  return (
    <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className="bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-5 text-white">
        <h2 className="text-xl font-bold">حضور الموظفين اليوم</h2>
      </div>

      {records.length === 0 ? (
        <p className="p-8 text-center text-gray-400">لا توجد سجلات حضور اليوم</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-right text-sm">
            <thead className="bg-gray-50 text-gray-600">
              <tr>
                <th className="px-6 py-3 font-semibold">الموظف</th>
                <th className="px-6 py-3 font-semibold">وقت الحضور</th>
                <th className="px-6 py-3 font-semibold">وقت الانصراف</th>
                <th className="px-6 py-3 font-semibold">الحالة</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {records.map((r, i) => {
                const status = r.checkOutTime
                  ? 'انصرف'
                  : r.checkInTime
                    ? 'متواجد'
                    : 'لم يحضر'
                const color = r.checkOutTime
                  ? 'text-amber-600 bg-amber-50'
                  : r.checkInTime
                    ? 'text-green-600 bg-green-50'
                    : 'text-red-600 bg-red-50'
                return (
                  <tr key={i} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-medium text-gray-800">{r.employeeName}</td>
                    <td className="px-6 py-4 text-gray-600">{r.checkInTime || '—'}</td>
                    <td className="px-6 py-4 text-gray-600">{r.checkOutTime || '—'}</td>
                    <td className="px-6 py-4">
                      <span className={`inline-block rounded-full px-3 py-1 text-xs font-bold ${color}`}>
                        {status}
                      </span>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
