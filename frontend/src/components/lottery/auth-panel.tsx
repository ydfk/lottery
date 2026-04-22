import type { FormEvent } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  LotteryDisplayModeToggle,
  type LotteryDisplayMode,
} from "@/components/lottery/display-mode-toggle";
import { BrandLogo } from "@/components/lottery/brand-logo";
import { APP_VERSION_LABEL } from "@/lib/app-meta";
import { APP_BRAND_NAME, APP_BRAND_SUBTITLE } from "@/lib/brand";
import { cn } from "@/lib/utils";

interface AuthPanelProps {
  username: string;
  password: string;
  pending: boolean;
  displayMode: LotteryDisplayMode;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onDisplayModeChange: (value: LotteryDisplayMode) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function AuthPanel(props: AuthPanelProps) {
  const {
    username,
    password,
    pending,
    displayMode,
    onUsernameChange,
    onPasswordChange,
    onDisplayModeChange,
    onSubmit,
  } = props;
  const isWebMode = displayMode === "web";

  return (
    <div
      className={cn(
        "min-h-screen px-4 py-6 sm:px-6",
        isWebMode
          ? "bg-[linear-gradient(180deg,rgba(248,250,252,0.96),rgba(226,232,240,0.88)),radial-gradient(circle_at_top_right,rgba(59,130,246,0.14),transparent_38%)]"
          : "flex items-center justify-center bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.92),rgba(226,232,240,0.9)_45%,rgba(203,213,225,1))] py-10"
      )}
    >
      <div className={cn("mx-auto w-full", isWebMode ? "max-w-6xl" : "max-w-md")}>
        {isWebMode ? (
          <div className="mb-6 flex justify-end">
            <LotteryDisplayModeToggle value={displayMode} onValueChange={onDisplayModeChange} />
          </div>
        ) : (
          <div className="mb-4 flex justify-end">
            <LotteryDisplayModeToggle
              value={displayMode}
              compact
              className="bg-white/84"
              onValueChange={onDisplayModeChange}
            />
          </div>
        )}

        <div
          className={cn(isWebMode && "grid gap-6 lg:grid-cols-[1.08fr_0.92fr] lg:items-stretch")}
        >
          {isWebMode ? (
            <section className="relative hidden overflow-hidden rounded-[2rem] border border-white/15 bg-[linear-gradient(140deg,rgba(15,23,42,0.98),rgba(30,64,175,0.94)_58%,rgba(8,47,73,0.92))] p-6 text-white shadow-[0_24px_60px_rgba(15,23,42,0.18)] lg:flex lg:flex-col lg:justify-between">
              <div className="absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.18),transparent_58%)]" />

              <div className="relative flex flex-col gap-8">
                <div className="flex items-center gap-2">
                  <Badge className="bg-white/12 text-white hover:bg-white/12">Web 展示</Badge>
                  <span className="text-sm text-white/70">
                    更适合桌面端查看推荐、录票和兑奖结果
                  </span>
                </div>

                <div className="max-w-xl">
                  <div className="flex items-center gap-4">
                    <img
                      alt={APP_BRAND_NAME}
                      className="size-14 rounded-[1.35rem] shadow-[0_14px_30px_rgba(15,23,42,0.28)]"
                      src="/lottery-logo.svg"
                    />
                    <div>
                      <p className="text-2xl font-semibold tracking-[0.18em]">{APP_BRAND_NAME}</p>
                      <p className="mt-1 text-sm text-white/70">{APP_BRAND_SUBTITLE}</p>
                    </div>
                  </div>

                  <h1 className="mt-8 text-4xl font-semibold leading-tight">
                    用更接近桌面工作台的方式管理你的每一次选号和兑奖。
                  </h1>
                  <p className="mt-4 max-w-lg text-sm leading-7 text-white/74">
                    登录后你可以随时在应用式界面和 Web 展示界面之间切换，同一套数据和功能都会保留。
                  </p>
                </div>

                <div className="grid gap-3 sm:grid-cols-3">
                  {[
                    ["推荐", "集中查看每期推荐结果"],
                    ["录票", "上传图片并识别号码"],
                    ["兑奖", "追踪历史与中奖状态"],
                  ].map(([title, description]) => (
                    <div
                      key={title}
                      className="rounded-[1.4rem] border border-white/10 bg-white/10 px-4 py-4 backdrop-blur"
                    >
                      <p className="text-sm font-semibold">{title}</p>
                      <p className="mt-2 text-xs leading-5 text-white/70">{description}</p>
                    </div>
                  ))}
                </div>
              </div>

              <p className="relative mt-8 text-sm text-white/60">{APP_VERSION_LABEL}</p>
            </section>
          ) : null}

          <div
            className={cn(
              "rounded-[2rem] p-4 sm:p-5",
              isWebMode
                ? "border border-white/60 bg-white/72 shadow-[0_24px_60px_rgba(15,23,42,0.14)] backdrop-blur"
                : "border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(37,99,235,0.9)_62%,rgba(12,74,110,0.82))] shadow-[0_24px_60px_rgba(15,23,42,0.18)]"
            )}
          >
            <Card className="border-white/50 bg-white/94 backdrop-blur">
              <CardHeader className="space-y-4">
                <div className="space-y-3">
                  <BrandLogo subtitle="你的彩票轨迹" />
                  <div>
                    <Badge className="bg-slate-900 text-white hover:bg-slate-900">登录</Badge>
                    <CardTitle className="mt-3 text-slate-900">记录每一次选号和兑奖结果</CardTitle>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <form className="space-y-4" onSubmit={onSubmit}>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-700">用户名</label>
                    <Input
                      value={username}
                      onChange={(event) => onUsernameChange(event.target.value)}
                      placeholder="请输入用户名"
                    />
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
                <p className="mt-4 text-center text-xs text-slate-400">{APP_VERSION_LABEL}</p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
