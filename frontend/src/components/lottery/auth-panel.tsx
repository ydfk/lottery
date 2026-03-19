import type { FormEvent } from "react";
import { ShieldCheck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

interface AuthPanelProps {
  username: string;
  password: string;
  pending: boolean;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function AuthPanel(props: AuthPanelProps) {
  const {
    username,
    password,
    pending,
    onUsernameChange,
    onPasswordChange,
    onSubmit,
  } = props;

  return (
    <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.92),rgba(226,232,240,0.9)_45%,rgba(203,213,225,1))] px-4 py-10">
      <div className="w-full max-w-md rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(37,99,235,0.9)_62%,rgba(12,74,110,0.82))] p-4 shadow-[0_24px_60px_rgba(15,23,42,0.18)] sm:p-5">
        <Card className="border-white/50 bg-white/94 backdrop-blur">
          <CardHeader className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="rounded-2xl bg-slate-900 p-3 text-white">
                <ShieldCheck className="size-5" />
              </div>
              <div>
                <Badge className="bg-slate-900 text-white hover:bg-slate-900">登录</Badge>
                <CardTitle className="mt-3 text-slate-900">彩票管理系统</CardTitle>
                <p className="mt-1 text-sm text-slate-500">管理票据、推荐号码和中奖记录</p>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={onSubmit}>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">用户名</label>
                <Input value={username} onChange={(event) => onUsernameChange(event.target.value)} placeholder="请输入用户名" />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">密码</label>
                <Input
                  type="password"
                  value={password}
                  onChange={(event) => onPasswordChange(event.target.value)}
                  placeholder="请输入密码"
                />
              </div>
              <Button className="h-11 w-full rounded-2xl" disabled={pending}>
                {pending ? "登录中..." : "登录"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
