import { format } from "date-fns";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { NumberBalls } from "@/components/lottery/number-balls";
import { TicketCard } from "@/components/lottery/ticket-card";
import { formatLotteryIssue, getLotteryDisplayName } from "@/lib/lottery-display";
import type { Ticket } from "@/types/lottery";

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

interface HistoryPanelProps {
  tickets: Ticket[];
  selectedTicket: Ticket | null;
  recheckPending: boolean;
  onSelectTicket: (ticket: Ticket | null) => void;
  onRecheckTicket: (ticketId: string) => void;
}

export function HistoryPanel(props: HistoryPanelProps) {
  const { tickets, selectedTicket, recheckPending, onSelectTicket, onRecheckTicket } = props;

  return (
    <div className="space-y-6">
      <section className="rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(255,255,255,0.96),rgba(248,250,252,0.98)_45%,rgba(226,232,240,1))] p-6 shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
        <Badge className="bg-slate-900 text-white hover:bg-slate-900">历史</Badge>
        <h2 className="mt-4 text-3xl font-semibold tracking-tight text-slate-950">
          已录入票据
        </h2>
      </section>

      {tickets.length > 0 ? (
        <div className="grid gap-4 lg:grid-cols-2">
          {tickets.map((ticket) => (
            <button
              key={ticket.id}
              type="button"
              className="text-left"
              onClick={() => onSelectTicket(ticket)}
            >
              <Card className="border-white/60 bg-white/85 shadow-[0_16px_40px_rgba(15,23,42,0.08)] transition hover:-translate-y-0.5 hover:shadow-[0_20px_48px_rgba(15,23,42,0.12)]">
                <CardContent className="p-5">
                  <div>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">{getLotteryDisplayName(ticket.lotteryCode)}</Badge>
                      <span
                        className={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold ${
                          statusClassMap[ticket.status] || "border-slate-200 bg-slate-100 text-slate-700"
                        }`}
                      >
                        {statusLabelMap[ticket.status] || ticket.status}
                      </span>
                    </div>
                    <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
                      <p className="text-lg font-semibold text-slate-900">
                        第 {formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期
                        {ticket.drawDate ? ` · ${format(new Date(ticket.drawDate), "yyyy-MM-dd")}` : ""}
                      </p>
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
                    <p className="mt-2 text-sm text-slate-500">
                      {format(new Date(ticket.purchasedAt), "yyyy-MM-dd HH:mm")}
                    </p>
                  </div>

                  <div className="mt-4 flex items-center justify-between gap-4 rounded-2xl bg-slate-50 px-4 py-3">
                    <div className="min-w-0">
                      <p className="text-xs text-slate-500">花费</p>
                      <p className="mt-1 text-base font-semibold text-slate-900">
                        ¥ {ticket.costAmount.toFixed(2)}
                      </p>
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
                </CardContent>
              </Card>
            </button>
          ))}
        </div>
      ) : (
        <Card className="border-white/60 bg-white/85 backdrop-blur">
          <CardContent className="py-14 text-center text-sm text-slate-500">
            当前还没有历史记录。
          </CardContent>
        </Card>
      )}

      <Dialog open={Boolean(selectedTicket)} onOpenChange={(open) => onSelectTicket(open ? selectedTicket : null)}>
        <DialogContent className="max-h-[92vh] overflow-y-auto border-white/70 bg-[rgba(255,255,255,0.98)] p-0 sm:max-w-4xl">
          <div className="p-6 sm:p-7">
            {selectedTicket ? (
              <>
                <DialogHeader className="mb-4">
                  <div className="flex items-center justify-between gap-3">
                    <DialogTitle>票据详情</DialogTitle>
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      disabled={recheckPending}
                      onClick={() => onRecheckTicket(selectedTicket.id)}
                    >
                      {recheckPending ? "判奖中..." : "重新判奖"}
                    </Button>
                  </div>
                </DialogHeader>
                <TicketCard ticket={selectedTicket} />
              </>
            ) : null}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
