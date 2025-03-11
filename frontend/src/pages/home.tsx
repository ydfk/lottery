import { useEffect, useRef, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLotteryStore } from '../store/lottery-store';
import { useAuthStore } from '../store/auth-store';
import { LotteryCard } from '../components/lottery/lottery-card';
import { Button } from '../components/ui/button';
import { Loader2 } from 'lucide-react';
import { BottomNav } from '../components/layout/bottom-nav';

export default function HomePage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const { 
    recommendations, 
    isLoading, 
    error, 
    fetchRecommendations, 
    fetchLotteryTypes,
    hasMore 
  } = useLotteryStore();
  
  // 调试状态，用于显示推荐数据加载情况
  const [debugInfo, setDebugInfo] = useState({ 
    loaded: false,
    count: 0 
  });
  
  // Effect to check authentication and redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);
  
  // Effect to fetch initial data
  useEffect(() => {
    if (isAuthenticated) {
      console.log('正在获取彩票数据...');
      // 获取彩票类型数据
      fetchLotteryTypes();
      // 获取推荐数据
      fetchRecommendations(true);  // Reset and fetch first page
    }
  }, [isAuthenticated, fetchRecommendations, fetchLotteryTypes]);
  
  // 记录推荐数据变化
  useEffect(() => {
    console.log('推荐数据更新:', recommendations);
    console.log('数据加载状态:', isLoading);
    console.log('是否有更多数据:', hasMore);
    
    setDebugInfo({
      loaded: true,
      count: recommendations?.length || 0
    });
  }, [recommendations, isLoading, hasMore]);
  
  // Setup intersection observer for infinite scrolling
  const lastElementRef = useCallback(
    (node: HTMLDivElement | null) => {
      if (isLoading) return;
      
      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && hasMore) {
            fetchRecommendations();
          }
        },
        { threshold: 0.5 }
      );
      
      if (node) observer.observe(node);
      
      return () => {
        if (node) observer.unobserve(node);
      };
    },
    [isLoading, hasMore, fetchRecommendations]
  );
  
  // 强制刷新数据
  const handleRefresh = () => {
    console.log('手动刷新数据');
    fetchRecommendations(true);
  };
  
  return (
    <div className="container max-w-lg mx-auto px-4 py-8 pb-24">
      {/* Header Section */}
      <header className="flex justify-between items-center mb-6 sticky top-0 bg-white/90 dark:bg-gray-900/90 backdrop-blur-sm py-2 z-10">
        <h1 className="text-2xl font-bold">彩票推荐</h1>
        <Button size="sm" variant="outline" onClick={handleRefresh} disabled={isLoading}>
          刷新
        </Button>
      </header>

      {/* Debug Info - 仅开发环境显示 */}
      {process.env.NODE_ENV === 'development' && (
        <div className="mb-4 p-2 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          <p>Debug: 数据已加载: {debugInfo.loaded ? '是' : '否'}</p>
          <p>Debug: 推荐数量: {debugInfo.count}</p>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div className="p-4 mb-4 bg-red-50 dark:bg-red-900/20 rounded-md text-red-700 dark:text-red-400">
          <p>{error}</p>
          <Button 
            variant="outline" 
            className="mt-2" 
            onClick={() => fetchRecommendations(true)}
          >
            重试
          </Button>
        </div>
      )}
      
      {/* Recommendations List */}
      {recommendations && recommendations.length > 0 ? (
        <div className="space-y-4 mb-8">
          {recommendations.map((recommendation) => (
            <LotteryCard 
              key={recommendation.id} 
              recommendation={recommendation}
            />
          ))}
        </div>
      ) : !isLoading ? (
        <div className="text-center py-20 text-gray-500">
          <p className="mb-4">暂无推荐数据</p>
          <Button onClick={() => fetchRecommendations(true)}>刷新</Button>
        </div>
      ) : null}
      
      {/* Loading indicator */}
      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}
      
      {/* Observer target for infinite scrolling */}
      <div ref={lastElementRef} className="h-4" />
      
      {/* No more results indicator */}
      {!isLoading && hasMore === false && recommendations && recommendations.length > 0 && (
        <div className="text-center py-4 text-gray-500">
          没有更多推荐了
        </div>
      )}

      {/* Bottom Navigation */}
      <BottomNav />
    </div>
  );
}