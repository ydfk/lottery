import type { ParsedEntry, RecommendationEntry, TicketEntryDraft } from "@/types/lottery";

const NUMBER_SPLIT_PATTERN = /[\s,，、;；|/]+/;

const ENTRY_RULES: Record<string, { redCount: number; blueCount: number }> = {
  ssq: { redCount: 6, blueCount: 1 },
  dlt: { redCount: 5, blueCount: 2 },
};

function splitCompactNumbers(value: string, expectedCount?: number) {
  const compact = value.replace(/\D/g, "");
  if (!compact || compact.length % 2 !== 0) {
    return [];
  }
  if (expectedCount && compact.length !== expectedCount * 2) {
    return [];
  }

  const numbers: number[] = [];
  for (let index = 0; index < compact.length; index += 2) {
    numbers.push(Number(compact.slice(index, index + 2)));
  }
  return numbers;
}

function parseNumberList(value: string, expectedCount?: number) {
  const source = value.trim();
  if (!source) {
    return [];
  }

  const tokens = source
    .split(NUMBER_SPLIT_PATTERN)
    .map((item) => item.replace(/\D/g, ""))
    .filter(Boolean);

  if (tokens.length === 1) {
    const compactNumbers = splitCompactNumbers(tokens[0], expectedCount);
    if (compactNumbers.length > 0) {
      return compactNumbers;
    }
  }

  return tokens
    .map((item) => Number(item))
    .filter((item) => Number.isFinite(item) && item >= 0);
}

function formatNumberList(numbers: number[]) {
  return numbers.map((value) => value.toString().padStart(2, "0")).join(",");
}

export function createEmptyEntryDraft(): TicketEntryDraft {
  return {
    redNumbers: "",
    blueNumbers: "",
    multiple: 1,
    isAdditional: false,
  };
}

export function buildDraftsFromParsedEntries(entries: ParsedEntry[]) {
  if (entries.length === 0) {
    return [createEmptyEntryDraft()];
  }

  return entries.map((entry) => ({
    redNumbers: formatNumberList(entry.red),
    blueNumbers: formatNumberList(entry.blue),
    multiple: Math.max(1, entry.multiple || 1),
    isAdditional: Boolean(entry.isAdditional),
  }));
}

export function buildDraftsFromRecommendationEntries(entries: RecommendationEntry[]) {
  if (entries.length === 0) {
    return [createEmptyEntryDraft()];
  }

  return entries.map((entry) => ({
    redNumbers: entry.redNumbers,
    blueNumbers: entry.blueNumbers,
    multiple: 1,
    isAdditional: false,
  }));
}

export function normalizeDraftsForLottery(entryDrafts: TicketEntryDraft[], lotteryCode: string) {
  if (entryDrafts.length === 0) {
    return [createEmptyEntryDraft()];
  }

  return entryDrafts.map((entry) => ({
    ...entry,
    isAdditional: lotteryCode === "dlt" ? entry.isAdditional : false,
  }));
}

export function draftToParsedEntry(entryDraft: TicketEntryDraft, lotteryCode: string): ParsedEntry | null {
  const rules = ENTRY_RULES[lotteryCode];
  const red = parseNumberList(entryDraft.redNumbers, rules?.redCount);
  const blue = parseNumberList(entryDraft.blueNumbers, rules?.blueCount);

  if (red.length === 0 || blue.length === 0) {
    return null;
  }

  return {
    red,
    blue,
    multiple: Math.max(1, entryDraft.multiple || 1),
    isAdditional: lotteryCode === "dlt" ? Boolean(entryDraft.isAdditional) : false,
  };
}

export function buildParsedEntriesFromDrafts(entryDrafts: TicketEntryDraft[], lotteryCode: string) {
  return entryDrafts
    .map((entryDraft) => draftToParsedEntry(entryDraft, lotteryCode))
    .filter((entry): entry is ParsedEntry => Boolean(entry));
}
