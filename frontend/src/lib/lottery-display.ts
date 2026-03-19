const lotteryDisplayNameMap: Record<string, string> = {
  ssq: "福彩双色球",
  dlt: "体彩大乐透",
};

export const lotteryDisplayOptions = Object.entries(lotteryDisplayNameMap).map(([code, name]) => ({
  code,
  name,
}));

export function getLotteryDisplayName(code?: string | null, fallback?: string) {
  if (!code) {
    return fallback || "待识别";
  }

  return lotteryDisplayNameMap[code.toLowerCase()] || fallback || code.toUpperCase();
}

export function formatLotteryIssue(code?: string | null, issue?: string | null) {
  if (!issue) {
    return "";
  }

  if (code === "dlt" && /^\d{5}$/.test(issue)) {
    return `20${issue}`;
  }

  return issue;
}

export function formatLotteryDrawDate(drawDate?: string | null) {
  if (!drawDate) {
    return "";
  }

  const parsed = new Date(drawDate);
  if (!Number.isNaN(parsed.getTime())) {
    const year = parsed.getFullYear();
    const month = `${parsed.getMonth() + 1}`.padStart(2, "0");
    const day = `${parsed.getDate()}`.padStart(2, "0");
    return `${year}-${month}-${day}`;
  }

  if (drawDate.length >= 10) {
    return drawDate.slice(0, 10);
  }
  return drawDate;
}
