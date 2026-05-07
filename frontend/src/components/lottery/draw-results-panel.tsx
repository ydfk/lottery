import {
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
  useTransition,
  type MouseEvent,
} from "react";
import { ArrowDownWideNarrow, ChevronDown, RefreshCw, SlidersHorizontal } from "lucide-react";
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
import { ScrollArea } from "@/components/ui/scroll-area";
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
  page: number;
  pageSize: number;
  total: number;
  completePendingId: string;
  syncPending: boolean;
  onFiltersChange: (filters: DrawResultFilters) => void;
  onPageChange: (page: number) => void;
  onSyncIssue: (lotteryCode: string, issue: string) => void;
  onCompleteDraw: (draw: DrawResult) => void;
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

function formatPrizeAmount(value: number) {
  return Number.isFinite(value) && value >= 0 ? formatAmount(value) : "-";
}

function hasPrizeLevel(draw: DrawResult, prizeName: string) {
  return draw.prizeDetails.some((item) => item.prizeName.includes(prizeName));
}

function formatDrawPrizeAmount(draw: DrawResult, prizeName: string, amount: number) {
  return hasPrizeLevel(draw, prizeName) ? formatPrizeAmount(amount) : "-";
}

function getSourceLabel(source?: string) {
  if (!source) {
    return "系统同步";
  }

  return sourceLabelMap[source] || source;
}

function formatCount(value: number) {
  return value > 0 ? value.toLocaleString("zh-CN") : "-";
}

function formatDetailValue(value: string | number) {
  if (typeof value === "number") {
    return value > 0 ? formatAmount(value) : "-";
  }
  return value || "-";
}

function DrawInfoItem(props: { label: string; value: string | number }) {
  const { label, value } = props;

  return (
    <div className="rounded-2xl bg-slate-50 px-4 py-3">
      <p className="text-xs text-slate-500">{label}</p>
      <p className="mt-1 text-sm font-semibold text-slate-900">{formatDetailValue(value)}</p>
    </div>
  );
}

function DrawHistoryTable(props: {
  items: DrawResult[];
  completePendingId: string;
  onCompleteDraw: (draw: DrawResult) => void;
  onSelectDraw: (draw: DrawResult) => void;
}) {
  const { items, completePendingId, onCompleteDraw, onSelectDraw } = props;

  return (
    <Card className="border-white/60 bg-white/85 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
      <CardContent className="p-0">
        <Table className="min-w-[1180px]">
          <TableHeader>
            <TableRow className="bg-slate-50/80">
              <TableHead className="pl-4">彩种</TableHead>
              <TableHead>期号</TableHead>
              <TableHead>开奖日期</TableHead>
              <TableHead>开奖号码</TableHead>
              <TableHead>一等奖</TableHead>
              <TableHead>二等奖</TableHead>
              <TableHead>来源</TableHead>
              <TableHead className="pr-4 text-right">操作</TableHead>
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
                  {formatDrawPrizeAmount(item, "一等奖", item.firstPrizeAmount)}
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  {formatDrawPrizeAmount(item, "二等奖", item.secondPrizeAmount)}
                </TableCell>
                <TableCell className="text-slate-600">{getSourceLabel(item.source)}</TableCell>
                <TableCell className="pr-4 text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-8 rounded-full border-slate-200 bg-white px-3 text-slate-700 hover:bg-slate-50"
                      disabled={completePendingId === item.id}
                      onClick={() => onCompleteDraw(item)}
                    >
                      <RefreshCw className="size-3.5" />
                      {completePendingId === item.id ? "补全中" : "补全"}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-8 rounded-full border-slate-200 bg-white px-3 text-slate-700 hover:bg-slate-50"
                      onClick={() => onSelectDraw(item)}
                    >
                      详情
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function DrawDetailDialog(props: {
  draw: DrawResult | null;
  completePendingId: string;
  onCompleteDraw: (draw: DrawResult) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const { draw, completePendingId, onCompleteDraw, onOpenChange } = props;

  return (
    <Dialog open={Boolean(draw)} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl rounded-[1.8rem] border-slate-200 bg-white p-0 shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
        <DialogHeader className="border-b border-slate-100 px-6 pb-4 pt-6">
          <DialogTitle className="text-xl text-slate-950">
            {draw
              ? `${getLotteryDisplayName(draw.lotteryCode)} 第 ${formatLotteryIssue(
                  draw.lotteryCode,
                  draw.issue
                )} 期`
              : "开奖详情"}
          </DialogTitle>
          <DialogDescription className="mt-2 text-sm text-slate-500">
            {draw ? `${formatLotteryDrawDate(draw.drawDate)} 开奖` : ""}
          </DialogDescription>
        </DialogHeader>

        {draw ? (
          <ScrollArea className="max-h-[72vh]">
            <div className="space-y-5 px-6 py-5">
              <div className="flex flex-wrap items-center justify-between gap-3 rounded-[1.4rem] border border-slate-200 bg-slate-50 px-4 py-3">
                <p className="text-sm font-medium text-slate-700">
                  可从第三方接口重新补全当前期完整信息
                </p>
                <Button
                  type="button"
                  variant="outline"
                  className="h-10 rounded-full border-slate-200 bg-white px-4 text-slate-700 hover:bg-slate-50"
                  disabled={completePendingId === draw.id}
                  onClick={() => onCompleteDraw(draw)}
                >
                  <RefreshCw className="size-4" />
                  {completePendingId === draw.id ? "补全中..." : "补全"}
                </Button>
              </div>

              <div className="flex flex-wrap items-center gap-3 rounded-[1.4rem] bg-slate-50 px-4 py-4">
                <span className="text-sm font-medium text-slate-700">开奖号码</span>
                <NumberBalls redNumbers={draw.redNumbers} blueNumbers={draw.blueNumbers} compact />
              </div>

              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <DrawInfoItem
                  label="一等奖金额"
                  value={formatDrawPrizeAmount(draw, "一等奖", draw.firstPrizeAmount)}
                />
                <DrawInfoItem
                  label="二等奖金额"
                  value={formatDrawPrizeAmount(draw, "二等奖", draw.secondPrizeAmount)}
                />
                <DrawInfoItem label="销量" value={draw.saleAmount} />
                <DrawInfoItem label="奖池" value={draw.prizePoolAmount} />
                <DrawInfoItem label="来源" value={getSourceLabel(draw.source)} />
                <DrawInfoItem label="同步时间" value={formatLotteryDateTime(draw.createdAt)} />
                <DrawInfoItem label="更新时间" value={formatLotteryDateTime(draw.updatedAt)} />
                <DrawInfoItem label="记录 ID" value={draw.id} />
              </div>

              <div className="rounded-[1.4rem] border border-slate-200 bg-white">
                <div className="border-b border-slate-100 px-4 py-3">
                  <p className="text-sm font-semibold text-slate-900">奖级明细</p>
                </div>
                <Table>
                  <TableHeader>
                    <TableRow className="bg-slate-50/80">
                      <TableHead className="pl-4">奖级</TableHead>
                      <TableHead>中奖条件</TableHead>
                      <TableHead>中奖注数</TableHead>
                      <TableHead className="pr-4">单注奖金</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {draw.prizeDetails.length > 0 ? (
                      draw.prizeDetails.map((item) => (
                        <TableRow key={item.id}>
                          <TableCell className="pl-4 font-medium text-slate-900">
                            {item.prizeName || "-"}
                          </TableCell>
                          <TableCell className="text-slate-600">{item.prizeRule || "-"}</TableCell>
                          <TableCell className="text-slate-600">
                            {formatCount(item.winnerCount)}
                          </TableCell>
                          <TableCell className="pr-4 font-medium text-slate-900">
                            {formatAmount(item.singleBonus)}
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={4} className="py-8 text-center text-sm text-slate-500">
                          暂无奖级明细
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>

              {draw.rawPayload ? (
                <div className="rounded-[1.4rem] border border-slate-200 bg-slate-950 p-4">
                  <p className="text-sm font-semibold text-white">原始数据</p>
                  <pre className="mt-3 max-h-72 overflow-auto whitespace-pre-wrap break-words text-xs leading-5 text-slate-200">
                    {draw.rawPayload}
                  </pre>
                </div>
              ) : null}
            </div>
          </ScrollArea>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

export function DrawResultsPanel(props: DrawResultsPanelProps) {
  const {
    items,
    filters,
    loading,
    page,
    pageSize,
    total,
    completePendingId,
    syncPending,
    onFiltersChange,
    onPageChange,
    onSyncIssue,
    onCompleteDraw,
  } = props;
  const [selectedDraw, setSelectedDraw] = useState<DrawResult | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [syncLotteryCode, setSyncLotteryCode] = useState("");
  const [syncIssue, setSyncIssue] = useState("");
  const [, startTransition] = useTransition();
  const deferredFilters = useDeferredValue(filters);
  const hasActiveFilters = Boolean(
    deferredFilters.lotteryCode ||
    deferredFilters.issue ||
    deferredFilters.drawDate ||
    deferredFilters.sort !== "latest"
  );
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const activeFilterLabels = useMemo(() => {
    const labels: string[] = [];
    if (deferredFilters.lotteryCode) {
      labels.push(getLotteryDisplayName(deferredFilters.lotteryCode));
    }
    if (deferredFilters.issue) {
      labels.push(`第 ${deferredFilters.issue} 期`);
    }
    if (deferredFilters.drawDate) {
      labels.push(`${deferredFilters.drawDate} 开奖`);
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
      issue: "",
      drawDate: "",
      sort: "latest",
    });
  }

  function handleSyncIssue() {
    const lotteryCode = syncLotteryCode.trim();
    const issue = syncIssue.trim();
    if (!lotteryCode || !issue || syncPending) {
      return;
    }
    onSyncIssue(lotteryCode, issue);
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

        <div className="mt-4 rounded-[1.25rem] border border-slate-200 bg-slate-50/90 p-3">
          <div className="flex flex-wrap items-center gap-3">
            <div className="min-w-[180px] flex-1">
              <select
                className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700 outline-none transition focus:border-slate-400"
                value={syncLotteryCode}
                onChange={(event) => setSyncLotteryCode(event.target.value)}
              >
                <option value="">选择彩票类型</option>
                {lotteryDisplayOptions.map((item) => (
                  <option key={item.code} value={item.code}>
                    {item.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="min-w-[180px] flex-1">
              <input
                className="h-11 w-full rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-slate-400"
                value={syncIssue}
                placeholder="输入期号，如 2026048"
                onChange={(event) => setSyncIssue(event.target.value)}
              />
            </div>
            <Button
              type="button"
              variant="secondary"
              className="h-11 rounded-2xl px-5"
              disabled={syncPending || !syncLotteryCode || !syncIssue.trim()}
              onClick={handleSyncIssue}
            >
              {syncPending ? "同步中..." : "同步开奖"}
            </Button>
          </div>
          <p className="mt-2 text-xs text-slate-500">
            手动调用第三方接口同步指定期号，已存在会覆盖，不存在会新增。
          </p>
        </div>

        {filtersOpen ? (
          <div className="mt-4 space-y-3">
            <div className="grid gap-3 md:grid-cols-[1fr_1fr_1fr_1fr]">
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

              <input
                className="h-11 rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-slate-400"
                value={filters.issue}
                placeholder="按期号筛选"
                onChange={(event) => updateFilters({ ...filters, issue: event.target.value })}
              />

              <input
                type="date"
                className="h-11 rounded-2xl border border-slate-200 bg-white px-4 text-sm text-slate-700 outline-none transition focus:border-slate-400"
                value={filters.drawDate}
                onChange={(event) => updateFilters({ ...filters, drawDate: event.target.value })}
              />

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
          <DrawHistoryTable
            items={items}
            completePendingId={completePendingId}
            onCompleteDraw={onCompleteDraw}
            onSelectDraw={setSelectedDraw}
          />

          <div className="flex flex-wrap items-center justify-between gap-3 rounded-[1.4rem] border border-white/60 bg-white/80 px-4 py-3 shadow-[0_12px_30px_rgba(15,23,42,0.06)] backdrop-blur">
            <p className="text-sm text-slate-500">
              第 {page} / {pageCount} 页，每页 {pageSize} 条
            </p>
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="secondary"
                className="h-10 rounded-full px-4"
                disabled={loading || page <= 1}
                onClick={() => onPageChange(page - 1)}
              >
                上一页
              </Button>
              <Button
                type="button"
                variant="secondary"
                className="h-10 rounded-full px-4"
                disabled={loading || page >= pageCount}
                onClick={() => onPageChange(page + 1)}
              >
                下一页
              </Button>
            </div>
          </div>
        </>
      ) : (
        <Card className="border-white/60 bg-white/85 backdrop-blur">
          <CardContent className="py-14 text-center text-sm text-slate-500">
            当前筛选条件下没有历史开奖记录。
          </CardContent>
        </Card>
      )}

      <DrawDetailDialog
        draw={selectedDraw}
        completePendingId={completePendingId}
        onCompleteDraw={(draw) => {
          onCompleteDraw(draw);
          setSelectedDraw(null);
        }}
        onOpenChange={(open) => setSelectedDraw(open ? selectedDraw : null)}
      />
    </div>
  );
}
