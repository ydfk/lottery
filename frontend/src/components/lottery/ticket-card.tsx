import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { HitNumberBalls, NumberBalls } from "./number-balls";
import { formatLotteryIssue, getLotteryDisplayName } from "@/lib/lottery-display";
import type { Ticket } from "@/types/lottery";
import { format } from "date-fns";

const statusLabelMap: Record<string, string> = {
  pending: "待开奖",
  won: "已中奖",
  not_won: "未中奖",
};

const statusClassMap: Record<string, string> = {
  pending: "border-amber-200 bg-amber-50 text-amber-700",
  won: "border-emerald-200 bg-emerald-50 text-emerald-700",
  not_won: "border-rose-200 bg-rose-50 text-rose-700",
};

export function TicketCard({ ticket }: { ticket: Ticket }) {
  return (
    <Card className="border-white/60 bg-white/85 shadow-[0_16px_40px_rgba(15,23,42,0.08)]">
      <CardHeader className="flex flex-row items-start justify-between gap-4 pb-2">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="secondary">{getLotteryDisplayName(ticket.lotteryCode)}</Badge>
            <Badge variant="secondary">
              {ticket.recommendationId ? "推荐购买" : "手动录入"}
            </Badge>
            <span
              className={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold ${
                statusClassMap[ticket.status] || "border-slate-200 bg-slate-100 text-slate-700"
              }`}
            >
              {statusLabelMap[ticket.status] || ticket.status}
            </span>
          </div>
          <CardTitle className="mt-3 text-lg text-slate-900">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <span>
                第 {formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期
                {ticket.drawDate ? ` · ${format(new Date(ticket.drawDate), "yyyy-MM-dd")}` : ""}
              </span>
              {ticket.drawRedNumbers && ticket.drawBlueNumbers ? (
                <NumberBalls
                  redNumbers={ticket.drawRedNumbers}
                  blueNumbers={ticket.drawBlueNumbers}
                  compact
                />
              ) : (
                <span className="text-sm font-medium text-slate-400">待开奖</span>
              )}
            </div>
          </CardTitle>
          <p className="mt-1 text-sm text-slate-500">{format(new Date(ticket.purchasedAt), "yyyy-MM-dd HH:mm")}</p>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between gap-4 rounded-2xl bg-slate-50 px-4 py-3">
          <div className="min-w-0">
            <p className="text-xs text-slate-500">花费</p>
            <p className="mt-1 text-base font-semibold text-slate-900">¥ {ticket.costAmount.toFixed(2)}</p>
          </div>
          <div className="h-10 w-px bg-slate-200" />
          <div className="min-w-0 text-right">
            <p className="text-xs text-slate-500">中奖</p>
            <p
              className={`mt-1 text-base font-semibold ${
                ticket.prizeAmount > 0 ? "text-emerald-600" : "text-slate-900"
              }`}
            >
              ¥ {ticket.prizeAmount.toFixed(2)}
            </p>
          </div>
        </div>

        {ticket.recommendation ? (
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="secondary">关联推荐</Badge>
              <span className="text-sm font-medium text-slate-700">
                第 {formatLotteryIssue(ticket.lotteryCode, ticket.recommendation.issue)} 期
                {ticket.recommendation.drawDate
                  ? ` · ${format(new Date(ticket.recommendation.drawDate), "yyyy-MM-dd")}`
                  : ""}
              </span>
            </div>
            {ticket.recommendation.summary ? (
              <p className="mt-2 text-sm text-slate-500">{ticket.recommendation.summary}</p>
            ) : null}
            <div className="mt-3 space-y-2">
              {ticket.recommendation.entries.map((entry) => (
                <div key={entry.id} className="rounded-2xl bg-white px-3 py-2">
                  <NumberBalls
                    redNumbers={entry.redNumbers}
                    blueNumbers={entry.blueNumbers}
                    compact
                  />
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {ticket.imageUrl && (
          <div className="overflow-hidden rounded-2xl border border-slate-200 bg-slate-100">
            <img
              src={ticket.imageUrl}
              alt={`第 ${formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期彩票原图`}
              className="h-44 w-full object-cover"
            />
            <div className="flex items-center justify-between gap-3 px-4 py-3">
              <p className="text-sm text-slate-500">原图已保存，可随时查看</p>
              <Dialog>
                <DialogTrigger asChild>
                  <Button type="button" variant="secondary" size="sm">
                    查看原图
                  </Button>
                </DialogTrigger>
                <DialogContent className="max-w-3xl border-slate-200 bg-white p-3 sm:p-4">
                  <DialogHeader>
                    <DialogTitle>第 {formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期票据原图</DialogTitle>
                    <DialogDescription>系统保留了上传原图，便于后续复核识别结果。</DialogDescription>
                  </DialogHeader>
                  <div className="overflow-hidden rounded-2xl bg-slate-100">
                    <img
                      src={ticket.imageUrl}
                      alt={`第 ${formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期彩票原图放大预览`}
                      className="max-h-[75vh] w-full object-contain"
                    />
                  </div>
                </DialogContent>
              </Dialog>
            </div>
          </div>
        )}

        {ticket.entries.map((entry) => (
          <div key={entry.id} className="rounded-2xl bg-slate-50 p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-600">
                <span>第 {entry.sequence} 注 · {entry.multiple || 1} 倍</span>
                {entry.isAdditional && <Badge variant="secondary">追加</Badge>}
              </div>
            </div>
            <div className="mt-3">
              <HitNumberBalls
                redNumbers={entry.redNumbers}
                blueNumbers={entry.blueNumbers}
                drawRedNumbers={ticket.drawRedNumbers}
                drawBlueNumbers={ticket.drawBlueNumbers}
                compact
              />
            </div>
            <div className="mt-3 flex items-center justify-between text-sm">
              <span className="text-slate-500">{entry.prizeName || "未命中奖级"}</span>
              <span className="font-semibold text-slate-900">¥ {entry.prizeAmount.toFixed(2)}</span>
            </div>
          </div>
        ))}

        {ticket.entries.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 py-5 text-sm text-slate-500">
            这条记录仅保存了票据原图附件，未识别号码。
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
