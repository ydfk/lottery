import { apiGet, apiPost } from "../client";
import type { DashboardData, Recommendation, Ticket } from "@/types/lottery";

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

export function scanTicket(formData: FormData) {
  return apiPost<Ticket, FormData>(`/api/lotteries/${LOTTERY_CODE}/tickets/scan`, formData);
}
