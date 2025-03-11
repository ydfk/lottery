import { useEffect, useRef, useCallback } from 'react';
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
  
  // Ref for the observer target element
  const observerTarget = useRef<HTMLDivElement>(null);
  
  // Effect to check authentication and redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);
  
  // Effect to fetch initial data
  useEffect(() => {
    if (isAuthenticated) {
      // 获取彩票类型数据
      fetchLotteryTypes();
      // 获取推荐数据
      fetchRecommendations(true);  // Reset and fetch first page
    }
  }, [isAuthenticated, fetchRecommendations, fetchLotteryTypes]);
  
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
  
  return (
    <div className="container max-w-lg mx-auto px-4 py-8 pb-24">
      <header className="flex justify-between items-center mb-6 sticky top-0 bg-white/90 backdrop-blur-sm py-2 z-10">
        <h1 className="text-2xl font-bold">彩票推荐列表</h1>
      </header>

      {error && (
        <div className="p-4 mb-4 bg-red-50 rounded-md text-red-700">
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
      
      <div className="space-y-4 mb-8">
        {recommendations?.map((recommendation) => (
          <LotteryCard 
            key={recommendation.id} 
            recommendation={recommendation}
          />
        ))}
      </div>
      
      {/* Loading indicator */}
      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}
      
      {/* Observer target for infinite scrolling */}
      <div ref={lastElementRef} className="h-4" />
      
      {/* No more results indicator */}
      {!isLoading && !hasMore && recommendations && recommendations.length > 0 && (
        <div className="text-center py-4 text-gray-500">
          没有更多推荐了
        </div>
      )}
      
      {/* Empty state */}
      {!isLoading && recommendations && recommendations.length === 0 && (
        <div className="text-center py-20 text-gray-500">
          <p className="mb-4">暂无推荐数据</p>
          <Button onClick={() => fetchRecommendations(true)}>刷新</Button>
        </div>
      )}

      {/* Bottom Navigation */}
      <BottomNav />
    </div>
  );
}