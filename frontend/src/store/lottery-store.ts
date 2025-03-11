import { create } from 'zustand';
import { 
  getRecommendations as apiGetRecommendations,
  updatePurchaseStatus as apiUpdatePurchaseStatus,
  getLotteryTypes as apiGetLotteryTypes,
  Recommendation,
  LotteryType
} from '../lib/api/methods/lottery';

interface LotteryState {
  // 推荐数据相关
  recommendations: Recommendation[];
  isLoading: boolean;
  error: string | null;
  page: number;
  pageSize: number;
  hasMore: boolean;
  total: number;
  
  // 彩票类型相关
  lotteryTypes: LotteryType[];
  isLoadingTypes: boolean;
  typesError: string | null;
  
  // Actions
  fetchRecommendations: (reset?: boolean) => Promise<void>;
  updatePurchaseStatus: (id: number, isPurchased: boolean) => Promise<void>;
  fetchLotteryTypes: () => Promise<void>;
  getLotteryTypeName: (typeId: number) => string;
  clearError: () => void;
}

export const useLotteryStore = create<LotteryState>((set, get) => ({
  // 推荐数据相关
  recommendations: [],
  isLoading: false,
  error: null,
  page: 1,
  pageSize: 20,
  hasMore: false,
  total: 0,
  
  // 彩票类型相关
  lotteryTypes: [],
  isLoadingTypes: false,
  typesError: null,
  
  fetchRecommendations: async (reset = false) => {
    const currentState = get();
    const newPage = reset ? 1 : currentState.page;
    
    // Don't fetch if we're already loading or there are no more results
    if (currentState.isLoading || (!reset && !currentState.hasMore)) {
      return;
    }
    
    set({ isLoading: true, error: null });
    
    try {
      const response = await apiGetRecommendations(newPage, currentState.pageSize);
      
      set((state) => ({ 
        recommendations: reset ? response.data : [...state.recommendations, ...response.data],
        page: newPage + 1,
        hasMore: response.hasMore,
        total: response.total,
        isLoading: false,
      }));
    } catch (error) {
      let errorMessage = 'Failed to fetch recommendations';
      if (error instanceof Error) {
        errorMessage = error.message;
      }
      set({ isLoading: false, error: errorMessage });
    }
  },
  
  updatePurchaseStatus: async (id: number, isPurchased: boolean) => {
    set({ isLoading: true, error: null });
    
    try {
      await apiUpdatePurchaseStatus(id, isPurchased);
      
      // Update local state
      set((state) => ({
        recommendations: state.recommendations.map((rec) => 
          rec.id === id ? { ...rec, isPurchased } : rec
        ),
        isLoading: false,
      }));
    } catch (error) {
      let errorMessage = 'Failed to update purchase status';
      if (error instanceof Error) {
        errorMessage = error.message;
      }
      set({ isLoading: false, error: errorMessage });
    }
  },
  
  fetchLotteryTypes: async () => {
    const currentState = get();
    
    // 如果已经有彩票类型数据，不重复获取
    if (currentState.lotteryTypes.length > 0) {
      return;
    }
    
    set({ isLoadingTypes: true, typesError: null });
    
    try {
      const response = await apiGetLotteryTypes();
      set({ lotteryTypes: response, isLoadingTypes: false });
    } catch (error) {
      let errorMessage = 'Failed to fetch lottery types';
      if (error instanceof Error) {
        errorMessage = error.message;
      }
      set({ isLoadingTypes: false, typesError: errorMessage });
    }
  },
  
  getLotteryTypeName: (typeId: number) => {
    const { lotteryTypes } = get();
    const type = lotteryTypes.find(type => type.id === typeId);
    return type ? type.name : `类型 ${typeId}`;
  },
  
  clearError: () => set({ error: null, typesError: null }),
}));