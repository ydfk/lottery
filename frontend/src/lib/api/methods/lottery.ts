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
  return apiGet<DashboardData>(`/api/lotteries/${LOTTERY_CODE}/dashboard`);
}

export function getTickets() {
  return apiGet<Ticket[]>(`/api/lotteries/${LOTTERY_CODE}/tickets?limit=20`);
}

export function getLatestRecommendation() {
  return apiGet<Recommendation>(`/api/lotteries/${LOTTERY_CODE}/recommendations/latest`);
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
  return apiPost<TicketUpload, FormData>(`/api/lotteries/${LOTTERY_CODE}/tickets/upload-image`, formData);
}

export function recognizeTicket(body: { uploadId: string; ocrText?: string }) {
  return apiPost<TicketRecognitionDraft, { uploadId: string; ocrText?: string }>(
    `/api/lotteries/${LOTTERY_CODE}/tickets/recognize`,
    body
  );
}

export function createTicket(body: {
  uploadId: string;
  issue?: string;
  purchasedAt?: string;
  notes?: string;
  entries?: Array<{ redNumbers: string; blueNumbers: string }>;
}) {
  return apiPost<Ticket, {
    uploadId: string;
    issue?: string;
    purchasedAt?: string;
    notes?: string;
    entries?: Array<{ redNumbers: string; blueNumbers: string }>;
  }>(`/api/lotteries/${LOTTERY_CODE}/tickets`, body);
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
