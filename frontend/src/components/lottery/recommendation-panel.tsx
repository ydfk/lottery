import { useDeferredValue, useEffect, useMemo, useRef, useState, useTransition, type MouseEvent } from "react";
import { ArrowDownWideNarrow, ChevronDown, ScanSearch, SlidersHorizontal } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DetailSheet } from "@/components/lottery/detail-sheet";
import { HitNumberBalls, NumberBalls } from "@/components/lottery/number-balls";
import { RecommendationStealthSheet } from "@/components/lottery/recommendation-stealth-sheet";
import { TicketCard } from "@/components/lottery/ticket-card";
import {
  formatLotteryDateTime,
  formatLotteryDrawDate,
  formatLotteryIssue,
  getLotteryDisplayName,
  lotteryDisplayOptions,
} from "@/lib/lottery-display";
import type { Recommendation, RecommendationFilters } from "@/types/lottery";

interface RecommendationPanelProps {
  recommendations: Recommendation[];
  filters: RecommendationFilters;
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  total: number;
  selectedRecommendation: Recommendation | null;
  detailPending: boolean;
  deletePending: boolean;
  onFiltersChange: (filters: RecommendationFilters) => void;
  onLoadMore: () => void;
  onSelectRecommendation: (recommendationId: string | null) => void;
  onRecordPurchase: (recommendation: Recommendation) => void;
  onDeleteRecommendation: (recommendation: Recommendation) => void;
}

const statusFilterOptions = [
  { value: "", label: "全部" },
  { value: "pending", label: "待开奖" },
  { value: "won", label: "已中奖" },
  { value: "not_won", label: "未中奖" },
];

const sortOptions = [
  { value: "latest", label: "最新生成" },
  { value: "oldest", label: "最早生成" },
  { value: "draw_latest", label: "最新开奖" },
  { value: "draw_oldest", label: "最早开奖" },
  { value: "prize_high", label: "中奖金额" },
];

function formatCurrency(value: number) {
  return `¥ ${value.toFixed(2)}`;
}

function getWinningCount(recommendation: Recommendation) {
  return recommendation.winningCount ?? recommendation.entries.filter((entry) => entry.isWinning).length;
}

function getRecommendationStatus(recommendation: Recommendation) {
  if (!recommendation.checkedAt) {
    return {
      label: "待开奖",
      className: "bg-amber-100 text-amber-700 hover:bg-amber-100",
    };
  }

  if (getWinningCount(recommendation) > 0) {
    return {
      label: "已中奖",
      className: "bg-emerald-100 text-emerald-700 hover:bg-emerald-100",
    };
  }

  return {
    label: "未中奖",
    className: "bg-rose-100 text-rose-700 hover:bg-rose-100",
  };
}

function getRecommendationStatusText(recommendation: Recommendation) {
  const status = getRecommendationStatus(recommendation);
  if (!recommendation.checkedAt) {
    return status.label;
  }
  if ((recommendation.prizeAmount || 0) > 0) {
    return `${status.label} · ${formatCurrency(recommendation.prizeAmount || 0)}`;
  }
  return status.label;
}

function getRecommendationPurchaseText(recommendation: Recommendation) {
  const purchasedCount = recommendation.purchasedCount || 0;
  if (purchasedCount <= 0) {
    return "未购买";
  }
  if (purchasedCount === 1) {
    return "已购买 1 次";
  }
  return `已购买 ${purchasedCount} 次`;
}

function RecommendationCard(props: {
  recommendation: Recommendation;
  onSelectRecommendation: (recommendationId: string) => void;
  onOpenStealth: (recommendationId: string) => void;
  onRecordPurchase: (recommendation: Recommendation) => void;
}) {
  const { recommendation, onSelectRecommendation, onOpenStealth, onRecordPurchase } = props;
  const status = getRecommendationStatus(recommendation);

  return (
    <div
      role="button"
      tabIndex={0}
      className="rounded-[1.75rem] border border-white/70 bg-white/88 p-5 text-left shadow-[0_16px_40px_rgba(15,23,42,0.08)] transition hover:-translate-y-0.5 hover:shadow-[0_20px_48px_rgba(15,23,42,0.12)]"
      onClick={() => onSelectRecommendation(recommendation.id)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onSelectRecommendation(recommendation.id);
        }
      }}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{getLotteryDisplayName(recommendation.lotteryCode)}</Badge>
          {recommendation.isPurchased ? (
            <Badge variant="secondary">{getRecommendationPurchaseText(recommendation)}</Badge>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="inline-flex h-8 items-center rounded-full border border-slate-200 bg-white px-3 text-xs font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
            onClick={(event) => {
              event.stopPropagation();
              onRecordPurchase(recommendation);
            }}
          >
            {recommendation.isPurchased ? "续购" : "购买"}
          </button>
          <button
            type="button"
            className="inline-flex h-8 items-center gap-1 rounded-full border border-slate-200 bg-white px-3 text-xs font-medium text-slate-600 transition hover:border-slate-300 hover:text-slate-900"
            onClick={(event) => {
              event.stopPropagation();
              onOpenStealth(recommendation.id);
            }}
          >
            <ScanSearch className="size-3.5" />
            隐览
          </button>
        </div>
      </div>

      <div className="mt-3 space-y-1.5 text-xs text-slate-500">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <span>第 {formatLotteryIssue(recommendation.lotteryCode, recommendation.issue)} 期</span>
          {recommendation.drawDate ? <span>{formatLotteryDrawDate(recommendation.drawDate)} 开奖</span> : null}
        </div>
        <div className="flex items-center justify-between gap-3">
          <span className="min-w-0 truncate">
            {recommendation.createdAt ? `${formatLotteryDateTime(recommendation.createdAt)} 生成` : ""}
          </span>
          <span className={`shrink-0 rounded-full px-2 py-0.5 ${status.className}`}>{status.label}</span>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 rounded-[1.1rem] bg-slate-50 px-4 py-3 text-sm">
        <span className="font-medium text-slate-900">{getRecommendationStatusText(recommendation)}</span>
        <span className="text-slate-500">{getRecommendationPurchaseText(recommendation)}</span>
        {recommendation.checkedAt ? (
          <span className="text-slate-500">命中 {getWinningCount(recommendation)} 注</span>
        ) : null}
      </div>

      <div className="mt-4 space-y-3">
        {recommendation.entries.map((entry) => (
          <div key={entry.id} className="rounded-[1.35rem] bg-slate-50 px-4 py-3">
            <NumberBalls redNumbers={entry.redNumbers} blueNumbers={entry.blueNumbers} compact />
          </div>
        ))}
      </div>
    </div>
  );
}

function RecommendationSection(props: {
  title: string;
  items: Recommendation[];
  onSelectRecommendation: (recommendationId: string) => void;
  onOpenStealth: (recommendationId: string) => void;
  onRecordPurchase: (recommendation: Recommendation) => void;
}) {
  const { title, items, onSelectRecommendation, onOpenStealth, onRecordPurchase } = props;

  if (items.length === 0) {
    return null;
  }

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-sm font-semibold text-slate-800">{title}</h3>
        <span className="text-xs text-slate-400">{items.length} 条</span>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {items.map((recommendation) => (
          <RecommendationCard
            key={recommendation.id}
            recommendation={recommendation}
            onSelectRecommendation={onSelectRecommendation}
            onOpenStealth={onOpenStealth}
            onRecordPurchase={onRecordPurchase}
          />
        ))}
      </div>
    </section>
  );
}

export function RecommendationPanel(props: RecommendationPanelProps) {
  const {
    recommendations,
    filters,
    loading,
    loadingMore,
    hasMore,
    total,
    selectedRecommendation,
    detailPending,
    deletePending,
    onFiltersChange,
    onLoadMore,
    onSelectRecommendation,
    onRecordPurchase,
    onDeleteRecommendation,
  } = props;
  const loadMoreTriggerRef = useRef<HTMLDivElement | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [stealthRecommendationId, setStealthRecommendationId] = useState<string | null>(null);
  const [, startTransition] = useTransition();
  const deferredFilters = useDeferredValue(filters);
  const hasActiveFilters = Boolean(
    deferredFilters.lotteryCode || deferredFilters.status || deferredFilters.sort !== "latest"
  );
  const activeFilterLabels = useMemo(() => {
    const labels: string[] = [];
    if (deferredFilters.lotteryCode) {
      labels.push(getLotteryDisplayName(deferredFilters.lotteryCode));
    }
    if (deferredFilters.status) {
      labels.push(statusFilterOptions.find((item) => item.value === deferredFilters.status)?.label || deferredFilters.status);
    }
    if (deferredFilters.sort && deferredFilters.sort !== "latest") {
      const sortLabel = sortOptions.find((item) => item.value === deferredFilters.sort)?.label;
      if (sortLabel) {
        labels.push(sortLabel);
      }
    }
    return labels;
  }, [deferredFilters]);

  const pendingRecommendations = recommendations.filter((item) => !item.checkedAt);
  const checkedRecommendations = recommendations.filter((item) => item.checkedAt);
  const stealthRecommendation = useMemo(
    () => recommendations.find((item) => item.id === stealthRecommendationId) ?? null,
    [recommendations, stealthRecommendationId]
  );

  useEffect(() => {
    const target = loadMoreTriggerRef.current;
    if (!target || !hasMore || loading || loadingMore) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          onLoadMore();
        }
      },
      {
        rootMargin: "160px 0px",
      }
    );

    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, loading, loadingMore, onLoadMore]);

  useEffect(() => {
    if (hasActiveFilters) {
      setFiltersOpen(true);
    }
  }, [hasActiveFilters]);

  function updateFilters(nextFilters: RecommendationFilters) {
    startTransition(() => {
      onFiltersChange(nextFilters);
    });
  }

  function handleClearFilters(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation();
    updateFilters({
      lotteryCode: "",
      status: "",
      sort: "latest",
    });
  }

  return (
    <>
      <div className="space-y-6">
        <section className="rounded-[1.6rem] border border-white/60 bg-white/88 p-4 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Badge className="bg-slate-900 text-white hover:bg-slate-900">推荐</Badge>
              <h2 className="text-base font-semibold text-slate-950">列表</h2>
              <span className="text-xs text-slate-500">{total} 条</span>
            </div>
            <div className="flex items-center gap-2">
              {hasActiveFilters && !filtersOpen ? (
                <div className="hidden items-center gap-2 sm:flex">
                  {activeFilterLabels.map((label) => (
                    <span
                      key={label}
                      className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-600"
                    >
                      {label}
                    </span>
                  ))}
                </div>
              ) : null}
              <button
                type="button"
                className="flex items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
                onClick={() => setFiltersOpen((value) => !value)}
              >
                <SlidersHorizontal className="size-4 text-slate-400" />
                筛选
                <ChevronDown
                  className={`size-4 text-slate-400 transition ${filtersOpen ? "rotate-180" : ""}`}
                />
              </button>
            </div>
          </div>

          {filtersOpen ? (
            <div className="mt-4 space-y-3">
              <div className="grid gap-3 sm:grid-cols-[1fr_1fr]">
                <select
                  className="h-11 rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700 outline-none transition focus:border-slate-400"
                  value={filters.lotteryCode}
                  onChange={(event) => updateFilters({ ...filters, lotteryCode: event.target.value })}
                >
                  <option value="">全部彩种</option>
                  {lotteryDisplayOptions.map((item) => (
                    <option key={item.code} value={item.code}>
                      {item.name}
                    </option>
                  ))}
                </select>

                <label className="flex h-11 items-center gap-3 rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700">
                  <ArrowDownWideNarrow className="size-4 text-slate-400" />
                  <select
                    className="w-full bg-transparent outline-none"
                    value={filters.sort}
                    onChange={(event) => updateFilters({ ...filters, sort: event.target.value })}
                  >
                    {sortOptions.map((item) => (
                      <option key={item.value} value={item.value}>
                        {item.label}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="flex items-center justify-between gap-3">
                <div className="flex gap-2 overflow-x-auto pb-1">
                  {statusFilterOptions.map((item) => {
                    const active = filters.status === item.value;
                    return (
                      <button
                        key={item.value || "all"}
                        type="button"
                        className={`shrink-0 rounded-full px-4 py-2 text-sm font-medium transition ${
                          active
                            ? "bg-slate-900 text-white"
                            : "border border-slate-200 bg-white text-slate-600"
                        }`}
                        onClick={() => updateFilters({ ...filters, status: item.value })}
                      >
                        {item.label}
                      </button>
                    );
                  })}
                </div>

                {hasActiveFilters ? (
                  <button
                    type="button"
                    className="shrink-0 text-sm text-slate-500 transition hover:text-slate-900"
                    onClick={handleClearFilters}
                  >
                    清空
                  </button>
                ) : null}
              </div>
            </div>
          ) : null}
        </section>

        {loading ? (
          <Card className="border-white/60 bg-white/85 backdrop-blur">
            <CardContent className="py-14 text-center text-sm text-slate-500">推荐加载中...</CardContent>
          </Card>
        ) : recommendations.length > 0 ? (
          <>
            {filters.status ? (
              <RecommendationSection
                title={statusFilterOptions.find((item) => item.value === filters.status)?.label || "推荐"}
                items={recommendations}
                onSelectRecommendation={(recommendationId) => onSelectRecommendation(recommendationId)}
                onOpenStealth={(recommendationId) => setStealthRecommendationId(recommendationId)}
                onRecordPurchase={onRecordPurchase}
              />
            ) : (
              <div className="space-y-6">
                <RecommendationSection
                  title="待开奖"
                  items={pendingRecommendations}
                  onSelectRecommendation={(recommendationId) => onSelectRecommendation(recommendationId)}
                  onOpenStealth={(recommendationId) => setStealthRecommendationId(recommendationId)}
                  onRecordPurchase={onRecordPurchase}
                />
                <RecommendationSection
                  title="已开奖"
                  items={checkedRecommendations}
                  onSelectRecommendation={(recommendationId) => onSelectRecommendation(recommendationId)}
                  onOpenStealth={(recommendationId) => setStealthRecommendationId(recommendationId)}
                  onRecordPurchase={onRecordPurchase}
                />
              </div>
            )}

            {hasMore ? (
              <div className="space-y-3">
                <div ref={loadMoreTriggerRef} className="h-4" />
                <div className="flex justify-center">
                  <Button
                    type="button"
                    variant="secondary"
                    className="h-11 rounded-2xl px-6"
                    disabled={loadingMore}
                    onClick={onLoadMore}
                  >
                    {loadingMore ? "加载中..." : "加载更多"}
                  </Button>
                </div>
              </div>
            ) : null}
          </>
        ) : (
          <Card className="border-white/60 bg-white/85 backdrop-blur">
            <CardContent className="py-14 text-center text-sm text-slate-500">
              当前筛选条件下没有推荐记录。
            </CardContent>
          </Card>
        )}
      </div>

      <DetailSheet
        open={Boolean(selectedRecommendation)}
        title="推荐详情"
        rightAction={
          selectedRecommendation ? (
            <>
              <Button
                type="button"
                variant="ghost"
                className="h-11 rounded-full px-3 text-sm text-slate-700"
                disabled={deletePending}
                onClick={() => onRecordPurchase(selectedRecommendation)}
              >
                {selectedRecommendation.isPurchased ? "续记" : "购买"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="h-11 rounded-full px-3 text-sm text-rose-600 hover:bg-rose-50 hover:text-rose-700"
                disabled={deletePending}
                onClick={() => onDeleteRecommendation(selectedRecommendation)}
              >
                {deletePending ? "删除中" : "删除"}
              </Button>
            </>
          ) : undefined
        }
        onOpenChange={(open) => onSelectRecommendation(open ? selectedRecommendation?.id ?? null : null)}
      >
        {selectedRecommendation ? (
          <div className="space-y-6">
            <div className="space-y-3">
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="secondary">{getLotteryDisplayName(selectedRecommendation.lotteryCode)}</Badge>
                <Badge variant="secondary">
                  第 {formatLotteryIssue(selectedRecommendation.lotteryCode, selectedRecommendation.issue)} 期
                </Badge>
                {selectedRecommendation.drawDate ? (
                  <Badge variant="secondary">{formatLotteryDrawDate(selectedRecommendation.drawDate)} 开奖</Badge>
                ) : null}
                {selectedRecommendation.createdAt ? (
                  <Badge variant="secondary">{formatLotteryDateTime(selectedRecommendation.createdAt)} 生成</Badge>
                ) : null}
                <Badge className={getRecommendationStatus(selectedRecommendation).className}>
                  {getRecommendationStatus(selectedRecommendation).label}
                </Badge>
                {selectedRecommendation.isPurchased ? (
                  <Badge variant="secondary">{getRecommendationPurchaseText(selectedRecommendation)}</Badge>
                ) : null}
              </div>
              {selectedRecommendation.summary ? (
                <h2 className="text-2xl font-semibold text-slate-950">
                  {selectedRecommendation.summary}
                </h2>
              ) : null}
            </div>

            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-[1.35rem] bg-slate-50 px-4 py-3 text-sm text-slate-600">
              <span className="font-medium text-slate-900">{getRecommendationStatusText(selectedRecommendation)}</span>
              <span>共 {selectedRecommendation.entryCount || selectedRecommendation.entries.length} 注</span>
              <span>命中 {selectedRecommendation.winningCount || 0} 注</span>
              <span>{getRecommendationPurchaseText(selectedRecommendation)}</span>
              <span>生成于 {formatLotteryDateTime(selectedRecommendation.createdAt)}</span>
              <span className="font-medium text-slate-900">
                总奖金 {formatCurrency(selectedRecommendation.prizeAmount || 0)}
              </span>
            </div>

            <div className="rounded-[1.35rem] border border-slate-200 bg-white px-4 py-4">
              <div className="flex items-center justify-between gap-3">
                <span className="text-sm font-medium text-slate-700">开奖号码</span>
                {selectedRecommendation.checkedAt ? (
                  <span className="text-xs text-slate-500">已开奖</span>
                ) : (
                  <span className="text-xs text-slate-400">待开奖</span>
                )}
              </div>
              <div className="mt-3">
                {selectedRecommendation.drawRedNumbers && selectedRecommendation.drawBlueNumbers ? (
                  <NumberBalls
                    redNumbers={selectedRecommendation.drawRedNumbers}
                    blueNumbers={selectedRecommendation.drawBlueNumbers}
                    compact
                  />
                ) : (
                  <p className="text-sm text-slate-400">暂未同步到开奖号码</p>
                )}
              </div>
            </div>

            <div className="space-y-3">
              {selectedRecommendation.entries.map((entry) => (
                <div
                  key={entry.id}
                  className="rounded-[1.35rem] border border-slate-200 bg-white px-4 py-3"
                >
                  <div className="flex items-center justify-between gap-3 text-sm font-medium text-slate-600">
                    <span>第 {entry.sequence} 注</span>
                    <span>{entry.prizeName || (selectedRecommendation.checkedAt ? "未命中奖级" : "待开奖")}</span>
                  </div>
                  <div className="mt-3">
                    <HitNumberBalls
                      redNumbers={entry.redNumbers}
                      blueNumbers={entry.blueNumbers}
                      drawRedNumbers={selectedRecommendation.drawRedNumbers}
                      drawBlueNumbers={selectedRecommendation.drawBlueNumbers}
                      compact
                    />
                  </div>
                  <div className="mt-3 flex items-center justify-between text-sm">
                    <span className="text-slate-500">{entry.matchSummary || "待开奖"}</span>
                    <span className="font-semibold text-slate-900">
                      {formatCurrency(entry.prizeAmount || 0)}
                    </span>
                  </div>
                  {entry.reason ? (
                    <p className="mt-2 text-xs leading-5 text-slate-500">{entry.reason}</p>
                  ) : null}
                </div>
              ))}
            </div>

            {detailPending ? (
              <Card className="border-white/60 bg-slate-50">
                <CardContent className="py-8 text-center text-sm text-slate-500">
                  正在加载推荐详情...
                </CardContent>
              </Card>
            ) : selectedRecommendation.purchasedTickets && selectedRecommendation.purchasedTickets.length > 0 ? (
              <div className="space-y-3">
                <h3 className="text-base font-semibold text-slate-900">购买记录</h3>
                <div className="space-y-4">
                  {selectedRecommendation.purchasedTickets.map((ticket) => (
                    <TicketCard key={ticket.id} ticket={ticket} />
                  ))}
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
      </DetailSheet>

      <RecommendationStealthSheet
        open={Boolean(stealthRecommendation)}
        recommendation={stealthRecommendation}
        onOpenChange={(open) => {
          if (!open) {
            setStealthRecommendationId(null);
          }
        }}
      />
    </>
  );
}
