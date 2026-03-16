import type { FormEvent } from "react";
import { ShieldCheck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

export type AuthMode = "login" | "register";

interface AuthPanelProps {
  mode: AuthMode;
  username: string;
  password: string;
  pending: boolean;
  onModeChange: (mode: AuthMode) => void;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function AuthPanel(props: AuthPanelProps) {
  const {
    mode,
    username,
    password,
    pending,
    onModeChange,
    onUsernameChange,
    onPasswordChange,
    onSubmit,
  } = props;

  return (
    <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.92),rgba(226,232,240,0.9)_45%,rgba(203,213,225,1))] px-4 py-10">
      <div className="w-full max-w-4xl overflow-hidden rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(37,99,235,0.9)_62%,rgba(22,163,74,0.78))] p-4 shadow-[0_24px_60px_rgba(15,23,42,0.18)] sm:p-6">
        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <section className="flex flex-col justify-between rounded-[1.75rem] bg-white/10 p-6 text-white backdrop-blur">
            <div>
              <Badge className="bg-white/15 text-white hover:bg-white/20">鉴权已启用</Badge>
              <h1 className="mt-4 text-3xl font-semibold tracking-tight sm:text-4xl">
                彩票管理系统
              </h1>
              <p className="mt-3 text-sm leading-6 text-white/80 sm:text-base">
                现在所有业务接口都统一挂在受保护的 `/api` 路由下。登录后才能同步开奖、拍照录票、查看推荐和浏览票据原图。
              </p>
            </div>

            <div className="mt-8 grid gap-3 text-sm text-white/80 sm:grid-cols-2">
              <div className="rounded-2xl bg-white/10 p-4">原图永久保留，支持回看和复核。</div>
              <div className="rounded-2xl bg-white/10 p-4">OCR 默认使用 PaddleOCR，不依赖大模型。</div>
            </div>
          </section>

          <Card className="border-white/50 bg-white/92 backdrop-blur">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="rounded-2xl bg-slate-900 p-3 text-white">
                  <ShieldCheck className="size-5" />
                </div>
                <div>
                  <CardTitle className="text-slate-900">{mode === "login" ? "登录" : "注册"}</CardTitle>
                  <p className="mt-1 text-sm text-slate-500">
                    {mode === "login" ? "输入账号进入系统" : "先创建账号，再自动登录"}
                  </p>
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
                  {pending ? "提交中..." : mode === "login" ? "登录并进入系统" : "注册并进入系统"}
                </Button>
              </form>

              <div className="mt-4 flex items-center justify-between rounded-2xl bg-slate-100 px-4 py-3 text-sm">
                <span className="text-slate-500">
                  {mode === "login" ? "还没有账号？" : "已经有账号了？"}
                </span>
                <button
                  type="button"
                  className="font-medium text-slate-900"
                  onClick={() => onModeChange(mode === "login" ? "register" : "login")}
                >
                  {mode === "login" ? "去注册" : "去登录"}
                </button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
