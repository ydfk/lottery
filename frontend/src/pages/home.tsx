import { useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {useAuthStore} from '@/store/auth-store';
import { LotteryCard } from '../components/lottery/lottery-card';
import { Button } from '../components/ui/button';
import { Loader2 } from 'lucide-react';
import { BottomNav } from '../components/layout/bottom-nav';
import { useLotteryData } from '../hooks/use-lottery-data';

export default function HomePage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();

  const { 
    recommendations, 
    isLoading, 
    error,
    hasMore,
    loadMore,
    refresh,
    getLotteryTypeName,
    updatePurchaseStatus
  } = useLotteryData();
  
  // 检查认证并重定向
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);
  
  // 设置交叉观察器实现无限滚动
  const lastElementRef = useCallback(
    (node: HTMLDivElement | null) => {
      if (!node || isLoading || !hasMore) return;
      
      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && hasMore) {
            loadMore();
          }
        },
        { threshold: 0.5 }
      );
      
      observer.observe(node);
      return () => observer.disconnect();
    },
    [isLoading, hasMore, loadMore]
  );
  
  // 强制刷新数据
  const handleRefresh = () => {
    console.log('手动刷新数据');
    refresh();
  };
  
  return (
    <div className="container max-w-lg mx-auto px-4 py-8 pb-24">
      {/* 页头区域 */}
      <header className="flex justify-between items-center mb-6 sticky top-0 bg-white/90 dark:bg-gray-900/90 backdrop-blur-sm py-2 z-10">
        <h1 className="text-2xl font-bold">彩票推荐</h1>
        <Button size="sm" variant="outline" onClick={handleRefresh} disabled={isLoading}>
          刷新
        </Button>
      </header>

      {/* 错误消息 */}
      {error && (
        <div className="p-4 mb-4 bg-red-50 dark:bg-red-900/20 rounded-md text-red-700 dark:text-red-400">
          <p>{error.toString()}</p>
          <Button 
            variant="outline" 
            className="mt-2" 
            onClick={refresh}
          >
            重试
          </Button>
        </div>
      )}
      
      {/* 推荐列表 */}
      {recommendations.length > 0 ? (
        <div className="space-y-4 mb-8">
          {recommendations.map((recommendation, index) => (
            <div
              key={recommendation.id}
              ref={index === recommendations.length - 1 ? lastElementRef : undefined}
            >
              <LotteryCard
                recommendation={recommendation}
                lotteryTypeName={getLotteryTypeName(recommendation.lotteryTypeID)}
                onPurchaseChange={updatePurchaseStatus}
              />
            </div>
          ))}
        </div>
      ) : !isLoading ? (
        <div className="text-center py-20 text-gray-500">
          <p className="mb-4">暂无推荐数据</p>
          <Button onClick={refresh}>刷新</Button>
        </div>
      ) : null}
      
      {/* 加载指示器 */}
      {isLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}
      
      {/* 无限滚动观察目标 */}
      <div className="h-4" />
      
      {/* 没有更多结果的提示 */}
      {!isLoading && !hasMore && recommendations.length > 0 && (
        <div className="text-center py-4 text-gray-500">
          没有更多推荐了
        </div>
      )}

      {/* 底部导航 */}
      <BottomNav />
    </div>
  );
}