import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { NumberBalls } from "@/components/lottery/number-balls";
import { TicketCard } from "@/components/lottery/ticket-card";
import { formatLotteryIssue, getLotteryDisplayName } from "@/lib/lottery-display";
import type { Recommendation } from "@/types/lottery";

interface RecommendationPanelProps {
  recommendations: Recommendation[];
  selectedRecommendation: Recommendation | null;
  detailPending: boolean;
  onSelectRecommendation: (recommendationId: string | null) => void;
  onRecordPurchase: (recommendation: Recommendation) => void;
}

function formatCurrency(value: number) {
  return `¥ ${value.toFixed(2)}`;
}

export function RecommendationPanel(props: RecommendationPanelProps) {
  const {
    recommendations,
    selectedRecommendation,
    detailPending,
    onSelectRecommendation,
    onRecordPurchase,
  } = props;

  return (
    <>
      <div className="space-y-6">
        <section className="rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(255,255,255,0.94),rgba(241,245,249,0.96)_46%,rgba(226,232,240,1))] p-6 shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <Badge className="bg-slate-900 text-white hover:bg-slate-900">推荐</Badge>
          <h2 className="mt-4 text-3xl font-semibold tracking-tight text-slate-950">
            推荐列表
          </h2>
        </section>

        {recommendations.length > 0 ? (
          <div className="grid gap-4 lg:grid-cols-2">
            {recommendations.map((recommendation) => (
              <button
                key={recommendation.id}
                type="button"
                className="rounded-[1.75rem] border border-white/70 bg-white/88 p-5 text-left shadow-[0_16px_40px_rgba(15,23,42,0.08)] transition hover:-translate-y-0.5 hover:shadow-[0_20px_48px_rgba(15,23,42,0.12)]"
                onClick={() => onSelectRecommendation(recommendation.id)}
              >
                <div className="flex items-center justify-between gap-3">
                  <Badge variant="secondary">{getLotteryDisplayName(recommendation.lotteryCode)}</Badge>
                  <span className="text-xs text-slate-500">
                    第 {formatLotteryIssue(recommendation.lotteryCode, recommendation.issue)} 期
                  </span>
                </div>

                <div className="mt-4 space-y-4">
                  {recommendation.entries.map((entry) => (
                    <div key={entry.id} className="rounded-[1.5rem] bg-slate-50 p-4">
                      <p className="text-sm font-medium text-slate-700">推荐 {entry.sequence}</p>
                      <div className="mt-3">
                        <NumberBalls
                          redNumbers={entry.redNumbers}
                          blueNumbers={entry.blueNumbers}
                          compact
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </button>
            ))}
          </div>
        ) : (
          <Card className="border-white/60 bg-white/85 backdrop-blur">
            <CardContent className="py-14 text-center text-sm text-slate-500">
              当前还没有推荐记录。
            </CardContent>
          </Card>
        )}
      </div>

      <Dialog
        open={Boolean(selectedRecommendation)}
        onOpenChange={(open) => onSelectRecommendation(open ? selectedRecommendation?.id ?? null : null)}
      >
        <DialogContent className="max-h-[92vh] overflow-y-auto border-white/70 bg-[rgba(255,255,255,0.98)] p-0 sm:max-w-4xl">
          <div className="p-6 sm:p-7">
            {selectedRecommendation ? (
              <div className="space-y-6">
                <DialogHeader className="space-y-3 text-left">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="secondary">{getLotteryDisplayName(selectedRecommendation.lotteryCode)}</Badge>
                    <Badge variant="secondary">
                      第 {formatLotteryIssue(selectedRecommendation.lotteryCode, selectedRecommendation.issue)} 期
                    </Badge>
                  </div>
                  <DialogTitle className="text-2xl text-slate-950">
                    {selectedRecommendation.summary || "推荐详情"}
                  </DialogTitle>
                  <DialogDescription className="text-sm leading-6 text-slate-600">
                    号码之外的信息都放在这里查看。
                  </DialogDescription>
                </DialogHeader>

                <div className="grid gap-4 sm:grid-cols-3">
                  <div className="rounded-[1.5rem] bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">推荐注数</p>
                    <p className="mt-2 text-lg font-semibold text-slate-900">
                      {selectedRecommendation.entryCount || selectedRecommendation.entries.length}
                    </p>
                  </div>
                  <div className="rounded-[1.5rem] bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">中奖注数</p>
                    <p className="mt-2 text-lg font-semibold text-slate-900">
                      {selectedRecommendation.winningCount || 0}
                    </p>
                  </div>
                  <div className="rounded-[1.5rem] bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">累计奖金</p>
                    <p className="mt-2 text-lg font-semibold text-slate-900">
                      {formatCurrency(selectedRecommendation.prizeAmount || 0)}
                    </p>
                  </div>
                </div>

                <div className="space-y-3">
                  {selectedRecommendation.entries.map((entry) => (
                    <div
                      key={entry.id}
                      className="rounded-[1.5rem] border border-slate-200 bg-white p-4"
                    >
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <div>
                          <p className="text-sm font-medium text-slate-800">推荐 {entry.sequence}</p>
                          <p className="mt-1 text-xs text-slate-500">{entry.reason}</p>
                        </div>
                        <div className="text-right text-sm text-slate-500">
                          <p>命中：{entry.matchSummary || "待开奖"}</p>
                          <p className="mt-1 font-semibold text-slate-900">
                            {formatCurrency(entry.prizeAmount || 0)}
                          </p>
                        </div>
                      </div>
                      <div className="mt-4">
                        <NumberBalls redNumbers={entry.redNumbers} blueNumbers={entry.blueNumbers} />
                      </div>
                    </div>
                  ))}
                </div>

                <div className="flex flex-wrap gap-3">
                  {!selectedRecommendation.isPurchased && (
                    <Button type="button" className="rounded-2xl" onClick={() => onRecordPurchase(selectedRecommendation)}>
                      记录已购买
                    </Button>
                  )}
                </div>

                {detailPending ? (
                  <Card className="border-white/60 bg-slate-50">
                    <CardContent className="py-8 text-center text-sm text-slate-500">
                      正在加载推荐详情...
                    </CardContent>
                  </Card>
                ) : selectedRecommendation.purchasedTicket ? (
                  <div className="space-y-3">
                    <h3 className="text-base font-semibold text-slate-900">关联票据</h3>
                    <TicketCard ticket={selectedRecommendation.purchasedTicket} />
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
