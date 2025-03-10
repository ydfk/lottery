import { Recommendation, PaginatedResponse } from '../types'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export async function getRecommendations(page = 1, pageSize = 20): Promise<PaginatedResponse<Recommendation>> {
    const token = localStorage.getItem('token')
    const response = await fetch(`${BASE_URL}/api/recommendations?page=${page}&pageSize=${pageSize}`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    if (!response.ok) {
        throw new Error('Failed to fetch recommendations')
    }
    return response.json()
}