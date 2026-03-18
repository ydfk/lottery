import { apiGet, apiPost } from "../client";
import type {
  DashboardData,
  ParsedEntry,
  Recommendation,
  Ticket,
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

export function getRecommendations(limit = 20) {
  return apiGet<Recommendation[]>(`/api/lotteries/${LOTTERY_CODE}/recommendations?limit=${limit}`);
}

export function getLatestRecommendation() {
  return apiGet<Recommendation>(`/api/lotteries/${LOTTERY_CODE}/recommendations/latest`);
}

export function getRecommendationDetail(recommendationId: string) {
  return apiGet<Recommendation>(`/api/lotteries/${LOTTERY_CODE}/recommendations/${recommendationId}`);
}

export function generateRecommendation(count = 5) {
  return apiPost<Recommendation, { count: number }>(
    `/api/lotteries/${LOTTERY_CODE}/recommendations/generate`,
    { count }
  );
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
    purchasedAt?: string;
    costAmount?: number;
    notes?: string;
    entries?: Array<{ redNumbers: string; blueNumbers: string; multiple?: number; isAdditional?: boolean }>;
  }>(`/api/lotteries/tickets`, body);
}

export function recheckTicket(ticketId: string) {
  return apiPost<Ticket>(`/api/lotteries/tickets/${ticketId}/recheck`);
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
