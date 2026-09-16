// ═══ محرّك الكاميرات والمراقبة ═══
//
// أكثر ثلاثة بلاغات تجينا بمشاريع المراقبة، وكلها **حساب مو تركيب**:
//
// ١) «التسجيل يمسح نفسه قبل الوقت» — القرص ما يكفي للأيام المطلوبة،
//    والرقم انحسب على الورق بالغلط أو ما انحسب أصلاً.
// ٢) «الكاميرات تفصل بالليل وترجع بالنهار» — الأشعة تحت الحمراء
//    تشتغل بالظلام فيقفز سحب PoE فوگ ميزانية السويچ.
// ٣) «البث متقطّع لمن نفتح كل الكاميرات» — النطاق فوگ سعة الرابط أو
//    فوگ إدخال الـNVR.
//
// ⚠️ والثلاثة **ما تنكشف بالفحص وقت التسليم**: التسليم يصير نهاراً،
// وبكاميرا أو اثنتين مفتوحتين، وبقرص فاضي. ولهذا المحاكي هنا يحسب
// **أسوأ حالة** مو حالة التسليم.
//
// ⚠️ درجة الدقة `F1`: أرقام بحالة مستقرة بمعدلات منشورة. ماكو
// تحليل مشهد ولا ضغط متغيّر لحظي. تكفي لتصميم منظومة وتشخيص
// الأعطال الثلاثة أعلاه، وما تكفي لضبط كودك بعينه.
//
// ⚠️⚠️ الأرقام أدناه **معدلات قياسية منشورة** لـH.264/H.265 — مو من
// كتالوگ موديل بعينه. المحتوى يبقى `verified = FALSE` لحد ما يجرّبه
// فني على منظومة حقيقية.

import { CABLE_BY_ID } from '../cables'
import { netFacts } from './network'
import type { DomainEngine, LabDoc, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))

/** ميغابكسل لكل دقة — الأساس الي ينبني عليه البتريت. */
const RES_MP: Record<string, number> = {
  '2mp': 2.07,   // 1080p
  '4mp': 4.0,
  '5mp': 5.0,
  '8mp': 8.29,   // 4K
}

/**
 * ═══ البتريت التقريبي بالميغابت/ثانية ═══
 *
 * ⚠️ **ينحسب مو ينخزن**: البتريت يعتمد على الدقة والإطارات والكودك
 * سوا. جدول ثابت «٤ ميغابت لكل كاميرا» يخلّي المتدرّب يصمّم بمعدل
 * وسطي — وبالميدان ما اكو معدل وسطي، اكو منظومة تشتغل أو ما تشتغل.
 *
 * المعامل ~٠٫١٠ ميغابت لكل ميغابكسل بـ١٥ إطاراً على H.264، والـH.265
 * ينزّل تقريباً للنصف — وهاي أعراف منشورة عامة.
 */
export function bitrateMbps(res: string, fps: number, codec: string): number {
  const mp = RES_MP[res] ?? 2.07
  const base = mp * 0.10 * (fps / 15)
  return codec === 'h265' ? base * 0.55 : base
}

/** سحب PoE الفعلي — **بالليل** لأن هاي أسوأ حالة. */
export function drawW(cam: { params: Record<string, unknown> }): number {
  const base = num(cam.params.poeW, 8)
  const ir = num(cam.params.irW, 4)
  return base + ir
}

export const cctvEngine: DomainEngine = {
  id: 'cctv',
  name: 'محرّك الكاميرات والمراقبة',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }

    // ⚠️ فحص الكيابل والوصولية **نفسه** الي بمحرّك الشبكة — مو نسخة.
    // منظومة كاميرات هي شبكة قبل كل شي، وأي اختلاف بالحكم بين
    // المحرّكين يعني المتدرّب يتعلّم قاعدتين متناقضتين لنفس الكيبل.
    const F = netFacts(doc)
    for (const l of doc.links) {
      const c = F.linkChecks.get(l.id)!
      linkState[l.id] = c.ok ? 'ok' : 'bad'
      if (c.forced) {
        messages.push({ kind: 'error', text: `كيبل ${c.aName} ⇄ ${c.bName}: الرابط مقطوع — ماكو إشارة.` })
      } else if (!c.ok) {
        for (const pr of c.problems) messages.push({ kind: 'error', text: `كيبل ${c.aName} ⇄ ${c.bName}: ${pr}` })
      }
    }

    const cams = doc.nodes.filter((n) => n.partId === 'ip_camera')
    const nvrs = doc.nodes.filter((n) => n.partId === 'nvr')

    if (cams.length === 0) {
      messages.push({ kind: 'warn', text: 'ماكو كاميرات — ضيف كاميرا على الأقل.' })
      return { ok: false, messages, nodeReadings, linkState }
    }

    // ═══ ١) البتريت لكل كاميرا ═══
    let totalMbps = 0
    let totalW = 0
    for (const c of cams) {
      const res = str(c.params.res, '2mp')
      const fps = num(c.params.fps, 15)
      const codec = str(c.params.codec, 'h264')
      const mb = bitrateMbps(res, fps, codec)
      totalMbps += mb
      totalW += drawW(c)
      add(c.id, `${mb.toFixed(1)} Mbps`)
      add(c.id, `${drawW(c).toFixed(1)} W ليلاً`)
      if (fps > 25) {
        messages.push({
          kind: 'warn',
          text: `«${str(c.params.name, c.id)}» على ${fps} إطاراً. المراقبة ما تحتاج أكثر من ١٥ إطاراً عادةً — الزيادة تاكل قرصاً ونطاقاً بلا فايدة بالتعرّف على الوجوه.`,
        })
      }
    }
    const anyNvr = nvrs[0]

    // ═══ ٢) حدود الـNVR ═══
    let bad = false
    if (!anyNvr) {
      messages.push({ kind: 'warn', text: 'ماكو NVR — ضيف مسجّلاً حتى نحسب التخزين والقنوات.' })
      bad = true
    } else {
      for (const nv of nvrs) {
        const ch = num(nv.params.channels, 8)
        const inMbps = num(nv.params.maxInMbps, 80)
        const poeBudget = num(nv.params.poeBudget, 0)
        const poePorts = num(nv.params.poePorts, 0)

        add(nv.id, `${cams.length}/${ch} قناة`, cams.length > ch ? 'bad' : 'ok')
        add(nv.id, `${totalMbps.toFixed(1)}/${inMbps} Mbps`, totalMbps > inMbps ? 'bad' : totalMbps > inMbps * 0.85 ? 'warn' : 'ok')

        if (cams.length > ch) {
          bad = true
          messages.push({
            kind: 'error',
            text: `🔴 ${cams.length} كاميرا على مسجّل ${ch} قنوات. الزيادة **ما تنسجّل أصلاً** — والمنظومة تبدو شغّالة لأن الكاميرات الأولى تشتغل عادي.`,
          })
        }
        if (totalMbps > inMbps) {
          bad = true
          messages.push({
            kind: 'error',
            text: `🔴 مجموع البث ${totalMbps.toFixed(1)} ميغابت وحد إدخال المسجّل ${inMbps}. التسجيل يصير متقطّعاً أو يطيح إطارات — والفجوة بالتسجيل تنكشف يوم تحتاجها.`,
          })
        }

        // ⚠️ PoE من المسجّل نفسه: أغلب مسجّلاتنا فيها منافذ PoE مدمجة،
        // وميزانيتها **أصغر** من ميزانية سويچ — وهذا الي يفاجئ الفني.
        if (poePorts > 0) {
          add(nv.id, `PoE ${totalW.toFixed(0)}/${poeBudget} W`, totalW > poeBudget ? 'bad' : 'ok')
          if (cams.length > poePorts) {
            bad = true
            messages.push({
              kind: 'error',
              text: `🔴 ${cams.length} كاميرا و${poePorts} منفذ PoE بالمسجّل. الباقيات تحتاج سويچاً PoE مستقلاً.`,
            })
          }
          if (totalW > poeBudget) {
            bad = true
            messages.push({
              kind: 'error',
              text: `🔴 سحب ${totalW.toFixed(0)} واط **بالليل** وميزانية المسجّل ${poeBudget} واط. ⚠️ نهاراً يشتغل كلشي — وبالظلام تشتغل الأشعة تحت الحمراء فيقفز السحب والكاميرات تفصل وترجع بالتناوب. وهذا العطل يضيّع أياماً لأن كل فحص نهاري يطلع سليماً.`,
            })
          }
        }

        // ═══ ٣) أيام التخزين ═══
        //
        // ⚠️ الحساب بأسوأ حالة (تسجيل مستمر ٢٤ ساعة). تسجيل بالحركة
        // يطوّل المدة، بس **ما ينبني عليه تصميم**: أول يوم فيه حركة
        // مستمرة (سوق، مخزن بالنهار) يرجّعك للرقم الأسوأ.
        const tb = num(nv.params.diskTb, 4)
        const days = num(nv.params.retentionDays, 30)
        // ميغابت/ثانية ← تيرابايت/يوم: ÷٨ للبايت، ×٨٦٤٠٠ ثانية، ÷١٠^٦
        const tbPerDay = (totalMbps / 8) * 86400 / 1e6
        const needTb = tbPerDay * days
        const actualDays = tbPerDay > 0 ? tb / tbPerDay : 0
        add(nv.id, `${actualDays.toFixed(1)} يوم تخزين`, actualDays < days ? 'bad' : 'ok')
        add(nv.id, `${tb} TB`)

        if (actualDays < days) {
          bad = true
          messages.push({
            kind: 'error',
            text: `🔴 المطلوب ${days} يوم تسجيل، والقرص ${tb} تيرابايت يكفي **${actualDays.toFixed(1)} يوم** بس. تحتاج ${needTb.toFixed(1)} تيرابايت. المسجّل راح يمسح الأقدم تلقائياً — والزبون يكتشف هذا يوم يطلب تسجيلاً قديماً، وساعتها ما اكو حل.`,
          })
        } else if (actualDays < days * 1.15) {
          messages.push({
            kind: 'warn',
            text: `⚠️ التخزين ${actualDays.toFixed(1)} يوم والمطلوب ${days} — بلا هامش. أي كاميرا تنضاف بعدين تنزّله تحت المطلوب.`,
          })
        }
      }
    }

    // ═══ ٤) النطاق على الرابط الفعلي ═══
    //
    // ⚠️ **أضعف وصلة بالمسار** هي الي تحكم، مو أسرع وحدة. رابط ١٠٠
    // ميغابت واحد بوسط المسار يخنق منظومة كلها بغيغابت.
    for (const l of doc.links) {
      const c = F.linkChecks.get(l.id)
      if (!c?.ok) continue
      const cap = Math.min(c.mbps, CABLE_BY_ID[c.cable]?.maxMbps ?? c.mbps)
      // نحسب البث المارّ بهالوصلة: الكاميرات الي **ورا** هالوصلة
      const from = doc.nodes.find((n) => n.id === l.from.node)
      const to = doc.nodes.find((n) => n.id === l.to.node)
      const camSide = from?.partId === 'ip_camera' ? from : to?.partId === 'ip_camera' ? to : null
      // وصلة كاميرا مفردة: بثها لحاله
      if (camSide) continue
      // وصلة صاعدة (سويچ ← مسجّل): كل البث يمر بيها
      const isUplink = (from?.partId === 'nvr' || to?.partId === 'nvr')
      if (isUplink && totalMbps > cap * 0.8) {
        messages.push({
          kind: totalMbps > cap ? 'error' : 'warn',
          text: `${totalMbps > cap ? '🔴' : '⚠️'} الرابط ${c.aName} ⇄ ${c.bName} سعته ${cap} ميغابت ويمر بيه ${totalMbps.toFixed(1)}. هذا **عنق الزجاجة** — كل الكاميرات تمر من هنا، والبث يتقطّع لمن تنفتح كلها سوا.`,
        })
        if (totalMbps > cap) bad = true
      }
    }

    if (!bad) {
      messages.push({
        kind: 'info',
        text: `✅ المنظومة سليمة: ${cams.length} كاميرا · ${totalMbps.toFixed(1)} ميغابت · ${totalW.toFixed(0)} واط ليلاً.`,
      })
    }
    return { ok: !bad, messages, nodeReadings, linkState }
  },
}
