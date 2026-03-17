export interface LotteryType {
  id: string;
  code: string;
  name: string;
  status: string;
  recommendationCount: number;
  recommendationProvider: string;
  recommendationModel: string;
  visionProvider: string;
  visionModel: string;
}

export interface DrawResult {
  id: string;
  issue: string;
  drawDate: string;
  redNumbers: string;
  blueNumbers: string;
  saleAmount: number;
  prizePoolAmount: number;
}

export interface RecommendationEntry {
  id: string;
  sequence: number;
  redNumbers: string;
  blueNumbers: string;
  confidence: number;
  reason: string;
  isWinning: boolean;
  prizeName: string;
  prizeAmount: number;
  matchSummary: string;
}

export interface Recommendation {
  id: string;
  lotteryCode: string;
  issue: string;
  provider: string;
  model: string;
  strategy: string;
  promptVersion: string;
  summary: string;
  basis: string;
  checkedAt?: string;
  prizeAmount: number;
  createdAt: string;
  entries: RecommendationEntry[];
  entryCount?: number;
  winningCount?: number;
  isPurchased?: boolean;
  purchasedTicket?: Ticket;
}

export interface TicketEntry {
  id: string;
  sequence: number;
  redNumbers: string;
  blueNumbers: string;
  multiple: number;
  isWinning: boolean;
  prizeName: string;
  prizeAmount: number;
  matchSummary: string;
}

export interface Ticket {
  id: string;
  lotteryCode: string;
  recommendationId?: string;
  issue: string;
  status: string;
  costAmount: number;
  prizeAmount: number;
  purchasedAt: string;
  drawDate?: string;
  recognizedText: string;
  imageUrl: string;
  entries: TicketEntry[];
}

export interface ParsedEntry {
  red: number[];
  blue: number[];
  multiple: number;
}

export interface TicketUpload {
  id: string;
  lotteryCode: string;
  status: string;
  originalFilename: string;
  imagePath: string;
  imageUrl: string;
  recognizedText: string;
  recognitionIssue: string;
  recognitionConfidence: number;
  errorMessage: string;
}

export interface TicketRecognitionDraft {
  upload: TicketUpload;
  lotteryCode: string;
  issue: string;
  rawText: string;
  confidence: number;
  entries: ParsedEntry[];
}

export interface DashboardStats {
  totalTickets: number;
  wonTickets: number;
  totalCost: number;
  totalPrize: number;
}

export interface DashboardData {
  lottery: LotteryType;
  latestDraw?: DrawResult;
  latestRecommendation?: Recommendation;
  recentTickets: Ticket[];
  stats: DashboardStats;
}

export interface ApiResponse<T> {
  flag: boolean;
  code: number;
  data: T;
  msg?: string;
  time: string;
}
