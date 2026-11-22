// (ك) 11-22 زينا: أيقونات خط بسيطة لتصاميم الصفحات (ماكو مكتبة أيقونات بالمشروع)
const P: Record<string, string> = {
  cart: 'M3 4h2l2.4 11h10.2L20 8H6.2M9 20a1 1 0 1 0 0-2 1 1 0 0 0 0 2Zm8 0a1 1 0 1 0 0-2 1 1 0 0 0 0 2Z',
  wallet: 'M4 7h14a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a1 1 0 0 1-1-1V7Zm0 0 11-3v3M15 13h5',
  clock: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Zm0-13v5l3 2',
  bars: 'M6 20V14m6 6V8m6 12V4',
  layers: 'M12 3l9 5-9 5-9-5 9-5Zm-9 9l9 5 9-5M3 16l9 5 9-5',
  shield: 'M12 3 5 6v6c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z',
  clip: 'M9 4h6v3H9zM7 5H5v16h14V5h-2M9 12h6m-6 4h4',
  tag: 'M3 12V4h8l9 9-8 8-9-9Zm4.5-4.5h.01',
  bell: 'M6 16V11a6 6 0 1 1 12 0v5l2 2H4l2-2Zm4 4h4',
  card: 'M3 6h18v12H3zM3 10h18M7 15h4',
  cal: 'M4 6h16v14H4zM4 10h16M8 3v4m8-4v4M8 14h2m4 0h2m-8 3h2',
  search: 'M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14Zm10 3-5-5',
  down: 'm6 9 6 6 6-6',
  left: 'm15 6-6 6 6 6',
  right: 'm9 6 6 6-6 6',
  box: 'M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3Zm0 9 8-4.5M12 12v9M12 12 4 7.5',
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9a7 7 0 0 1 14 0',
  trendDown: 'M3 7l6 6 4-4 8 8m0 0v-5m0 5h-5',
}
export function LineIcon({ n, className = 'h-4 w-4' }: { n: string; className?: string }) {
  return <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className={className} aria-hidden="true"><path d={P[n]} /></svg>
}
