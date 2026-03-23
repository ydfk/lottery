import { LogOut, User2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { APP_VERSION_LABEL } from "@/lib/app-meta";
import type { AuthUser } from "@/types/auth";
import type { DashboardData } from "@/types/lottery";

interface DashboardPanelProps {
  currentUser: AuthUser;
  dashboard: DashboardData | null;
  onLogout: () => void;
}

function formatCurrency(value: number) {
  return `¥ ${value.toFixed(2)}`;
}

export function DashboardPanel({ currentUser, dashboard, onLogout }: DashboardPanelProps) {
  const stats = dashboard?.stats;

  const statItems = [
    { label: "累计花费", value: formatCurrency(stats?.totalCost ?? 0) },
    { label: "累计中奖", value: formatCurrency(stats?.totalPrize ?? 0) },
    { label: "票据数量", value: String(stats?.totalTickets ?? 0) },
    { label: "中奖票数", value: String(stats?.wonTickets ?? 0) },
    { label: "推荐数量", value: String(stats?.totalRecommendations ?? 0) },
    { label: "已购推荐", value: String(stats?.purchasedRecommendations ?? 0) },
  ];

  return (
    <div className="space-y-4">
      <section className="rounded-[1.8rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,64,175,0.9)_58%,rgba(12,74,110,0.88))] p-5 text-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Badge className="bg-white/15 text-white hover:bg-white/20">看板</Badge>
            <span className="text-xs text-white/65">账户与统计</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <div className="flex min-w-0 items-center gap-3">
              <div className="flex size-12 shrink-0 items-center justify-center rounded-2xl bg-white/10">
                <User2 className="size-5" />
              </div>
              <div className="min-w-0">
                <p className="truncate text-lg font-semibold">{currentUser.username}</p>
                <p className="mt-1 text-sm text-white/70">已登录</p>
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              className="h-10 shrink-0 rounded-2xl border border-white/20 bg-white/10 px-4 text-white hover:bg-white/15"
              onClick={onLogout}
            >
              <LogOut className="mr-2 size-4" />
              退出登录
            </Button>
          </div>
          <div className="rounded-full border border-white/8 bg-black/8 px-3 py-2">
            <p className="overflow-x-auto font-mono text-[12px] leading-5 text-white/78 [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
              {APP_VERSION_LABEL}
            </p>
          </div>
        </div>

        <Card className="mt-5 border-white/10 bg-white/10 text-white backdrop-blur">
          <CardContent className="grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
            {statItems.map((item) => (
              <div key={item.label} className="rounded-[1.25rem] bg-white/8 px-4 py-3">
                <p className="text-xs text-white/65">{item.label}</p>
                <p className="mt-2 text-xl font-semibold">{item.value}</p>
              </div>
            ))}
          </CardContent>
        </Card>
      </section>
    </div>
  );
}
