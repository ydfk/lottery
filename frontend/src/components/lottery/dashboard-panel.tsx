import { LogOut, User2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { SummaryCard } from "@/components/lottery/summary-card";
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
    <div className="space-y-6">
      <section className="rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(30,64,175,0.9)_58%,rgba(12,74,110,0.88))] p-6 text-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
        <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-3xl">
            <Badge className="bg-white/15 text-white hover:bg-white/20">个人看板</Badge>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight sm:text-4xl">
              账户信息与核心统计
            </h1>
          </div>

          <Button
            type="button"
            variant="ghost"
            className="h-11 rounded-2xl border border-white/20 bg-white/10 text-white hover:bg-white/15"
            onClick={onLogout}
          >
            <LogOut className="mr-2 size-4" />
            退出登录
          </Button>
        </div>

        <Card className="mt-8 border-white/10 bg-white/10 text-white backdrop-blur">
          <CardContent className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-4">
              <div className="flex size-14 items-center justify-center rounded-2xl bg-white/10">
                <User2 className="size-6" />
              </div>
              <div>
                <p className="text-lg font-semibold">{currentUser.username}</p>
                <p className="mt-1 text-sm text-white/70">已登录，可查看推荐、录票和历史记录</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <SummaryCard
          title="累计花费"
          value={formatCurrency(dashboard?.stats.totalCost ?? 0)}
          hint="已录入票据总投入"
        />
        <SummaryCard
          title="累计中奖"
          value={formatCurrency(dashboard?.stats.totalPrize ?? 0)}
          hint="已完成判奖的奖金汇总"
        />
        <SummaryCard
          title="票据数量"
          value={String(dashboard?.stats.totalTickets ?? 0)}
          hint="全部已入库票据"
        />
        <SummaryCard
          title="中奖票数"
          value={String(dashboard?.stats.wonTickets ?? 0)}
          hint="已命中奖级的票据"
        />
      </div>
    </div>
  );
}
