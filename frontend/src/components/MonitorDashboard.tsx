import { useSearchParams } from 'react-router-dom'
import { useSession } from '../session'
import MonitorHome from './MonitorHome'
import ActionsTable from './ActionsTable'
import { FollowBoard } from './AdminHome'
import MissionsPage from '../pages/MissionsPage'

// رئيسية المراقب — (ع) 10-16: «المتابعة» و«الإجراءات» نفس مكتب المدير، بس بلا حسابات
// الشركة (حسابات التدقيق بس)، وويّاهن «تتبع المهام». كل نداء ينقاس بصلاحيات الحساب
// نفسه — مراقب جديد بنفس الصلاحيات يشوف نفس الشي، وما ينادي شي يرجعله 403.
type Tab = 'home' | 'follow' | 'actions' | 'missions'

export default function MonitorDashboard() {
  const { employee, permissions } = useSession()
  const [params, setParams] = useSearchParams()
  const raw = params.get('board')
  const tab: Tab = raw === 'follow' || raw === 'actions' || raw === 'missions' ? raw : 'home'
  const gate = (perm: string) =>
    permissions.includes(perm) || (perm === 'booking_delete_approve' && (employee?.role === 'MONITOR' || permissions.includes('monitoring')))
  const btn = (t: Tab, label: string) => (
    <button key={t} type="button" onClick={() => setParams(t === 'home' ? {} : { board: t })}
      className={`rounded-xl px-4 py-2 text-sm font-extrabold transition-colors ${tab === t ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
      {label}
    </button>
  )
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {btn('home', '🏠 مكتبي')}
        {btn('follow', '📊 المتابعة')}
        {btn('actions', '⚡ الإجراءات')}
        {btn('missions', '🎯 تتبع المهام')}
      </div>
      {tab === 'home' ? <MonitorHome /> : tab === 'follow' ? <FollowBoard monitor /> : tab === 'actions' ? <ActionsTable gate={gate} /> : <MissionsPage />}
    </div>
  )
}
