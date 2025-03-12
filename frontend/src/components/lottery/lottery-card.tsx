import { useState } from 'react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Switch } from '../ui/switch';
import { Label } from '../ui/label';
import type { Recommendation } from '../../lib/api/methods/lottery';

interface LotteryCardProps {
  recommendation: Recommendation;
  lotteryTypeName: string; // 直接传入彩票类型名称
  onPurchaseChange: (id: number, isPurchased: boolean) => Promise<boolean | void>; // 直接传入更新购买状态的函数
}

export function LotteryCard({ 
  recommendation,
  lotteryTypeName,
  onPurchaseChange
}: LotteryCardProps) {
  const [isPurchasing, setIsPurchasing] = useState(false);

  // 检查是否已开奖
  const isDrawn = !!recommendation.drawTime;
  
  // 格式化开奖日期
  const formatDate = (dateString: string | null) => {
    if (!dateString) return '未开奖';
    return new Date(dateString).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // 处理购买状态切换
  const handlePurchaseToggle = async (checked: boolean) => {
    setIsPurchasing(true);
    try {
      await onPurchaseChange(recommendation.id, checked);
    } finally {
      setIsPurchasing(false);
    }
  };

  // 解析号码显示（假设格式为 "01,02,03,04,05,06+07"）
  const renderNumbers = () => {
    // 分割为主号码和特殊号码（如适用）
    const hasSpecial = recommendation.numbers.includes('+');
    
    if (hasSpecial) {
      const [mainNumbers, specialNumbers] = recommendation.numbers.split('+');
      return (
        <div className="flex flex-wrap gap-1 justify-center">
          {mainNumbers.split(',').map((num, index) => (
            <span 
              key={index} 
              className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-red-100 text-red-800 font-medium text-sm"
            >
              {num}
            </span>
          ))}
          {specialNumbers.split(',').map((num, index) => (
            <span
              key={`special-${index}`}
              className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-blue-100 text-blue-800 font-medium text-sm"
            >
              {num}
            </span>
          ))}
        </div>
      );
    }
    
    // 如果没有特殊格式，只显示所有号码
    return (
      <div className="flex flex-wrap gap-1 justify-center">
        {recommendation.numbers.split(',').map((num, index) => (
          <span
            key={index}
            className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-800 font-medium text-sm"
          >
            {num}
          </span>
        ))}
      </div>
    );
  };

  return (
    <Card className={`w-full ${isDrawn ? 'bg-white' : 'bg-blue-50'}`}>
      <CardHeader className="pb-2">
        <div className="flex justify-between items-center">
          <CardTitle className="text-lg">期号: {recommendation.drawNumber}</CardTitle>
          <Badge variant={isDrawn ? 'outline' : 'secondary'}>
            {isDrawn ? '已开奖' : '未开奖'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-2">
        {renderNumbers()}
        
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="font-semibold">类型:</span> {lotteryTypeName}
          </div>
          <div>
            <span className="font-semibold">开奖日期:</span> {formatDate(recommendation.drawTime)}
          </div>
          
          {isDrawn && (
            <>
              <div>
                <span className="font-semibold">中奖状态:</span>{' '}
                <Badge variant={recommendation.winStatus ? 'success' : 'destructive'}>
                  {recommendation.winStatus || '未中奖'}
                </Badge>
              </div>
              <div>
                <span className="font-semibold">中奖金额:</span>{' '}
                {recommendation.winAmount > 0 ? `¥${recommendation.winAmount}` : '0'}
              </div>
              {recommendation.drawResult && (
                <div className="col-span-2">
                  <span className="font-semibold">开奖结果:</span> {recommendation.drawResult}
                </div>
              )}
            </>
          )}
        </div>
      </CardContent>
      <CardFooter className="flex justify-between border-t pt-4">
        <div className="flex items-center space-x-2">
          <Switch
            id={`purchase-${recommendation.id}`}
            checked={recommendation.isPurchased}
            onCheckedChange={handlePurchaseToggle}
            disabled={isPurchasing}
          />
          <Label htmlFor={`purchase-${recommendation.id}`}>已购买</Label>
        </div>
        <div className="text-xs text-gray-500">
          推荐时间: {new Date(recommendation.createdAt).toLocaleString('zh-CN')}
        </div>
      </CardFooter>
    </Card>
  );
}