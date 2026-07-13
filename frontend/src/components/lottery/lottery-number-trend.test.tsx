import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { LotteryNumberTrend } from "./lottery-number-trend";
import type { LotteryNumberTrendMap } from "@/hooks/use-lottery-number-trends";
import type { DrawResult, Recommendation } from "@/types/lottery";

function createDraw(
  id: string,
  issue: string,
  redNumbers: string,
  blueNumbers: string
): DrawResult {
  return {
    id,
    createdAt: "2026-07-13T00:00:00Z",
    updatedAt: "2026-07-13T00:00:00Z",
    lotteryCode: "ssq",
    issue,
    drawDate: "2026-07-13T00:00:00Z",
    redNumbers,
    blueNumbers,
    saleAmount: 0,
    prizePoolAmount: 0,
    firstPrizeAmount: 0,
    secondPrizeAmount: 0,
    source: "jisuapi",
    rawPayload: "",
    prizeDetails: [],
  };
}

function createRecommendation(): Recommendation {
  return {
    id: "recommendation-1",
    lotteryCode: "ssq",
    issue: "2026080",
    provider: "openai-compatible",
    model: "test-model",
    strategy: "",
    promptVersion: "",
    summary: "",
    basis: "",
    prizeAmount: 0,
    createdAt: "2026-07-13T00:00:00Z",
    entries: [
      {
        id: "entry-2",
        sequence: 2,
        redNumbers: "02,04,06,08,10,12",
        blueNumbers: "16",
        confidence: 0.7,
        reason: "",
        isWinning: false,
        prizeName: "",
        prizeAmount: 0,
        matchSummary: "",
      },
      {
        id: "entry-1",
        sequence: 1,
        redNumbers: "01,03,05,07,09,11",
        blueNumbers: "15",
        confidence: 0.8,
        reason: "",
        isWinning: false,
        prizeName: "",
        prizeAmount: 0,
        matchSummary: "",
      },
    ],
  };
}

function createTrends(): LotteryNumberTrendMap {
  return {
    ssq: {
      draws: [
        createDraw("draw-new", "2026079", "01,02,03,04,05,06", "16"),
        createDraw("draw-old", "2026078", "07,08,09,10,11,12", "15"),
      ],
      latestRecommendation: createRecommendation(),
      loading: false,
      error: "",
    },
    dlt: {
      draws: [
        {
          ...createDraw("dlt-draw", "26079", "01,08,16,24,35", "01,12"),
          lotteryCode: "dlt",
        },
      ],
      latestRecommendation: null,
      loading: false,
      error: "",
    },
  };
}

test("renders lottery ranges and switches between lottery types", async () => {
  const user = userEvent.setup();
  render(<LotteryNumberTrend trends={createTrends()} onRetry={() => undefined} />);

  expect(screen.getByRole("columnheader", { name: "红球区" })).toHaveAttribute("colspan", "33");
  expect(screen.getByRole("columnheader", { name: "蓝球区" })).toHaveAttribute("colspan", "16");
  expect(screen.getByRole("table")).toHaveClass("w-full");
  expect(screen.getByRole("table")).toHaveStyle({ minWidth: "1360px" });

  await user.click(screen.getByRole("button", { name: "体彩大乐透" }));

  expect(screen.getByRole("columnheader", { name: "前区" })).toHaveAttribute("colspan", "35");
  expect(screen.getByRole("columnheader", { name: "后区" })).toHaveAttribute("colspan", "12");
  expect(screen.getByRole("table")).toHaveStyle({ minWidth: "1312px" });
});

test("places sorted recommendation rows before newest draw rows", () => {
  render(<LotteryNumberTrend trends={createTrends()} onRetry={() => undefined} />);

  const rowText = screen.getAllByRole("row").map((row) => row.textContent || "");
  const firstRecommendationIndex = rowText.findIndex((text) => text.includes("推荐第 1 注"));
  const secondRecommendationIndex = rowText.findIndex((text) => text.includes("推荐第 2 注"));
  const newestDrawIndex = rowText.findIndex((text) => text.includes("第 2026079 期"));
  const oldestDrawIndex = rowText.findIndex((text) => text.includes("第 2026078 期"));

  expect(firstRecommendationIndex).toBeGreaterThan(0);
  expect(firstRecommendationIndex).toBeLessThan(secondRecommendationIndex);
  expect(secondRecommendationIndex).toBeLessThan(newestDrawIndex);
  expect(newestDrawIndex).toBeLessThan(oldestDrawIndex);
  expect(screen.getByLabelText("红球区号码 01 近 2 期开奖 1 次")).toHaveTextContent("1");
  expect(screen.getByLabelText("红球区号码 13 近 2 期开奖 0 次")).toHaveTextContent("0");
  expect(screen.getByLabelText("蓝球区号码 16 近 2 期开奖 1 次")).toHaveTextContent("1");
  expect(screen.getAllByLabelText("推荐号码 01")[0]).toHaveClass("border-rose-400");
  expect(screen.getByLabelText("开奖号码 01")).toHaveClass("bg-rose-500");
});

test("keeps one lottery usable when the other lottery fails", async () => {
  const user = userEvent.setup();
  const onRetry = vi.fn();
  const trends = createTrends();
  trends.dlt.error = "大乐透走势暂不可用";

  render(<LotteryNumberTrend trends={trends} onRetry={onRetry} />);

  expect(screen.getByText("第 2026079 期")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "体彩大乐透" }));
  expect(screen.getByText("大乐透走势暂不可用")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "重新加载" }));
  expect(onRetry).toHaveBeenCalledWith("dlt");
});
