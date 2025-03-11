import { createContext, useContext, useState } from 'react'
import { useNavigate } from 'react-router-dom'

interface AuthContextType {
    token: string | null
    setToken: (token: string | null) => void
    logout: () => void
    isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [token, setTokenState] = useState<string | null>(() => {
        return localStorage.getItem('token')
    })
    const navigate = useNavigate()

    const setToken = (newToken: string | null) => {
        setTokenState(newToken)
        if (newToken) {
            localStorage.setItem('token', newToken)
        } else {
            localStorage.removeItem('token')
        }
    }

    const logout = () => {
        setToken(null)
        navigate('/login')
    }

    const value = {
        token,
        setToken,
        logout,
        isAuthenticated: !!token,
    }

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth() {
    const context = useContext(AuthContext)
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider')
    }
    return context
}