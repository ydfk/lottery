import { LogOut, User2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
  return (
    <div className="space-y-4">
      <section className="rounded-[1.8rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,64,175,0.9)_58%,rgba(12,74,110,0.88))] p-5 text-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <Badge className="bg-white/15 text-white hover:bg-white/20">看板</Badge>
              <span className="text-xs text-white/65">账户与统计</span>
            </div>
            <div className="mt-4 flex items-center gap-3">
              <div className="flex size-12 shrink-0 items-center justify-center rounded-2xl bg-white/10">
                <User2 className="size-5" />
              </div>
              <div className="min-w-0">
                <p className="truncate text-lg font-semibold">{currentUser.username}</p>
                <p className="mt-1 text-sm text-white/70">已登录</p>
              </div>
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

        <Card className="mt-5 border-white/10 bg-white/10 text-white backdrop-blur">
          <CardContent className="grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-4">
            <div className="rounded-[1.25rem] bg-white/8 px-4 py-3">
              <p className="text-xs text-white/65">累计花费</p>
              <p className="mt-2 text-xl font-semibold">{formatCurrency(dashboard?.stats.totalCost ?? 0)}</p>
            </div>
            <div className="rounded-[1.25rem] bg-white/8 px-4 py-3">
              <p className="text-xs text-white/65">累计中奖</p>
              <p className="mt-2 text-xl font-semibold">{formatCurrency(dashboard?.stats.totalPrize ?? 0)}</p>
            </div>
            <div className="rounded-[1.25rem] bg-white/8 px-4 py-3">
              <p className="text-xs text-white/65">票据数量</p>
              <p className="mt-2 text-xl font-semibold">{String(dashboard?.stats.totalTickets ?? 0)}</p>
            </div>
            <div className="rounded-[1.25rem] bg-white/8 px-4 py-3">
              <p className="text-xs text-white/65">中奖票数</p>
              <p className="mt-2 text-xl font-semibold">{String(dashboard?.stats.wonTickets ?? 0)}</p>
            </div>
          </CardContent>
        </Card>
      </section>
    </div>
  );
}
