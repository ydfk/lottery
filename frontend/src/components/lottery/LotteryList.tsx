import { useEffect, useState } from 'react'
import { useInView } from 'react-intersection-observer'
import { Recommendation } from '@/lib/types'
import { getRecommendations } from '@/lib/api/recommendations'
import { LotteryCard } from './LotteryCard'
import { Spinner } from '../ui/spinner'

export function LotteryList() {
    const [recommendations, setRecommendations] = useState<Recommendation[]>([])
    const [page, setPage] = useState(1)
    const [loading, setLoading] = useState(false)
    const [hasMore, setHasMore] = useState(true)
    const { ref, inView } = useInView()

    const loadMore = async () => {
        if (loading || !hasMore) return
        
        setLoading(true)
        try {
            const response = await getRecommendations(page)
            const newItems = response.data
            
            if (newItems.length === 0) {
                setHasMore(false)
                return
            }

            setRecommendations(prev => [...prev, ...newItems])
            setPage(prev => prev + 1)
        } catch (error) {
            console.error('Failed to load recommendations:', error)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        if (inView) {
            loadMore()
        }
    }, [inView])

    return (
        <div className="container mx-auto px-4 py-8 max-w-2xl">
            <div className="space-y-4">
                {recommendations.map((recommendation) => (
                    <LotteryCard 
                        key={recommendation.id} 
                        recommendation={recommendation} 
                    />
                ))}
                
                {loading && (
                    <div className="flex justify-center p-4">
                        <Spinner />
                    </div>
                )}
                
                {!loading && hasMore && (
                    <div ref={ref} className="h-20" />
                )}
                
                {!hasMore && recommendations.length > 0 && (
                    <div className="text-center text-gray-500 py-4">
                        没有更多数据了
                    </div>
                )}
                
                {!loading && recommendations.length === 0 && (
                    <div className="text-center text-gray-500 py-4">
                        暂无推荐数据
                    </div>
                )}
            </div>
        </div>
    )
}