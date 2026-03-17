const lotteryDisplayNameMap: Record<string, string> = {
  ssq: "福彩双色球",
  dlt: "体彩大乐透",
};

export function getLotteryDisplayName(code?: string | null, fallback?: string) {
  if (!code) {
    return fallback || "待识别";
  }

  return lotteryDisplayNameMap[code.toLowerCase()] || fallback || code.toUpperCase();
}
