import type { CSSProperties, ReactNode } from 'react'
import type { ServiceItem, ServiceKind } from '../utils/loginServices'

function Artwork({ kind }: { kind: ServiceKind }) {
  const common = {
    fill: 'none',
    stroke: 'currentColor',
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
  }

  const art: Record<ServiceKind, ReactNode> = {
    camera: (
      <>
        <path {...common} strokeWidth="5" d="M35 34l72-15 25 14-10 35-76 14-18-13z" />
        <path {...common} strokeWidth="4" d="M45 82l-8 12m65-20 11 13M24 69l-12 4" />
        <circle {...common} strokeWidth="6" cx="104" cy="47" r="17" />
        <circle fill="currentColor" cx="104" cy="47" r="6" opacity=".9" />
        <path {...common} strokeWidth="3" d="M27 48l-15-4m19-7-13-9" opacity=".65" />
      </>
    ),
    network: (
      <>
        <path {...common} strokeWidth="5" d="M25 53h108l7 30H18z" />
        <path {...common} strokeWidth="4" d="M35 53V19m88 34V18" />
        <path {...common} strokeWidth="3" d="M27 28c6-7 11-10 18-12m70 0c8 3 13 7 18 13" opacity=".75" />
        <circle fill="currentColor" cx="43" cy="69" r="3" />
        <circle fill="currentColor" cx="55" cy="69" r="3" opacity=".65" />
        <path {...common} strokeWidth="3" d="M82 73c7-8 15-8 22 0m-17-6c4-4 8-4 12 0" />
      </>
    ),
    solar: (
      <>
        <path {...common} strokeWidth="4" d="M23 37h83l13 45H10z" />
        <path {...common} strokeWidth="2.5" d="M36 37l-8 45m36-45-3 45m30-45 5 45M17 52h94M13 67h103" opacity=".8" />
        <path {...common} strokeWidth="4" d="M52 82l-7 13m47-13 9 13M37 95h74" />
        <circle {...common} strokeWidth="3" cx="130" cy="25" r="10" />
        <path {...common} strokeWidth="2" d="M130 7v7m0 22v7m-18-18h7m22 0h7" />
      </>
    ),
    home: (
      <>
        <path {...common} strokeWidth="5" d="M20 51L75 13l55 38v43H20z" />
        <path {...common} strokeWidth="4" d="M60 94V65h30v29" />
        <path {...common} strokeWidth="3" d="M57 49c11-11 25-11 36 0M65 57c6-6 14-6 20 0" />
        <circle fill="currentColor" cx="75" cy="64" r="3.5" />
        <path {...common} strokeWidth="3" d="M32 59h16v15H32zm70 0h16v15h-16z" opacity=".75" />
      </>
    ),
    lock: (
      <>
        <rect {...common} strokeWidth="5" x="43" y="33" width="64" height="62" rx="9" />
        <path {...common} strokeWidth="6" d="M56 33V22c0-22 38-22 38 0v11" />
        <path {...common} strokeWidth="3" d="M75 50c-16 0-17 27 0 31m0-24c-8 0-8 16 0 16m0-9v20" />
        <path {...common} strokeWidth="2" d="M116 42c9 7 13 15 13 25m-5-34c14 10 21 22 21 36" opacity=".6" />
      </>
    ),
    audio: (
      <>
        <rect {...common} strokeWidth="5" x="26" y="16" width="40" height="78" rx="5" />
        <rect {...common} strokeWidth="5" x="87" y="8" width="48" height="86" rx="5" />
        <circle {...common} strokeWidth="4" cx="46" cy="66" r="14" />
        <circle {...common} strokeWidth="4" cx="111" cy="59" r="18" />
        <circle fill="currentColor" cx="46" cy="31" r="5" />
        <circle fill="currentColor" cx="111" cy="25" r="6" />
        <path {...common} strokeWidth="2" d="M11 45v24m6-33v42m130-35v28m-6-37v47" opacity=".6" />
      </>
    ),
    fire: (
      <>
        <path {...common} strokeWidth="5" d="M82 27h36v67H72V38z" />
        <path {...common} strokeWidth="4" d="M84 27V14h21v13m13 8 16-11m-5-7 9 13" />
        <circle {...common} strokeWidth="5" cx="40" cy="48" r="26" />
        <circle {...common} strokeWidth="3" cx="40" cy="48" r="13" />
        <path fill="currentColor" d="M93 78c-13-10-1-18 3-26 4 8 15 15 5 26-3 4-6 4-8 0z" opacity=".9" />
      </>
    ),
    gps: (
      <>
        <path {...common} strokeWidth="5" d="M19 70l12-23h74l20 23 13 4v13H12V75z" />
        <circle {...common} strokeWidth="5" cx="39" cy="86" r="11" />
        <circle {...common} strokeWidth="5" cx="111" cy="86" r="11" />
        <path {...common} strokeWidth="3" d="M44 47l12-17h35l14 17M59 47V30" />
        <path {...common} strokeWidth="4" d="M119 13c0 13-16 28-16 28S87 26 87 13c0-19 32-19 32 0z" />
        <circle fill="currentColor" cx="103" cy="13" r="5" />
      </>
    ),
    technical: (
      <>
        <rect {...common} strokeWidth="5" x="18" y="14" width="91" height="58" rx="5" />
        <path {...common} strokeWidth="4" d="M8 84h112l-11-12H18zM35 57l12-13 12 7 16-22 15 9" />
        <path {...common} strokeWidth="3" d="M124 50l9-9m-3 18 12-1m-20 10 7 9" />
        <circle {...common} strokeWidth="5" cx="124" cy="61" r="13" />
        <circle fill="currentColor" cx="124" cy="61" r="4" />
      </>
    ),
  }

  return (
    <svg className="login-service-art" viewBox="0 0 150 105" aria-hidden="true">
      {art[kind]}
    </svg>
  )
}

export default function ServiceCard({ item, side }: { item: ServiceItem; side: 'right' | 'left' }) {
  return (
    <article
      className={`login-service-card ${side === 'right' ? 'login-service-right' : 'login-service-left'}`}
      style={{ '--service-color': item.color } as CSSProperties}
    >
      <div className="login-service-copy">
        <h2>{item.title}</h2>
        <p>{item.desc}</p>
      </div>
      <div className="login-service-visual">
        <span className="login-service-halo" />
        <span className="login-service-orbit orbit-one" />
        <span className="login-service-orbit orbit-two" />
        <span className="login-service-scan" />
        <Artwork kind={item.kind} />
      </div>
      <span className="login-service-line"><i /></span>
    </article>
  )
}
