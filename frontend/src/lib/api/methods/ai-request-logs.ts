import { apiGet } from "@/lib/api/client";

const endpoint = "/api/lotteries/ai-request-logs";

export type AIRequestPurpose = "recommendation" | "ocr";

export interface AIRequestLogSummary {
  id: number;
  startedAt: string;
  purpose: AIRequestPurpose;
  provider: string;
  model: string;
  lotteryCode: string;
  issue: string;
  attempt: number;
  statusCode: number;
  durationMs: number;
  success: boolean;
}

export interface AIRequestLog extends AIRequestLogSummary {
  method: string;
  endpoint: string;
  contentType: string;
  request: string;
  response: string;
  error: string;
}

interface AIRequestLogPage {
  items: AIRequestLogSummary[];
  page: number;
  pageSize: number;
  total: number;
  keepLimit: number;
}

export function listAIRequestLogs(
  page: number,
  purpose: AIRequestPurpose | "",
  signal: AbortSignal
) {
  const query = new URLSearchParams({ page: String(page) });
  if (purpose) query.set("purpose", purpose);
  return apiGet<AIRequestLogPage>(`${endpoint}?${query}`, signal);
}

export function getAIRequestLog(id: number, signal: AbortSignal) {
  return apiGet<AIRequestLog>(`${endpoint}/${id}`, signal);
}
