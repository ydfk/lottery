import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from "@/components/ui/toaster"
import { AuthProvider } from '@/contexts/AuthContext'
import { Login } from '@/components/auth/Login'
import { LotteryList } from '@/components/lottery/LotteryList'
import { useAuth } from './contexts/AuthContext'

function PrivateRoute({ children }: { children: React.ReactNode }) {
    const { token } = useAuth()
    return token ? <>{children}</> : <Navigate to="/login" />
}

function AppRoutes() {
    return (
        <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/" element={
                <PrivateRoute>
                    <LotteryList />
                </PrivateRoute>
            } />
        </Routes>
    )
}

export default function App() {
    return (
        <Router>
            <AuthProvider>
                <AppRoutes />
                <Toaster />
            </AuthProvider>
        </Router>
    )
}
