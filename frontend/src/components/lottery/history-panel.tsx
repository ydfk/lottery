import {
  useDeferredValue,
  useEffect,
  useMemo,
  useRef,
  useState,
  useTransition,
  type MouseEvent,
} from "react";
import { format } from "date-fns";
import { ArrowDownWideNarrow, ChevronDown, SlidersHorizontal } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { TicketImportDialog } from "@/components/lottery/ticket-import-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { LotteryDisplayMode } from "@/components/lottery/display-mode-toggle";
import { DetailSheet } from "@/components/lottery/detail-sheet";
import { NumberBalls } from "@/components/lottery/number-balls";
import { TicketCard } from "@/components/lottery/ticket-card";
import {
  formatLotteryIssue,
  getLotteryDisplayName,
  lotteryDisplayOptions,
} from "@/lib/lottery-display";
import type { Ticket, TicketHistoryFilters, TicketImportResult } from "@/types/lottery";

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
  displayMode: LotteryDisplayMode;
  tickets: Ticket[];
  filters: TicketHistoryFilters;
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  total: number;
  selectedTicket: Ticket | null;
  recheckPending: boolean;
  deletePending: boolean;
  onFiltersChange: (filters: TicketHistoryFilters) => void;
  onLoadMore: () => void;
  onImportTickets: (workbook: File, imagesZip: File | null) => Promise<TicketImportResult>;
  onOpenRecord: () => void;
  onSelectTicket: (ticket: Ticket | null) => void;
  onRecheckTicket: (ticketId: string) => void;
  onDeleteTicket: (ticket: Ticket) => void;
}

const statusFilterOptions = [
  { value: "", label: "全部" },
  { value: "won", label: "已中奖" },
  { value: "not_won", label: "未中奖" },
  { value: "pending", label: "待开奖" },
];

const sortOptions = [
  { value: "latest", label: "最新录入" },
  { value: "oldest", label: "最早录入" },
  { value: "prize_high", label: "中奖金额" },
  { value: "cost_high", label: "花费金额" },
];

function HistoryTable(props: {
  items: Ticket[];
  recheckPending: boolean;
  onSelectTicket: (ticket: Ticket) => void;
  onRecheckTicket: (ticketId: string) => void;
}) {
  const { items, recheckPending, onSelectTicket, onRecheckTicket } = props;

  return (
    <Card className="border-white/60 bg-white/85 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
      <CardContent className="p-0">
        <Table className="min-w-[1120px]">
          <TableHeader>
            <TableRow className="bg-slate-50/80">
              <TableHead className="pl-4">彩种</TableHead>
              <TableHead>来源</TableHead>
              <TableHead>期号</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>开奖号码</TableHead>
              <TableHead>花费</TableHead>
              <TableHead>中奖</TableHead>
              <TableHead>录入时间</TableHead>
              <TableHead className="w-[180px] pr-4 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((ticket) => (
              <TableRow key={ticket.id}>
                <TableCell className="pl-4">
                  <Badge variant="secondary">{getLotteryDisplayName(ticket.lotteryCode)}</Badge>
                </TableCell>
                <TableCell className="text-slate-600">
                  {ticket.recommendationId ? "推荐购买" : "手动录入"}
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  第 {formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期
                  {ticket.drawDate ? (
                    <p className="mt-1 text-xs font-normal text-slate-500">
                      {format(new Date(ticket.drawDate), "yyyy-MM-dd")}
                    </p>
                  ) : null}
                </TableCell>
                <TableCell>
                  <span
                    className={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold ${
                      statusClassMap[ticket.status] ||
                      "border-slate-200 bg-slate-100 text-slate-700"
                    }`}
                  >
                    {statusLabelMap[ticket.status] || ticket.status}
                  </span>
                </TableCell>
                <TableCell className="min-w-[220px]">
                  {ticket.drawRedNumbers && ticket.drawBlueNumbers ? (
                    <NumberBalls
                      redNumbers={ticket.drawRedNumbers}
                      blueNumbers={ticket.drawBlueNumbers}
                      compact
                    />
                  ) : (
                    <span className="text-sm text-slate-400">待开奖</span>
                  )}
                </TableCell>
                <TableCell className="font-medium text-slate-900">
                  ¥ {ticket.costAmount.toFixed(2)}
                </TableCell>
                <TableCell
                  className={
                    ticket.prizeAmount > 0 ? "font-medium text-emerald-600" : "text-slate-900"
                  }
                >
                  ¥ {ticket.prizeAmount.toFixed(2)}
                </TableCell>
                <TableCell className="text-slate-500">
                  {format(new Date(ticket.purchasedAt), "yyyy-MM-dd HH:mm")}
                </TableCell>
                <TableCell className="pr-4">
                  <div className="flex justify-end gap-2">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="h-8 rounded-full"
                      onClick={() => onSelectTicket(ticket)}
                    >
                      详情
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-8 rounded-full"
                      disabled={recheckPending}
                      onClick={() => onRecheckTicket(ticket.id)}
                    >
                      {recheckPending ? "判奖中" : "重判"}
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

export function HistoryPanel(props: HistoryPanelProps) {
  const {
    displayMode,
    tickets,
    filters,
    loading,
    loadingMore,
    hasMore,
    total,
    selectedTicket,
    recheckPending,
    deletePending,
    onFiltersChange,
    onLoadMore,
    onImportTickets,
    onOpenRecord,
    onSelectTicket,
    onRecheckTicket,
    onDeleteTicket,
  } = props;
  const loadMoreTriggerRef = useRef<HTMLDivElement | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [, startTransition] = useTransition();
  const isWebDisplay = displayMode === "web";
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
      labels.push(statusLabelMap[deferredFilters.status] || deferredFilters.status);
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

  function updateFilters(nextFilters: TicketHistoryFilters) {
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
    <div className="space-y-6">
      <section className="rounded-[1.6rem] border border-white/60 bg-white/88 p-4 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Badge className="bg-slate-900 text-white hover:bg-slate-900">历史</Badge>
            <h2 className="text-base font-semibold text-slate-950">记录</h2>
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
            <TicketImportDialog onImport={onImportTickets} />
            {isWebDisplay ? (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                className="h-10 rounded-full px-4"
                onClick={onOpenRecord}
              >
                录入票据
              </Button>
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
          <CardContent className="py-14 text-center text-sm text-slate-500">
            历史记录加载中...
          </CardContent>
        </Card>
      ) : tickets.length > 0 ? (
        <>
          {isWebDisplay ? (
            <HistoryTable
              items={tickets}
              recheckPending={recheckPending}
              onSelectTicket={onSelectTicket}
              onRecheckTicket={onRecheckTicket}
            />
          ) : (
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
                          <Badge variant="secondary">
                            {getLotteryDisplayName(ticket.lotteryCode)}
                          </Badge>
                          <Badge variant="secondary">
                            {ticket.recommendationId ? "推荐购买" : "手动录入"}
                          </Badge>
                          <span
                            className={`inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold ${
                              statusClassMap[ticket.status] ||
                              "border-slate-200 bg-slate-100 text-slate-700"
                            }`}
                          >
                            {statusLabelMap[ticket.status] || ticket.status}
                          </span>
                        </div>
                        <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
                          <p className="text-lg font-semibold text-slate-900">
                            第 {formatLotteryIssue(ticket.lotteryCode, ticket.issue)} 期
                            {ticket.drawDate
                              ? ` · ${format(new Date(ticket.drawDate), "yyyy-MM-dd")}`
                              : ""}
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
          )}

          {hasMore && (
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
          )}
        </>
      ) : (
        <Card className="border-white/60 bg-white/85 backdrop-blur">
          <CardContent className="py-14 text-center text-sm text-slate-500">
            当前筛选条件下没有历史记录。
          </CardContent>
        </Card>
      )}

      <DetailSheet
        open={Boolean(selectedTicket)}
        title="票据详情"
        rightAction={
          selectedTicket ? (
            <>
              <Button
                type="button"
                variant="ghost"
                className="h-11 rounded-full px-3 text-sm text-slate-700"
                disabled={recheckPending || deletePending}
                onClick={() => onRecheckTicket(selectedTicket.id)}
              >
                {recheckPending ? "判奖中" : "重判"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="h-11 rounded-full px-3 text-sm text-rose-600 hover:bg-rose-50 hover:text-rose-700"
                disabled={deletePending || recheckPending}
                onClick={() => onDeleteTicket(selectedTicket)}
              >
                {deletePending ? "删除中" : "删除"}
              </Button>
            </>
          ) : undefined
        }
        onOpenChange={(open) => onSelectTicket(open ? selectedTicket : null)}
      >
        {selectedTicket ? <TicketCard ticket={selectedTicket} /> : null}
      </DetailSheet>
    </div>
  );
}
