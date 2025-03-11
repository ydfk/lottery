import alovaInstance from '../index';

// Types based on backend models
export interface LotteryType {
  id: number;
  code: string;
  name: string;
  scheduleCron: string;
  modelName: string;
  isActive: boolean;
  caipiaoId: number;
}

export interface Recommendation {
  id: number;
  lotteryTypeID: number;
  numbers: string;
  modelName: string;
  drawTime: string | null;
  expectedDrawTime: string;
  drawNumber: string;
  isPurchased: boolean;
  drawResult: string;
  winStatus: string;
  winAmount: number;
  createdAt: string;
  updatedAt: string;
}

export interface RecommendationResponse {
  data: Recommendation[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

// Get lottery types
export const getLotteryTypes = () => {
  return alovaInstance.Get<LotteryType[]>('/lottery-types');
};

// Get recommendations with pagination
export const getRecommendations = (page: number = 1, pageSize: number = 20) => {
  return alovaInstance.Get<RecommendationResponse>('/recommendations', {
    params: { page, pageSize }
  });
};

// Update purchase status of a recommendation
export const updatePurchaseStatus = (id: number, isPurchased: boolean) => {
  return alovaInstance.Put(`/recommendations/${id}/purchase`, { isPurchased });
};