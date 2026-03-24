import { apiDelete, apiGet, apiPost } from "../client";
import type {
  DashboardData,
  ParsedEntry,
  Recommendation,
  RecommendationFilters,
  RecommendationPage,
  Ticket,
  TicketHistoryFilters,
  TicketHistoryPage,
  TicketRecognitionDraft,
  TicketUpload,
} from "@/types/lottery";

const LOTTERY_CODE = "ssq";

export function getDashboard() {
  return apiGet<DashboardData>(`/api/lotteries/dashboard`);
}

export function getTickets() {
  return apiGet<Ticket[]>(`/api/lotteries/${LOTTERY_CODE}/tickets?limit=20`);
}

export function getAllTickets() {
  return apiGet<Ticket[]>(`/api/lotteries/tickets?limit=20`);
}

export function getTicketHistory(page: number, pageSize: number, filters: TicketHistoryFilters) {
  const query = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    sort: filters.sort || "latest",
  });
  if (filters.lotteryCode) {
    query.set("lotteryCode", filters.lotteryCode);
  }
  if (filters.status) {
    query.set("status", filters.status);
  }
  return apiGet<TicketHistoryPage>(`/api/lotteries/tickets/history?${query.toString()}`);
}

function normalizeRecommendationPage(data: RecommendationPage | Recommendation[]): RecommendationPage {
  if (Array.isArray(data)) {
    return {
      items: data,
      page: 1,
      pageSize: data.length,
      total: data.length,
      hasMore: false,
    };
  }

  return {
    items: Array.isArray(data.items) ? data.items : [],
    page: data.page || 1,
    pageSize: data.pageSize || 10,
    total: data.total || 0,
    hasMore: Boolean(data.hasMore),
  };
}

export async function getRecommendations(page: number, pageSize: number, filters: RecommendationFilters) {
  const query = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    sort: filters.sort || "latest",
  });
  if (filters.lotteryCode) {
    query.set("lotteryCode", filters.lotteryCode);
  }
  if (filters.status) {
    query.set("status", filters.status);
  }
  const data = await apiGet<RecommendationPage | Recommendation[]>(
    `/api/lotteries/recommendations?${query.toString()}`
  );
  return normalizeRecommendationPage(data);
}

export function getLatestRecommendation() {
  return apiGet<Recommendation>(`/api/lotteries/${LOTTERY_CODE}/recommendations/latest`);
}

export function getRecommendationDetail(lotteryCode: string, recommendationId: string) {
  return apiGet<Recommendation>(`/api/lotteries/${lotteryCode}/recommendations/${recommendationId}`);
}

export function generateRecommendation(lotteryCode: string) {
  return apiPost<Recommendation>(`/api/lotteries/${lotteryCode}/recommendations/generate`);
}

export function syncDraws() {
  return apiPost<{ syncedCount: number }>(`/api/lotteries/${LOTTERY_CODE}/draws/sync`);
}

export function uploadTicketImage(formData: FormData) {
  return apiPost<TicketUpload, FormData>(`/api/lotteries/tickets/upload-image`, formData);
}

export function recognizeTicket(body: { uploadId: string; ocrText?: string }) {
  return apiPost<TicketRecognitionDraft, { uploadId: string; ocrText?: string }>(
    `/api/lotteries/tickets/recognize`,
    body
  );
}

export function createTicket(body: {
  lotteryCode?: string;
  uploadId: string;
  recommendationId?: string;
  issue?: string;
  drawDate?: string;
  purchasedAt?: string;
  costAmount?: number;
  notes?: string;
  entries?: Array<{ redNumbers: string; blueNumbers: string; multiple?: number; isAdditional?: boolean }>;
}) {
  return apiPost<Ticket, {
    lotteryCode?: string;
    uploadId: string;
    recommendationId?: string;
    issue?: string;
    drawDate?: string;
    purchasedAt?: string;
    costAmount?: number;
    notes?: string;
    entries?: Array<{ redNumbers: string; blueNumbers: string; multiple?: number; isAdditional?: boolean }>;
  }>(`/api/lotteries/tickets`, body);
}

export function recheckTicket(ticketId: string) {
  return apiPost<Ticket>(`/api/lotteries/tickets/${ticketId}/recheck`);
}

export function deleteTicket(ticketId: string) {
  return apiDelete<{ deleted: boolean }>(`/api/lotteries/tickets/${ticketId}`);
}

export function deleteRecommendation(lotteryCode: string, recommendationId: string) {
  return apiDelete<{ deleted: boolean }>(`/api/lotteries/${lotteryCode}/recommendations/${recommendationId}`);
}

export function scanTicket(formData: FormData) {
  return apiPost<Ticket, FormData>(`/api/lotteries/${LOTTERY_CODE}/tickets/scan`, formData);
}

export function formatParsedEntry(entry: ParsedEntry) {
  return {
    redNumbers: entry.red.map((value) => value.toString().padStart(2, "0")).join(","),
    blueNumbers: entry.blue.map((value) => value.toString().padStart(2, "0")).join(","),
  };
}
