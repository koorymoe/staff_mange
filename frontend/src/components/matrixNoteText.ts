// يقسّم النص لجمل: نقطة، أو «؛»، أو سطر جديد. «—» داخل الجملة يبقى.
export function splitSentences(t: string): string[] {
  return t.split(/(?<=[.!؟?؛])\s+|\n+/).map((x) => x.trim()).filter(Boolean)
}

