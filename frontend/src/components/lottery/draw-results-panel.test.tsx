import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DrawResultsPanel } from "./draw-results-panel";
import type { LotteryNumberTrendMap } from "@/hooks/use-lottery-number-trends";
import type { DrawResult } from "@/types/lottery";

const emptyTrends: LotteryNumberTrendMap = {
  ssq: {
    draws: [],
    latestRecommendation: null,
    loading: false,
    error: "",
  },
  dlt: {
    draws: [],
    latestRecommendation: null,
    loading: false,
    error: "",
  },
};

const draw: DrawResult = {
  id: "draw-1",
  createdAt: "2026-08-03T10:00:00Z",
  updatedAt: "2026-08-03T10:00:00Z",
  lotteryCode: "ssq",
  issue: "2026080",
  drawDate: "2026-08-03",
  redNumbers: "01,02,03,04,05,06",
  blueNumbers: "07",
  saleAmount: 0,
  prizePoolAmount: 0,
  firstPrizeAmount: 0,
  secondPrizeAmount: 0,
  source: "jisuapi",
  rawPayload: "",
  prizeDetails: [],
};

test("defaults to trend view and keeps the existing draw list available", async () => {
  const user = userEvent.setup();

  render(
    <DrawResultsPanel
      displayMode="web"
      items={[]}
      numberTrends={emptyTrends}
      filters={{ lotteryCode: "", issue: "", drawDate: "", sort: "latest" }}
      loading={false}
      loadingMore={false}
      hasMore={false}
      total={0}
      completePendingId=""
      syncPending={false}
      onFiltersChange={() => undefined}
      onLoadMore={() => undefined}
      onRetryTrend={() => undefined}
      onSyncIssue={() => undefined}
      onCompleteDraw={() => undefined}
    />
  );

  expect(screen.getByRole("button", { name: "号码走势" })).toHaveAttribute("aria-pressed", "true");
  expect(screen.getByText("当前彩种暂无可展示的开奖记录或推荐号码。")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "开奖列表" }));

  expect(screen.getByText("当前筛选条件下没有历史开奖记录。")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "筛选" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "同步开奖" })).toBeInTheDocument();
});

test("does not expose trend view in app display mode", () => {
  render(
    <DrawResultsPanel
      displayMode="app"
      items={[]}
      numberTrends={emptyTrends}
      filters={{ lotteryCode: "", issue: "", drawDate: "", sort: "latest" }}
      loading={false}
      loadingMore={false}
      hasMore={false}
      total={0}
      completePendingId=""
      syncPending={false}
      onFiltersChange={() => undefined}
      onLoadMore={() => undefined}
      onRetryTrend={() => undefined}
      onSyncIssue={() => undefined}
      onCompleteDraw={() => undefined}
    />
  );

  expect(screen.queryByRole("button", { name: "号码走势" })).not.toBeInTheDocument();
  expect(screen.getByText("当前筛选条件下没有历史开奖记录。")).toBeInTheDocument();
});

test("keeps draw completion in the detail dialog only", async () => {
  const user = userEvent.setup();

  render(
    <DrawResultsPanel
      displayMode="app"
      items={[draw]}
      numberTrends={emptyTrends}
      filters={{ lotteryCode: "", issue: "", drawDate: "", sort: "latest" }}
      loading={false}
      loadingMore={false}
      hasMore={false}
      total={1}
      completePendingId=""
      syncPending={false}
      onFiltersChange={() => undefined}
      onLoadMore={() => undefined}
      onRetryTrend={() => undefined}
      onSyncIssue={() => undefined}
      onCompleteDraw={() => undefined}
    />
  );

  expect(screen.queryByRole("button", { name: "补全" })).not.toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "详情" }));
  expect(screen.getByRole("button", { name: "补全" })).toBeInTheDocument();
});
