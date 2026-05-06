import {
  useDeferredValue,
  useEffect,
  useMemo,
  useRef,
  useState,
  useTransition,
  type MouseEvent,
} from "react";
import { ArrowDownWideNarrow, ChevronDown, SlidersHorizontal } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { NumberBalls } from "@/components/lottery/number-balls";
import {
  formatLotteryDateTime,
  formatLotteryDrawDate,
  formatLotteryIssue,
  getLotteryDisplayName,
  lotteryDisplayOptions,
} from "@/lib/lottery-display";
import type { DrawResult, DrawResultFilters } from "@/types/lottery";

interface DrawResultsPanelProps {
  items: DrawResult[];
  filters: DrawResultFilters;
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  total: number;
  onFiltersChange: (filters: DrawResultFilters) => void;
  onLoadMore: () => void;
}

const sortOptions = [
  { value: "latest", label: "最新开奖" },
  { value: "oldest", label: "最早开奖" },
];

const sourceLabelMap: Record<string, string> = {
  jisuapi: "极速数据",
};

function formatAmount(value: number) {
  return `¥ ${value.toLocaleString("zh-CN", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`;
}

function getSourceLabel(source?: string) {
  if (!source) {
    return "系统同步";
  }

  return sourceLabelMap[source] || source;
}

function DrawHistoryTable(props: { items: DrawResult[] }) {
  const { items } = props;

  return (
    <Card className="border-white/60 bg-white/85 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
      <CardContent className="p-0">
        <Table className="min-w-[1220px]">
          <TableHeader>
            <TableRow className="bg-slate-50/80">
              <TableHead className="pl-4">彩种</TableHead>
              <TableHead>期号</TableHead>
              <TableHead>开奖日期</TableHead>
              <TableHead>开奖号码</TableHead>
              <TableHead>销量</TableHead>
              <TableHead>奖池</TableHead>
              <TableHead>来源</TableHead>
              <TableHead className="pr-4">同步时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                <TableCell className="pl-4">
                  <Badge variant="secondary">{getLotteryDisplayName(item.lotteryCode)}</Badge>
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  第 {formatLotteryIssue(item.lotteryCode, item.issue)} 期
                </TableCell>
                <TableCell className="text-slate-600">
                  {formatLotteryDrawDate(item.drawDate)}
                </TableCell>
                <TableCell className="min-w-[240px]">
                  <NumberBalls
                    redNumbers={item.redNumbers}
                    blueNumbers={item.blueNumbers}
                    compact
                  />
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  {formatAmount(item.saleAmount)}
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  {formatAmount(item.prizePoolAmount)}
                </TableCell>
                <TableCell className="text-slate-600">{getSourceLabel(item.source)}</TableCell>
                <TableCell className="pr-4 text-slate-500">
                  {formatLotteryDateTime(item.createdAt)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export function DrawResultsPanel(props: DrawResultsPanelProps) {
  const { items, filters, loading, loadingMore, hasMore, total, onFiltersChange, onLoadMore } =
    props;
  const loadMoreTriggerRef = useRef<HTMLDivElement | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [, startTransition] = useTransition();
  const deferredFilters = useDeferredValue(filters);
  const hasActiveFilters = Boolean(
    deferredFilters.lotteryCode || deferredFilters.sort !== "latest"
  );
  const activeFilterLabels = useMemo(() => {
    const labels: string[] = [];
    if (deferredFilters.lotteryCode) {
      labels.push(getLotteryDisplayName(deferredFilters.lotteryCode));
    }
    if (deferredFilters.sort && deferredFilters.sort !== "latest") {
      const sortLabel = sortOptions.find((item) => item.value === deferredFilters.sort)?.label;
      if (sortLabel) {
        labels.push(sortLabel);
      }
    }
    return labels;
  }, [deferredFilters]);

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

  function updateFilters(nextFilters: DrawResultFilters) {
    startTransition(() => {
      onFiltersChange(nextFilters);
    });
  }

  function handleClearFilters(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation();
    updateFilters({
      lotteryCode: "",
      sort: "latest",
    });
  }

  return (
    <div className="space-y-6">
      <section className="rounded-[1.6rem] border border-white/60 bg-white/88 p-4 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Badge className="bg-slate-900 text-white hover:bg-slate-900">开奖</Badge>
            <h2 className="text-base font-semibold text-slate-950">历史列表</h2>
            <span className="text-xs text-slate-500">{total} 期</span>
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

            <div className="flex items-center justify-end">
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
          <CardContent className="py-14 text-center text-sm text-slate-500">
            开奖历史加载中...
          </CardContent>
        </Card>
      ) : items.length > 0 ? (
        <>
          <DrawHistoryTable items={items} />

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
            当前筛选条件下没有历史开奖记录。
          </CardContent>
        </Card>
      )}
    </div>
  );
}
