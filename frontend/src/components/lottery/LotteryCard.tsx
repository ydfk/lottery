import { Card, CardContent } from "@/components/ui/card"
import { Recommendation } from "@/lib/types"
import { format } from "date-fns"

interface LotteryCardProps {
    recommendation: Recommendation
}

export function LotteryCard({ recommendation }: LotteryCardProps) {
    const isDrawn = recommendation.drawTime !== null
    const cardClassName = `transform transition-all duration-200 hover:scale-105 ${
        isDrawn 
            ? recommendation.winStatus === 'win' 
                ? 'bg-green-50 dark:bg-green-950' 
                : 'bg-gray-50 dark:bg-gray-900'
            : 'bg-blue-50 dark:bg-blue-950'
    }`

    return (
        <Card className={cardClassName}>
            <CardContent className="p-4">
                <div className="space-y-2">
                    <div className="flex justify-between items-center">
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                            期号: {recommendation.drawNumber}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                            {isDrawn ? '已开奖' : '未开奖'}
                        </div>
                    </div>
                    
                    <div className="text-lg font-bold text-center py-2">
                        {recommendation.numbers}
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-sm">
                        <div>
                            预计开奖: {format(new Date(recommendation.expectedDrawTime), 'yyyy-MM-dd')}
                        </div>
                        {isDrawn && (
                            <>
                                <div>
                                    实际开奖: {format(new Date(recommendation.drawTime!), 'yyyy-MM-dd')}
                                </div>
                                <div>
                                    开奖结果: {recommendation.drawResult || '无'}
                                </div>
                                <div>
                                    中奖金额: {recommendation.winAmount > 0 ? `¥${recommendation.winAmount}` : '未中奖'}
                                </div>
                            </>
                        )}
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}