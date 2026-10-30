import {
  orgAssignmentPermissions,
  orgPositions,
  type OrgDivision,
  type OrgPosition,
} from '../org'
import type { AccessGroup } from '../permissionAccess'

export default function OrgAssignmentPreview({
  division,
  unitId,
  position,
  catalog,
}: {
  division?: OrgDivision
  unitId: string
  position: OrgPosition
  catalog: AccessGroup[] | null
}) {
  const names = orgAssignmentPermissions(division, unitId, position)
  const unit = division?.units?.find((u) => u.id === unitId)
  const labels = new Map(
    (catalog ?? []).flatMap((g) =>
      g.perms.map((p) => [p.name, p.label] as const),
    ),
  )
  return (
    <section
      className="org-wide org-grant-preview"
      aria-label="الصلاحيات الي راح ياخذها"
    >
      <header>
        <h3>🔑 الصلاحيات الي راح ياخذها</h3>
        <p role="status" aria-live="polite">
          {names !== null
            ? `${names.length.toLocaleString('ar-IQ')} صلاحية من هالمنصب`
            : 'المعاينة بعد مو متوفرة'}
        </p>
      </header>
      {!division ? (
        <p>اختار الإدارة والوحدة والمنصب حتى تشوف صلاحيات التعيين قبل الحفظ.</p>
      ) : (
        <>
          <p className="org-grant-place">
            {division.name}
            {unit && ` / ${unit.name}`} · {orgPositions[position]}
          </p>
          <p>
            {position === 'MEMBER'
              ? 'العضو ياخذ صلاحيات العمل اليومية، بدون الصلاحيات المخصصة للرئيس والنائب.'
              : unit
                ? 'رئيس الوحدة ونائبه ياخذون حزمة الوحدة كاملة.'
                : 'مدير الإدارة ونائبه ياخذون حزمة الإدارة وحزم كل وحداتها، بما بيها صلاحيات الرؤساء.'}
          </p>
          {names === null ? (
            <p role="alert" className="org-warning">
              بيانات الحزمة مو مكتملة؛ ما نكدر نعرض معاينة مؤكدة. أعد تحميل
              الهيكل.
            </p>
          ) : names.length === 0 ? (
            <p className="org-grant-empty">
              هالمنصب بهالمكان ما يضيف أي صلاحية.
            </p>
          ) : (
            <>
              {catalog === null && (
                <p role="alert">
                  أسماء الصلاحيات العربية مو متوفرة؛ نعرض الرموز الأصلية بدل
                  التخمين.
                </p>
              )}
              <ul
                className="org-permissions org-grant-list"
                aria-label="قائمة صلاحيات المنصب"
              >
                {names.map((name) => (
                  <li key={name}>
                    {labels.get(name) || (
                      <>
                        <span>صلاحية مو معرّفة بالقاموس: </span>
                        <bdi>{name}</bdi>
                      </>
                    )}
                  </li>
                ))}
              </ul>
            </>
          )}
        </>
      )}
      <p className="org-grant-note">
        هاي صلاحيات هالتعيين بس، مو كل صلاحيات الموظف. أي صلاحية عنده يدوياً
        تبقى وياه، ودوره القديم ما يتغيّر.
      </p>
    </section>
  )
}
