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
import { NumberBalls } from "./number-balls";
import type { Ticket } from "@/types/lottery";
import { format } from "date-fns";

const statusLabelMap: Record<string, string> = {
  pending: "待开奖",
  won: "已中奖",
  not_won: "未中奖",
};

export function TicketCard({ ticket }: { ticket: Ticket }) {
  return (
    <Card className="border-white/60 bg-white/80 backdrop-blur">
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle className="text-base text-slate-900">第 {ticket.issue} 期</CardTitle>
          <p className="mt-1 text-xs text-slate-500">
            {format(new Date(ticket.purchasedAt), "yyyy-MM-dd HH:mm")}
          </p>
        </div>
        <Badge variant={ticket.status === "won" ? "default" : "secondary"}>
          {statusLabelMap[ticket.status] || ticket.status}
        </Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        {ticket.imageUrl && (
          <div className="overflow-hidden rounded-2xl border border-slate-200 bg-slate-100">
            <img src={ticket.imageUrl} alt={`第 ${ticket.issue} 期彩票原图`} className="h-44 w-full object-cover" />
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
                    <DialogTitle>第 {ticket.issue} 期票据原图</DialogTitle>
                    <DialogDescription>系统保留了上传原图，便于后续复核识别结果。</DialogDescription>
                  </DialogHeader>
                  <div className="overflow-hidden rounded-2xl bg-slate-100">
                    <img
                      src={ticket.imageUrl}
                      alt={`第 ${ticket.issue} 期彩票原图放大预览`}
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
              <span className="text-sm font-medium text-slate-600">第 {entry.sequence} 注</span>
              <span className="text-sm text-slate-500">{entry.matchSummary || "待开奖"}</span>
            </div>
            <div className="mt-3">
              <NumberBalls redNumbers={entry.redNumbers} blueNumbers={entry.blueNumbers} compact />
            </div>
            <div className="mt-3 flex items-center justify-between text-sm">
              <span className="text-slate-500">{entry.prizeName || "未命中奖级"}</span>
              <span className="font-semibold text-slate-900">¥ {entry.prizeAmount.toFixed(2)}</span>
            </div>
          </div>
        ))}

        <div className="flex items-center justify-between border-t border-dashed border-slate-200 pt-3 text-sm">
          <span className="text-slate-500">票据总金额</span>
          <span className="font-semibold text-slate-900">¥ {ticket.prizeAmount.toFixed(2)}</span>
        </div>
      </CardContent>
    </Card>
  );
}
