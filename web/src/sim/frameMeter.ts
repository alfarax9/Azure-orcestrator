/**
 * Counts animation frames and the ones that ran long. The frame budget comes from the
 * display itself (1.5 × the median of the first 30 frames), so 60 Hz and 120 Hz both work.
 */
export function startFrameMeter() {
  let frames = 0;
  let dropped = 0;
  let worst = 0;
  let longTasks = 0;
  let last = 0;
  let budget = 0;
  const calibration: number[] = [];

  let raf = requestAnimationFrame(function tick(t) {
    if (last) {
      const delta = t - last;
      frames += 1;
      if (!budget) {
        calibration.push(delta);
        if (calibration.length === 30) budget = 1.5 * (calibration.sort((a, b) => a - b)[15] ?? 16.7);
      } else {
        if (delta > budget) dropped += 1;
        worst = Math.max(worst, delta);
      }
    }
    last = t;
    raf = requestAnimationFrame(tick);
  });

  let observer: PerformanceObserver | undefined;
  try {
    observer = new PerformanceObserver((list) => {
      longTasks += list.getEntries().length;
    });
    observer.observe({ type: "longtask" });
  } catch {
    observer = undefined; // Long task timing is Chromium-only.
  }

  return {
    read: () => ({ frames, dropped, worstMs: Math.round(worst), longTasks }),
    stop() {
      cancelAnimationFrame(raf);
      observer?.disconnect();
    },
  };
}
