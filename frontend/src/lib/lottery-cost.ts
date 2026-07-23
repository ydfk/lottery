export function calculateEntriesCost(entries: Array<{ multiple: number; isAdditional: boolean }>) {
  return entries.reduce((total, entry) => {
    const perBetCost = entry.isAdditional ? 3 : 2;
    return total + Math.max(1, Math.min(99, entry.multiple)) * perBetCost;
  }, 0);
}
