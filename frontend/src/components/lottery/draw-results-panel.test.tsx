import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DrawResultsPanel } from "./draw-results-panel";
import type { LotteryNumberTrendMap } from "@/hooks/use-lottery-number-trends";

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

test("defaults to trend view and keeps the existing draw list available", async () => {
  const user = userEvent.setup();

  render(
    <DrawResultsPanel
      displayMode="web"
      items={[]}
      numberTrends={emptyTrends}
      filters={{ lotteryCode: "", issue: "", drawDate: "", sort: "latest" }}
      loading={false}
      page={1}
      pageSize={20}
      total={0}
      completePendingId=""
      syncPending={false}
      onFiltersChange={() => undefined}
      onPageChange={() => undefined}
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
      page={1}
      pageSize={20}
      total={0}
      completePendingId=""
      syncPending={false}
      onFiltersChange={() => undefined}
      onPageChange={() => undefined}
      onRetryTrend={() => undefined}
      onSyncIssue={() => undefined}
      onCompleteDraw={() => undefined}
    />
  );

  expect(screen.queryByRole("button", { name: "号码走势" })).not.toBeInTheDocument();
  expect(screen.getByText("当前筛选条件下没有历史开奖记录。")).toBeInTheDocument();
});
