import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Label } from "../ui/label"
import { Input } from "../ui/input"
import { Button } from "../ui/button"
import { useToast } from "../ui/use-toast"
import { login } from '@/lib/api/auth'
import { useAuth } from '@/contexts/AuthContext'

export function Login() {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)
    const { toast } = useToast()
    const navigate = useNavigate()
    const { setToken } = useAuth()

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)
        try {
            const response = await login({ username, password })
            setToken(response.token)
            toast({
                title: "登录成功",
                description: "正在跳转到首页...",
            })
            navigate('/')
        } catch {
            toast({
                variant: "destructive",
                title: "登录失败",
                description: "请检查用户名和密码",
            })
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen w-full flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
            <Card className="w-full max-w-md shadow-lg">
                <CardHeader className="space-y-2">
                    <CardTitle className="text-2xl md:text-3xl font-bold text-center text-gray-800">
                        抽奖系统
                    </CardTitle>
                    <p className="text-center text-gray-600">
                        请登录您的账户
                    </p>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div className="space-y-2">
                            <Label htmlFor="username" className="text-sm font-medium">用户名</Label>
                            <Input
                                id="username"
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                required
                                className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                                placeholder="请输入用户名"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password" className="text-sm font-medium">密码</Label>
                            <Input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                                className="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                                placeholder="请输入密码"
                            />
                        </div>
                        <Button 
                            type="submit" 
                            className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2 rounded-md transition-colors"
                            disabled={loading}
                        >
                            {loading ? '登录中...' : '登录'}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}