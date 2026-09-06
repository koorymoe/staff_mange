// ═══ لوحة الكونسول بتبويباتها ═══
//
// «CLI · Config · Interfaces · Summary» — مثل التصميم الي طلبه صاحب
// النظام.
//
// ⚠️⚠️ **الأربعة كلها مشتقّة من `cliState` و`result`** — ماكو ولا حالة
// جديدة تنخزن. لو خزّنّا نسخة للعرض، أول أمر ينكتب بالترمنال يخلّي
// التبويبات تعرض ماضياً: الفني يغيّر VLAN ويشوف القديم بتبويب
// «الواجهات»، فيظن أمره ما نفّذ ويعيده مرة ثانية.
//
// ⚠️ وتبويب «الإعداد» يستعمل **نفس** `renderShow` الي يستعملها أمر
// `show running-config` — مو مولّداً ثانياً. مولّدان يفترقان بأول
// تعديل، والمتدرّب يشوف إعدادين مختلفين لنفس الجهاز.

import { lazy, Suspense, useState } from 'react'

import { renderShow } from '../cli/engine'
import type { CliGrammar } from '../cli/grammar'
import { expandIfName } from '../cli/engine'
import { PART_BY_ID } from './catalog'
import type { LabDoc, LabNode, SimResult } from './types'

const CliTerminal = lazy(() => import('../cli/CliTerminal'))

interface Props {
  node: LabNode
  doc: LabDoc
  result: SimResult | null
  grammar: CliGrammar
  onStateChange: (st: Record<string, unknown>) => void
  onClose: () => void
}

type Tab = 'cli' | 'config' | 'ifaces' | 'summary'

const TABS: [Tab, string][] = [
  ['cli', 'CLI'],
  ['config', 'الإعداد'],
  ['ifaces', 'الواجهات'],
  ['summary', 'الملخّص'],
]

export default function ConsolePanel({ node, doc, result, grammar, onStateChange, onClose }: Props) {
  const [tab, setTab] = useState<Tab>('cli')
  const st = (node.cliState ?? {}) as Record<string, unknown>
  const host = String(st.hostname ?? node.params.hostname ?? 'Switch')
  const ifs = (st.interfaces ?? {}) as Record<string, Record<string, unknown>>
  const vlans = (st.vlans ?? {}) as Record<string, { name?: string }>
  const part = PART_BY_ID[node.partId]

  // ═══ الواجهات ═══
  //
  // ⚠️ **كل منافذ الجهاز** مو الي انهيّأت بس: منفذ ما انلمس هو منفذ
  // بـVLAN 1 وحالة «مو مربوط» — وهاي معلومة، مو فراغ. جدول يعرض
  // المهيَّأة بس يخلّي الفني يظن الباقي مو موجود.
  const rows = (part?.ports ?? []).map((p) => {
    const link = doc.links.find((l) =>
      (l.from.node === node.id && l.from.port === p.id) || (l.to.node === node.id && l.to.port === p.id))
    const peerId = link ? (link.from.node === node.id ? link.to.node : link.from.node) : null
    const peer = peerId ? doc.nodes.find((n) => n.id === peerId) : null
    const cfg = ifs[expandIfName(p.id)] ?? {}
    const state = !link ? 'down' : result?.linkState[link.id] === 'bad' ? 'err'
      : cfg.shutdown ? 'shut' : result ? 'up' : 'idle'
    return {
      id: p.id,
      label: p.label,
      kind: p.kind,
      vlan: Number(cfg.accessVlan) || 1,
      mode: String(cfg.mode ?? 'access'),
      desc: String(cfg.description ?? ''),
      peer: peer ? String(peer.cliState?.hostname ?? peer.params.name ?? peer.params.hostname ?? peer.id) : '—',
      state,
    }
  })

  const STATE_UI: Record<string, { t: string; cls: string }> = {
    up: { t: 'up', cls: 'text-emerald-300' },
    down: { t: 'not connected', cls: 'text-slate-500' },
    err: { t: 'err-disabled', cls: 'text-red-300' },
    shut: { t: 'admin down', cls: 'text-amber-300' },
    idle: { t: 'connected', cls: 'text-sky-300' },
  }

  const pre = 'h-[280px] overflow-auto rounded-xl bg-black p-3 text-left font-mono text-[12px] leading-relaxed text-slate-200 ring-1 ring-slate-800'

  return (
    <div className="border-t border-slate-800 bg-[#0b1220] p-3">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <button onClick={onClose} className="text-[11px] font-bold text-slate-500 hover:text-slate-300">✕ سكّر</button>
        <div className="flex items-center gap-2">
          <div className="inline-flex overflow-hidden rounded-lg ring-1 ring-slate-700">
            {TABS.map(([id, label]) => (
              <button key={id} onClick={() => setTab(id)}
                className={`px-3 py-1 text-[11.5px] font-bold transition ${
                  tab === id ? 'bg-slate-700 text-white' : 'bg-[#0b1220] text-slate-400 hover:text-slate-200'}`}>
                {label}
              </button>
            ))}
          </div>
          <p className="text-xs font-bold text-slate-200">🖥️ {host}</p>
        </div>
      </div>

      {tab === 'cli' && (
        <>
          <Suspense fallback={<div className="h-[280px] rounded-xl bg-black" />}>
            <CliTerminal
              key={node.id}
              grammar={grammar}
              initialState={node.cliState ?? { hostname: host }}
              onStateChange={onStateChange}
              heightClass="h-[280px]"
            />
          </Suspense>
          <p className="mt-1.5 text-[10.5px] text-slate-600">
            مثال: <span dir="ltr" className="font-mono">en → conf t → int gi0/2 → switchport access vlan 20 → end</span>
          </p>
        </>
      )}

      {tab === 'config' && (
        <pre dir="ltr" className={pre}>{renderShow('running-config', st, grammar).join('\n')}</pre>
      )}

      {tab === 'ifaces' && (
        <div className="h-[280px] overflow-auto rounded-xl bg-black p-2 ring-1 ring-slate-800">
          <table dir="ltr" className="w-full text-left font-mono text-[11.5px]">
            <thead className="text-slate-500">
              <tr className="border-b border-slate-800">
                <th className="px-2 py-1.5">Port</th>
                <th className="px-2 py-1.5">Status</th>
                <th className="px-2 py-1.5">VLAN</th>
                <th className="px-2 py-1.5">Mode</th>
                <th className="px-2 py-1.5">Neighbor</th>
                <th className="px-2 py-1.5">Description</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => {
                const ui = STATE_UI[r.state]
                return (
                  <tr key={r.id} className="border-b border-slate-800/50">
                    <td className="px-2 py-1.5 text-slate-300">{r.label}</td>
                    <td className={`px-2 py-1.5 ${ui.cls}`}>{ui.t}</td>
                    <td className="px-2 py-1.5 text-slate-300">{r.vlan}</td>
                    <td className="px-2 py-1.5 text-slate-500">{r.mode}</td>
                    <td className="px-2 py-1.5 text-slate-400" dir="rtl">{r.peer}</td>
                    <td className="px-2 py-1.5 text-slate-500">{r.desc || '—'}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          {rows.length === 0 && <p className="p-4 text-center text-[11px] text-slate-600">ماكو منافذ بهذا الجهاز.</p>}
        </div>
      )}

      {tab === 'summary' && (
        <div className="h-[280px] overflow-auto rounded-xl bg-[#0e1626] p-4 ring-1 ring-slate-800">
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {[
              ['اسم الجهاز', host],
              ['المنافذ', `${rows.filter((r) => r.state !== 'down').length}/${rows.length} مربوط`],
              ['VLANات معرَّفة', Object.keys(vlans).length ? Object.keys(vlans).join('، ') : 'الافتراضي بس'],
              ['منافذ مهيَّأة', String(Object.keys(ifs).length)],
            ].map(([k, v]) => (
              <div key={k} className="rounded-xl bg-[#0b1220] p-3 ring-1 ring-slate-800">
                <p className="text-[10.5px] text-slate-500">{k}</p>
                <p className="mt-1 text-[13px] font-bold text-slate-100">{v}</p>
              </div>
            ))}
          </div>

          {/* ⚠️ قراءات المحرّك على هذا الجهاز — نفس الي تنعرض على
              القطعة باللوح. مصدر واحد، فما تختلف بين مكانين. */}
          {(result?.nodeReadings[node.id]?.length ?? 0) > 0 && (
            <div className="mt-3 rounded-xl bg-[#0b1220] p-3 ring-1 ring-slate-800">
              <p className="mb-2 text-[10.5px] text-slate-500">قراءات المحاكاة</p>
              <div className="flex flex-wrap gap-2">
                {result!.nodeReadings[node.id].map((r, i) => (
                  <span key={i} className={`rounded-lg px-2.5 py-1 text-[11.5px] font-bold ring-1 ${
                    r.tone === 'bad' ? 'bg-red-500/10 text-red-300 ring-red-500/30'
                      : r.tone === 'warn' ? 'bg-amber-500/10 text-amber-300 ring-amber-500/30'
                        : 'bg-emerald-500/10 text-emerald-300 ring-emerald-500/30'}`}>
                    {r.text}
                  </span>
                ))}
              </div>
            </div>
          )}
          {!result && (
            <p className="mt-3 text-[11px] text-slate-600">شغّل المحاكاة حتى تظهر القراءات.</p>
          )}
        </div>
      )}
    </div>
  )
}
