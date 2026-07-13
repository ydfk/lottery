import { memo, useState } from "react";
import { RefreshCw, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type {
  LotteryNumberTrendData,
  LotteryNumberTrendMap,
  TrendLotteryCode,
} from "@/hooks/use-lottery-number-trends";
import {
  formatLotteryDrawDate,
  formatLotteryIssue,
  getLotteryDisplayName,
} from "@/lib/lottery-display";
import { cn } from "@/lib/utils";

interface LotteryNumberTrendProps {
  trends: LotteryNumberTrendMap;
  onRetry: (lotteryCode: TrendLotteryCode) => void;
}

interface TrendConfig {
  primaryLabel: string;
  primaryMax: number;
  secondaryLabel: string;
  secondaryMax: number;
}

interface TrendNumberRowProps {
  label: string;
  meta: string;
  primaryNumbers: string;
  secondaryNumbers: string;
  config: TrendConfig;
  kind: "recommendation" | "draw";
  separateAfter?: boolean;
}

const TREND_CONFIG: Record<TrendLotteryCode, TrendConfig> = {
  ssq: {
    primaryLabel: "红球区",
    primaryMax: 33,
    secondaryLabel: "蓝球区",
    secondaryMax: 16,
  },
  dlt: {
    primaryLabel: "前区",
    primaryMax: 35,
    secondaryLabel: "后区",
    secondaryMax: 12,
  },
};

const LOTTERY_TABS: Array<{ code: TrendLotteryCode; label: string }> = [
  { code: "ssq", label: "福彩双色球" },
  { code: "dlt", label: "体彩大乐透" },
];

function createNumberRange(maximum: number) {
  return Array.from({ length: maximum }, (_, index) => `${index + 1}`.padStart(2, "0"));
}

function parseNumberSet(value: string) {
  return new Set(
    value
      .split(",")
      .map((item) => item.trim().padStart(2, "0"))
      .filter((item) => item !== "00")
  );
}

function countTrendNumbers(draws: LotteryNumberTrendData["draws"]) {
  const primaryCounts = new Map<string, number>();
  const secondaryCounts = new Map<string, number>();

  for (const draw of draws) {
    for (const number of parseNumberSet(draw.redNumbers)) {
      primaryCounts.set(number, (primaryCounts.get(number) ?? 0) + 1);
    }
    for (const number of parseNumberSet(draw.blueNumbers)) {
      secondaryCounts.set(number, (secondaryCounts.get(number) ?? 0) + 1);
    }
  }

  return { primaryCounts, secondaryCounts };
}

function TrendCell(props: {
  number: string;
  selected: boolean;
  color: "red" | "blue";
  kind: "recommendation" | "draw";
}) {
  const { number, selected, color, kind } = props;

  return (
    <td className="h-9 w-6 min-w-6 border-r border-b border-slate-100 p-0 text-center">
      {selected ? (
        <span
          aria-label={`${kind === "recommendation" ? "推荐号码" : "开奖号码"} ${number}`}
          className={cn(
            "mx-auto flex size-5 items-center justify-center rounded-full text-[10px] font-semibold tabular-nums",
            kind === "draw" && color === "red" && "bg-rose-500 text-white shadow-sm",
            kind === "draw" && color === "blue" && "bg-sky-500 text-white shadow-sm",
            kind === "recommendation" &&
              color === "red" &&
              "border border-rose-400 bg-rose-50 text-rose-700 ring-2 ring-amber-200/70",
            kind === "recommendation" &&
              color === "blue" &&
              "border border-sky-400 bg-sky-50 text-sky-700 ring-2 ring-amber-200/70"
          )}
        >
          {number}
        </span>
      ) : null}
    </td>
  );
}

const TrendNumberRow = memo(function TrendNumberRow(props: TrendNumberRowProps) {
  const { label, meta, primaryNumbers, secondaryNumbers, config, kind, separateAfter } = props;
  const primarySet = parseNumberSet(primaryNumbers);
  const secondarySet = parseNumberSet(secondaryNumbers);
  const primaryRange = createNumberRange(config.primaryMax);
  const secondaryRange = createNumberRange(config.secondaryMax);
  const isRecommendation = kind === "recommendation";

  return (
    <tr
      className={cn(
        "group transition-colors hover:bg-slate-50/80",
        isRecommendation && "bg-amber-50/70 hover:bg-amber-50",
        separateAfter && "border-b-2 border-b-amber-200"
      )}
    >
      <th
        scope="row"
        className={cn(
          "sticky left-0 z-10 h-9 w-[104px] min-w-[104px] border-r border-b border-slate-200 px-2 text-left text-[11px] font-semibold",
          isRecommendation ? "bg-amber-50 text-amber-900" : "bg-white text-slate-800"
        )}
      >
        <span className="block truncate">{label}</span>
      </th>
      <td
        className={cn(
          "sticky left-[104px] z-10 h-9 w-[80px] min-w-[80px] border-r border-b border-slate-200 px-1 text-center text-[10px] whitespace-nowrap tabular-nums",
          isRecommendation ? "bg-amber-50 text-amber-700" : "bg-white text-slate-500"
        )}
      >
        {meta}
      </td>
      {primaryRange.map((number) => (
        <TrendCell
          key={`primary-${number}`}
          number={number}
          selected={primarySet.has(number)}
          color="red"
          kind={kind}
        />
      ))}
      {secondaryRange.map((number) => (
        <TrendCell
          key={`secondary-${number}`}
          number={number}
          selected={secondarySet.has(number)}
          color="blue"
          kind={kind}
        />
      ))}
    </tr>
  );
});

function TrendCountRow(props: {
  config: TrendConfig;
  drawCount: number;
  primaryCounts: Map<string, number>;
  secondaryCounts: Map<string, number>;
}) {
  const { config, drawCount, primaryCounts, secondaryCounts } = props;
  const primaryRange = createNumberRange(config.primaryMax);
  const secondaryRange = createNumberRange(config.secondaryMax);

  return (
    <tr className="border-b-2 border-b-slate-300 bg-slate-100/90">
      <th
        scope="row"
        className="sticky left-0 z-20 h-9 w-[104px] min-w-[104px] border-r border-b border-slate-300 bg-slate-100 px-2 text-left text-[11px] font-semibold text-slate-800"
      >
        开奖次数
      </th>
      <td className="sticky left-[104px] z-20 h-9 w-[80px] min-w-[80px] border-r border-b border-slate-300 bg-slate-100 px-1 text-center text-[10px] whitespace-nowrap text-slate-500">
        近 {drawCount} 期
      </td>
      {primaryRange.map((number) => {
        const count = primaryCounts.get(number) ?? 0;
        return (
          <td
            key={`primary-count-${number}`}
            aria-label={`${config.primaryLabel}号码 ${number} 近 ${drawCount} 期开奖 ${count} 次`}
            className="h-9 w-6 min-w-6 border-r border-b border-rose-100 bg-rose-50/70 p-0 text-center text-[10px] font-semibold text-rose-700 tabular-nums"
          >
            {count}
          </td>
        );
      })}
      {secondaryRange.map((number) => {
        const count = secondaryCounts.get(number) ?? 0;
        return (
          <td
            key={`secondary-count-${number}`}
            aria-label={`${config.secondaryLabel}号码 ${number} 近 ${drawCount} 期开奖 ${count} 次`}
            className="h-9 w-6 min-w-6 border-r border-b border-sky-100 bg-sky-50/70 p-0 text-center text-[10px] font-semibold text-sky-700 tabular-nums"
          >
            {count}
          </td>
        );
      })}
    </tr>
  );
}

function TrendTable(props: { lotteryCode: TrendLotteryCode; trend: LotteryNumberTrendData }) {
  const { lotteryCode, trend } = props;
  const config = TREND_CONFIG[lotteryCode];
  const primaryRange = createNumberRange(config.primaryMax);
  const secondaryRange = createNumberRange(config.secondaryMax);
  const recommendationEntries = trend.latestRecommendation
    ? [...trend.latestRecommendation.entries].sort((left, right) => left.sequence - right.sequence)
    : [];
  const { primaryCounts, secondaryCounts } = countTrendNumbers(trend.draws);
  const hasRows = recommendationEntries.length > 0 || trend.draws.length > 0;

  if (!hasRows) {
    return (
      <Card className="border-dashed border-slate-200 bg-white/85">
        <CardContent className="py-14 text-center text-sm text-slate-500">
          当前彩种暂无可展示的开奖记录或推荐号码。
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="overflow-hidden rounded-[1.5rem] border border-slate-200 bg-white shadow-[0_18px_45px_rgba(15,23,42,0.08)]">
      <div className="max-h-[70vh] overflow-auto">
        <table
          className="w-full border-separate border-spacing-0 text-xs"
          style={{ minWidth: 184 + (config.primaryMax + config.secondaryMax) * 24 }}
        >
          <thead>
            <tr>
              <th
                rowSpan={2}
                className="sticky top-0 left-0 z-40 h-[72px] w-[104px] min-w-[104px] border-r border-b border-slate-200 bg-slate-950 px-2 text-left text-xs font-semibold text-white"
              >
                期次
              </th>
              <th
                rowSpan={2}
                className="sticky top-0 left-[104px] z-40 h-[72px] w-[80px] min-w-[80px] border-r border-b border-slate-200 bg-slate-950 px-1 text-center text-[11px] font-semibold text-white"
              >
                日期 / 期号
              </th>
              <th
                colSpan={config.primaryMax}
                className="sticky top-0 z-30 h-9 border-r border-b border-rose-200 bg-rose-50 text-center text-xs font-semibold tracking-[0.18em] text-rose-700"
              >
                {config.primaryLabel}
              </th>
              <th
                colSpan={config.secondaryMax}
                className="sticky top-0 z-30 h-9 border-b border-sky-200 bg-sky-50 text-center text-xs font-semibold tracking-[0.18em] text-sky-700"
              >
                {config.secondaryLabel}
              </th>
            </tr>
            <tr>
              {primaryRange.map((number) => (
                <th
                  key={`primary-header-${number}`}
                  scope="col"
                  className="sticky top-9 z-30 h-9 w-6 min-w-6 border-r border-b border-rose-100 bg-rose-50 text-center text-[9px] font-medium text-rose-500 tabular-nums"
                >
                  {number}
                </th>
              ))}
              {secondaryRange.map((number) => (
                <th
                  key={`secondary-header-${number}`}
                  scope="col"
                  className="sticky top-9 z-30 h-9 w-6 min-w-6 border-r border-b border-sky-100 bg-sky-50 text-center text-[9px] font-medium text-sky-500 tabular-nums"
                >
                  {number}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <TrendCountRow
              config={config}
              drawCount={trend.draws.length}
              primaryCounts={primaryCounts}
              secondaryCounts={secondaryCounts}
            />
            {recommendationEntries.map((entry, index) => (
              <TrendNumberRow
                key={`recommendation-${entry.id || entry.sequence}`}
                label={`推荐第 ${entry.sequence} 注`}
                meta={formatLotteryIssue(lotteryCode, trend.latestRecommendation?.issue)}
                primaryNumbers={entry.redNumbers}
                secondaryNumbers={entry.blueNumbers}
                config={config}
                kind="recommendation"
                separateAfter={index === recommendationEntries.length - 1}
              />
            ))}
            {trend.draws.map((draw) => (
              <TrendNumberRow
                key={draw.id}
                label={`第 ${formatLotteryIssue(lotteryCode, draw.issue)} 期`}
                meta={formatLotteryDrawDate(draw.drawDate)}
                primaryNumbers={draw.redNumbers}
                secondaryNumbers={draw.blueNumbers}
                config={config}
                kind="draw"
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TrendLegend() {
  return (
    <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-slate-500">
      <span className="inline-flex items-center gap-2">
        <span className="size-3 rounded-full bg-rose-500" />
        开奖红球 / 前区
      </span>
      <span className="inline-flex items-center gap-2">
        <span className="size-3 rounded-full bg-sky-500" />
        开奖蓝球 / 后区
      </span>
      <span className="inline-flex items-center gap-2">
        <span className="size-3 rounded-full border border-amber-400 bg-amber-50 ring-2 ring-amber-200/70" />
        当期推荐
      </span>
    </div>
  );
}

export function LotteryNumberTrend(props: LotteryNumberTrendProps) {
  const { trends, onRetry } = props;
  const [activeLotteryCode, setActiveLotteryCode] = useState<TrendLotteryCode>("ssq");
  const activeTrend = trends[activeLotteryCode];
  const hasCachedRows =
    activeTrend.draws.length > 0 || Boolean(activeTrend.latestRecommendation?.entries.length);

  return (
    <div className="space-y-4">
      <section className="overflow-hidden rounded-[1.6rem] border border-white/70 bg-[linear-gradient(135deg,rgba(255,255,255,0.96),rgba(248,250,252,0.92))] p-4 shadow-[0_16px_40px_rgba(15,23,42,0.08)]">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
          <div className="flex items-center gap-2 rounded-2xl bg-slate-100 p-1">
            {LOTTERY_TABS.map((item) => {
              const active = activeLotteryCode === item.code;
              return (
                <button
                  key={item.code}
                  type="button"
                  aria-pressed={active}
                  className={cn(
                    "rounded-xl px-4 py-2 text-sm font-medium transition",
                    active
                      ? "bg-slate-950 text-white shadow-sm"
                      : "text-slate-500 hover:bg-white hover:text-slate-900"
                  )}
                  onClick={() => setActiveLotteryCode(item.code)}
                >
                  {item.label}
                </button>
              );
            })}
          </div>

          <div className="flex flex-col gap-2 xl:items-end">
            <p className="text-xs font-medium tracking-[0.16em] text-slate-400 uppercase">
              {getLotteryDisplayName(activeLotteryCode)} · 最近 50 期
            </p>
            <TrendLegend />
          </div>
        </div>
      </section>

      {activeTrend.error ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-[1.3rem] border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          <div className="flex items-center gap-2">
            <Sparkles className="size-4" />
            <span>{activeTrend.error}</span>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="rounded-full border-amber-300 bg-white text-amber-900 hover:bg-amber-100"
            disabled={activeTrend.loading}
            onClick={() => onRetry(activeLotteryCode)}
          >
            <RefreshCw className={cn("size-3.5", activeTrend.loading && "animate-spin")} />
            重新加载
          </Button>
        </div>
      ) : null}

      {activeTrend.loading && !hasCachedRows ? (
        <Card className="border-white/60 bg-white/85 backdrop-blur">
          <CardContent className="py-14 text-center text-sm text-slate-500">
            正在加载{getLotteryDisplayName(activeLotteryCode)}号码走势...
          </CardContent>
        </Card>
      ) : activeTrend.error && !hasCachedRows ? null : (
        <TrendTable lotteryCode={activeLotteryCode} trend={activeTrend} />
      )}
    </div>
  );
}
