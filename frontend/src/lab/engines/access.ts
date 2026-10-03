// ═══ محرّك التحكم بالدخول ═══
//
// هذا المجال الوحيد بالمختبر الي الغلط بيه **ما يكلّف جهازاً — يكلّف
// ناساً**. قفل على باب مخرج طوارئ منصوب غلط يحبس الناس داخل حريق،
// والمنظومة تبدو شغّالة تماماً: تفتح بالبطاقة، وتقفل، وكل الأضوية
// خضر. ما تنكشف إلا يوم ما تنفع فيه المعرفة.
//
// ⚠️⚠️ **الوحدة الأساسية هنا هي «الباب» مو «القطعة».** كل فحص
// ينحسب لكل باب على حدة. محرّك يجمع الأرقام كمجموع واحد يخفي إن
// باباً واحداً من ستة خطر — وباب واحد يكفي.
//
// وخمسة أشياء يحسبها:
// ١) fail-safe مقابل fail-secure على مخارج الطوارئ — **قاتل**
// ٢) وجود مخرج حر بكل باب
// ٣) ميزانية التيار والبطارية — **ومقلوبة عن الحريق**
// ٤) هبوط الجهد على مسافة القفل
// ٥) الدايود على ملف القفل، وحد مسافة القارئ
//
// ⚠️ درجة الدقة `F1`: منطق ونقاط عمل بحالة مستقرة. ماكو نمذجة زمنية
// لتأخير الريلاي ولا لمنحنى الملف. تكفي لتصميم منظومة باب وتشخيص
// الأعطال أعلاه، وما تكفي لدرس توقيت.
//
// ⚠️⚠️ الأرقام أدناه **أعراف منشورة عامة** — مو من كتالوگ موديل
// بعينه. المحتوى يبقى `verified = FALSE` لحد ما يجرّبه فني.
//
// ⚠️ وما يناقض المحتوى الموجود بتمرين الكيباد (`seed_sim_lock.go`):
// هناك يتعلّم **توصيل جهاز واحد**، وهنا **تصميم منظومة باب**.

import { CABLE_GAUGES } from '../../sim3d/cable'
import type { DomainEngine, LabDoc, LabNode, SimResult } from '../types'

const num = (v: unknown, d: number) => {
  const n = typeof v === 'number' ? v : parseFloat(String(v))
  return Number.isFinite(n) ? n : d
}
const str = (v: unknown, d = '') => (v === undefined || v === null ? d : String(v))
const bool = (v: unknown, d: boolean) => (v === undefined || v === null ? d : v === true || v === 'true')

/** مقاومة النحاس النوعية — نفس ثابت `sim3d/cable.ts`. */
const RHO_COPPER = 0.0172

/** الأقفال. */
const LOCKS = new Set(['mag_lock', 'electric_strike'])
/** وسائل الخروج الحر — ما تحتاج بطاقة ولا معرفة. */
const EGRESS_DEVICES = new Set(['exit_button', 'rex_motion', 'break_glass'])

/** ساعات الاستعداد ودقائق التشغيل — نفس قاعدة الحريق. */
const STANDBY_H = 24
/** هامش تدهور البطارية. */
const DERATE = 1.25

/**
 * ═══ الجهد الواصل للقفل ═══
 *
 * ⚠️ نفس معادلة `cableResistance` بـ`sim3d/cable.ts` (الطول ×٢ لأن
 * الدائرة تروح وترجع) — بس تشتغل على وصلة اللوح مو على كيان الكيبل
 * القديم. المعادلة وحدة عمداً: قيمتان مختلفتان لنفس السلك بين
 * مختبرين تعني متدرّباً يتعلّم رقمين متناقضين.
 */
export function lockVoltage(sourceV: number, currentA: number, lengthM: number, gauge: string): number {
  const g = CABLE_GAUGES.find((x) => x.id === gauge) ?? CABLE_GAUGES[1]
  const r = (RHO_COPPER * lengthM * 2) / g.areaMm2
  return sourceV - currentA * r
}

/**
 * ═══ الأمبير·ساعة المطلوبة ═══
 *
 * ⚠️⚠️ **الفرق الجوهري عن الحريق**: القفل المغناطيسي يسحب تياره
 * **٢٤ ساعة كاملة** — لأنه يسحب حتى **يبقى مقفلاً**، مو حتى يفتح.
 * الاسترايك عكسه: ساكن بلا تيار، ويسحب لحظة الفتح بس.
 *
 * فبابان بنفس المنظومة، واحد مغناطيسي وواحد استرايك، يحتاجون
 * بطاريتين تختلفان **أضعافاً**. والفني الي يحسبها بطريقة الحريق
 * (استعداد صغير + ذروة قصيرة) يطلع برقم يخلّي المنظومة تنطفي بأول
 * انقطاع طويل — وهذا بالضبط وقت ما تنحتاج.
 */
export function requiredAh(standbyMa: number, hours = STANDBY_H): number {
  return (standbyMa / 1000) * hours * DERATE
}

interface Door {
  lock: LabNode
  name: string
  isEgress: boolean
  /** أجهزة الخروج المربوطة بنفس وحدة التحكم لهالباب. */
  egress: LabNode[]
  reader: LabNode | null
}

export const accessEngine: DomainEngine = {
  id: 'access',
  name: 'محرّك التحكم بالدخول',

  run(doc: LabDoc): SimResult {
    const messages: SimResult['messages'] = []
    const nodeReadings: SimResult['nodeReadings'] = {}
    const linkState: SimResult['linkState'] = {}
    const add = (id: string, text: string, tone?: 'ok' | 'warn' | 'bad') => {
      ;(nodeReadings[id] ??= []).push({ text, tone })
    }
    for (const l of doc.links) linkState[l.id] = 'ok'

    const neighbors = (id: string) =>
      doc.links
        .filter((l) => l.from.node === id || l.to.node === id)
        .map((l) => doc.nodes.find((n) => n.id === (l.from.node === id ? l.to.node : l.from.node)))
        .filter((n): n is LabNode => !!n)

    const linkOf = (a: string, b: string) =>
      doc.links.find((l) =>
        (l.from.node === a && l.to.node === b) || (l.to.node === a && l.from.node === b))

    const locks = doc.nodes.filter((n) => LOCKS.has(n.partId))
    const ctrl = doc.nodes.find((n) => n.partId === 'ac_controller')
    const psu = doc.nodes.find((n) => n.partId === 'ac_psu')

    if (locks.length === 0) {
      messages.push({ kind: 'warn', text: 'ماكو قفل — ضيف قفلاً حتى نبني باباً.' })
      return { ok: false, messages, nodeReadings, linkState }
    }
    if (!ctrl) {
      messages.push({ kind: 'warn', text: 'ماكو وحدة تحكم — ضيف وحدة واربط عليها القفل والقارئ.' })
    }

    // ═══ نبني الأبواب ═══
    //
    // ⚠️ الباب = قفل + الي مربوط بنفس وحدة التحكم. بمنظومة بأكثر من
    // باب، ربط كل شي بوحدة وحدة يخلّي الفحص يمرّ زوراً: زر خروج على
    // باب A يبين كأنه يغطّي باب B. ولهذا نطلب **رقم الباب** على كل
    // قطعة — وهذا بالضبط شلون تنكتب المخططات بالميدان.
    const doorNo = (n: LabNode) => num(n.params.door, 1)
    const doors: Door[] = locks.map((lock) => {
      const d = doorNo(lock)
      const same = (n: LabNode) => doorNo(n) === d
      return {
        lock,
        name: str(lock.params.name, `باب ${d}`),
        isEgress: bool(lock.params.isEgress, false),
        egress: doc.nodes.filter((n) => EGRESS_DEVICES.has(n.partId) && same(n)),
        reader: doc.nodes.find((n) => n.partId === 'card_reader' && same(n)) ?? null,
      }
    })

    let anyFatal = false
    let anyBad = false
    let totalStandbyMa = 0

    const fireRelay = doc.nodes.find((n) => n.partId === 'fire_relay')

    for (const dr of doors) {
      const isMag = dr.lock.partId === 'mag_lock'
      const holdMa = num(dr.lock.params.holdMa, isMag ? 500 : 0)
      const pulseMa = num(dr.lock.params.pulseMa, isMag ? 0 : 350)
      add(dr.lock.id, isMag ? 'fail-safe · يفصل بانقطاع التيار' : 'fail-secure · يبقى مقفلاً')

      // ═══ ١) القاتل: مخرج طوارئ بقفل fail-secure ═══
      //
      // ⚠️ ما ننسب الخطأ للقفل — ننسبه **للباب**. نفس الاسترايك على
      // باب مخزن صحيح تماماً، وعلى مخرج طوارئ جريمة. الفحص يقرا
      // «هذا الباب مخرج طوارئ؟» مو «شنو نوع القفل؟»
      if (dr.isEgress && !isMag) {
        anyFatal = true
        anyBad = true
        add(dr.lock.id, '☠️ يحبس الناس بالحريق', 'bad')
        messages.push({
          kind: 'error',
          text: `☠️ **«${dr.name}» مخرج طوارئ وعليه قفل كهربائي (استرايك)** — هذا النوع يحتاج كهرباء حتى **يفتح**، فينقطع التيار ويبقى الباب مقفلاً والناس داخل. المطلوب قفل مغناطيسي (fail-safe) ينفصل بانقطاع التيار. ⚠️ هذا مو خطأ يخرب جهازاً — هذا الي يموّت.`,
        })
      }
      if (dr.isEgress && isMag) add(dr.lock.id, '✅ ينفتح بانقطاع التيار', 'ok')

      // ═══ ٢) المخرج الحر ═══
      if (dr.egress.length === 0) {
        anyBad = true
        add(dr.lock.id, 'بلا مخرج حر', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${dr.name}» ماكو عليه وسيلة خروج حر (زر خروج أو حسّاس حركة أو كسر زجاج). الي جوّا ما يگدر يطلع إلا ببطاقة — وهذا يتحوّل لكارثة أول ما تنقطع الكهرباء أو تنعطّل الوحدة.`,
        })
      } else {
        add(dr.lock.id, `${dr.egress.length} وسيلة خروج`, 'ok')
      }

      // ⚠️ كسر الزجاج لازم **يقطع التغذية مباشرة** مو يمر بالوحدة:
      // وحدة معلّقة أو محروقة تخلّي زر الخروج بلا فايدة، والزجاج هو
      // الي يشتغل لمن كلشي ثاني يفشل. زجاج مربوط بالوحدة يعطي إحساس
      // أمان كاذب — وهذا أخطر من ماكو زجاج، لأن أحد ما راح يدوّر على
      // بديل.
      for (const bg of dr.egress.filter((e) => e.partId === 'break_glass')) {
        const onCtrl = ctrl ? neighbors(bg.id).some((n) => n.id === ctrl.id) : false
        if (onCtrl) {
          anyBad = true
          add(bg.id, 'مربوط بالوحدة', 'bad')
          messages.push({
            kind: 'error',
            text: `🔴 كسر الزجاج بـ«${dr.name}» مربوط بوحدة التحكم. لازم يقطع **تغذية القفل مباشرة** — وحدة معلّقة أو محروقة تخلّيه بلا فايدة، وهو المفروض آخر خط دفاع لمن كلشي ثاني يفشل.`,
          })
        }
      }

      // ═══ ٣) الدايود على ملف القفل ═══
      //
      // ⚠️ العَرَض هنا **متأخّر شهوراً**: نبضة الرجوع من الملف تاكل
      // تلامسات الريلاي كل مرة يفصل، فاللوحة «تخرب بلا سبب» بعد
      // أشهر. وأحد ما يربطها بقفل بلا دايود — يبدّلون اللوحة ويرجع
      // نفس الشي.
      if (!bool(dr.lock.params.diode, true)) {
        anyBad = true
        add(dr.lock.id, 'بلا دايود', 'bad')
        messages.push({
          kind: 'error',
          text: `🔴 «${dr.name}»: القفل بلا دايود على ملفه. كل مرة يفصل، نبضة الرجوع تاكل تلامسات الريلاي أو ترانزستور اللوحة. ⚠️ العَرَض يظهر **بعد أشهر** بشكل «اللوحة خربت بلا سبب» — وتبديلها بلا دايود يرجّع نفس العطل.`,
        })
      }

      // ═══ ٤) هبوط الجهد على مسافة القفل ═══
      const lk = ctrl ? linkOf(dr.lock.id, ctrl.id) : undefined
      const lenM = num(lk?.params?.lengthM, 15)
      const gauge = str(lk?.params?.gauge, 'awg20')
      const srcV = num(psu?.params.voltage, 12)
      const drawA = (isMag ? holdMa : pulseMa) / 1000
      const vAtLock = lockVoltage(srcV, drawA, lenM, gauge)
      const minV = num(dr.lock.params.minV, isMag ? 10.5 : 10.0)
      add(dr.lock.id, `${vAtLock.toFixed(2)} V واصل`, vAtLock < minV ? 'bad' : 'ok')
      if (vAtLock < minV) {
        anyBad = true
        messages.push({
          kind: 'error',
          text: `🔴 «${dr.name}»: الجهد الواصل ${vAtLock.toFixed(2)} فولت والقفل يحتاج ${minV} على الأقل — ${lenM} متر بمقطع ${gauge}. ${isMag ? 'المغناطيسي يمسك ضعيفاً: الباب **يبين مقفلاً وينفتح بدفعة**.' : 'الاسترايك يطقطق وما يسحب اللسان.'} الحل مقطع أكبر مو جهد أعلى.`,
        })
      }

      // ═══ ٥) مسافة القارئ (Wiegand) ═══
      if (dr.reader && ctrl) {
        const rl = linkOf(dr.reader.id, ctrl.id)
        const rlen = num(rl?.params?.lengthM, 20)
        add(dr.reader.id, `${rlen} م`)
        if (rlen > 150) {
          anyBad = true
          add(dr.reader.id, 'فوگ حد Wiegand', 'bad')
          messages.push({
            kind: 'error',
            text: `🔴 قارئ «${dr.name}» على ${rlen} متر، وحد Wiegand العملي ~١٥٠ متر. القارئ **يقرا مرة ويطنّش مرتين** — والفني يبدّل قارئين قبل ما يشك بالمسافة. الحل تقريب الوحدة أو محوّل RS-485.`,
          })
        }
      }

      // ═══ فصل الحريق ═══
      //
      // ⚠️ القفل المغناطيسي **لازم** يفصل بإشارة الحريق — وهذا شرط
      // منفصل عن انقطاع التيار: الحريق ممكن يصير والكهرباء شغّالة.
      if (isMag) {
        totalStandbyMa += holdMa
        if (!fireRelay) {
          anyBad = true
          messages.push({
            kind: 'error',
            text: `🔴 «${dr.name}» قفل مغناطيسي بلا ربط بإنذار الحريق. الحريق يصير **والكهرباء شغّالة** — فالقفل يبقى ماسكاً والباب ما ينفتح. لازم تماس من لوحة الحريق يقطع تغذية القفل.`,
          })
        } else if (bool(fireRelay.params.bypassed, false)) {
          anyBad = true
          add(fireRelay.id, 'مجسور', 'bad')
          messages.push({
            kind: 'error',
            text: `🔴 تماس الحريق **مجسور** — أحد جسّره حتى «يوگف الإزعاج». المنظومة تبدو شغّالة تماماً، والأقفال ما راح تفصل بالحريق.`,
          })
        }
      }
    }

    // ═══ ميزانية البطارية ═══
    //
    // ⚠️ نفس نمط الحريق بالضبط (mA ÷١٠٠٠ × ساعات × ١٫٢٥) — بس
    // المدخل مختلف جذرياً: هنا **سحب الإمساك المستمر**.
    const ctrlMa = num(ctrl?.params.standbyMa, 120)
    const readerMa = doc.nodes.filter((n) => n.partId === 'card_reader')
      .reduce((s, r) => s + num(r.params.standbyMa, 60), 0)
    const standbyMa = totalStandbyMa + ctrlMa + readerMa
    const needAh = requiredAh(standbyMa)

    if (ctrl) add(ctrl.id, `استعداد ${standbyMa.toFixed(0)} mA`)
    if (psu) {
      const have = num(psu.params.ah, 7)
      const maxA = num(psu.params.maxA, 3)
      add(psu.id, `${have} Ah`, have >= needAh ? 'ok' : 'bad')
      add(psu.id, `المطلوب ${needAh.toFixed(1)} Ah`, have >= needAh ? 'ok' : 'bad')
      if (have < needAh) {
        anyBad = true
        messages.push({
          kind: 'error',
          text: `🔋 البطارية ${have} أمبير·ساعة والمطلوب **${needAh.toFixed(1)}** لتغطية ${STANDBY_H} ساعة. ⚠️ وانتبه للسبب: الأقفال المغناطيسية تسحب ${totalStandbyMa} مللي أمبير **باستمرار** — تسحب حتى تبقى **مقفلة**، مو حتى تفتح. حساب على طريقة إنذار الحريق (استعداد صغير + ذروة قصيرة) يطلع برقم أصغر بأضعاف.`,
        })
      }
      if (standbyMa / 1000 > maxA) {
        anyBad = true
        messages.push({
          kind: 'error',
          text: `🔴 سحب ${(standbyMa / 1000).toFixed(2)} أمبير ومغذّي ${maxA} أمبير — فوگ طاقته. يسخن ويفصل، والأقفال تفتح لحالها.`,
        })
      }
    } else {
      messages.push({
        kind: 'warn',
        text: `ماكو مغذّي. المطلوب تقريباً **${needAh.toFixed(1)} أمبير·ساعة** احتياط لتغطية ${STANDBY_H} ساعة.`,
      })
    }

    if (!anyBad) {
      messages.push({
        kind: 'info',
        text: `✅ ${doors.length} باب سليم: كلها عليها مخرج حر، والمخارج بأقفال تنفصل بانقطاع التيار وبالحريق.`,
      })
    }
    if (anyFatal) {
      messages.push({
        kind: 'warn',
        text: '☠️ اكو خطأ **قاتل** بالمنظومة — صلّحه قبل أي شي ثاني.',
      })
    }
    return { ok: !anyBad, messages, nodeReadings, linkState }
  },
}
