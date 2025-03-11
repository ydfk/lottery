export interface LoginRequest {
    username: string
    password: string
}

export interface LoginResponse {
    token: string
}

export interface Recommendation {
    id: number
    lotteryTypeId: number
    numbers: string
    modelName: string
    drawTime: string | null
    expectedDrawTime: string
    drawNumber: string
    isPurchased: boolean
    drawResult: string
    winStatus: string
    winAmount: number
    createdAt: string
    updatedAt: string
}

export interface PaginatedResponse<T> {
    data: T[]
    total: number
    page: number
    pageSize: number
}