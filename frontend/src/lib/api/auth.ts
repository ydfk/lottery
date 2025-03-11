import { LoginRequest, LoginResponse } from '../types'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export async function login(data: LoginRequest): Promise<LoginResponse> {
    const response = await fetch(`${BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
    if (!response.ok) {
        throw new Error('Login failed')
    }
    return response.json()
}