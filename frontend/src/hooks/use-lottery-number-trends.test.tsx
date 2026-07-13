import { renderHook, waitFor } from "@testing-library/react";
import { getDrawResults, getRecommendations } from "@/lib/api/methods/lottery";
import { useLotteryNumberTrends } from "./use-lottery-number-trends";

vi.mock("@/lib/api/methods/lottery", () => ({
  getDrawResults: vi.fn(),
  getRecommendations: vi.fn(),
}));

const mockedGetDrawResults = vi.mocked(getDrawResults);
const mockedGetRecommendations = vi.mocked(getRecommendations);

beforeEach(() => {
  vi.clearAllMocks();
});

test("loads both lotteries independently with a 50 draw window", async () => {
  mockedGetDrawResults.mockImplementation(async (_page, _pageSize, filters) => ({
    items: [],
    page: 1,
    pageSize: 50,
    total: filters.lotteryCode === "ssq" ? 80 : 60,
    hasMore: true,
  }));
  mockedGetRecommendations.mockImplementation(async (_page, _pageSize, filters) => {
    if (filters.lotteryCode === "ssq") {
      throw new Error("双色球推荐加载失败");
    }
    return {
      items: [],
      page: 1,
      pageSize: 1,
      total: 0,
      hasMore: false,
    };
  });

  const { result } = renderHook(() => useLotteryNumberTrends(true));

  await waitFor(() => {
    expect(result.current.trends.ssq.loading).toBe(false);
    expect(result.current.trends.dlt.loading).toBe(false);
    expect(result.current.trends.ssq.error).toBe("双色球推荐加载失败");
  });

  expect(result.current.trends.dlt.error).toBe("");
  expect(mockedGetDrawResults).toHaveBeenCalledWith(
    1,
    50,
    expect.objectContaining({ lotteryCode: "ssq", sort: "latest" })
  );
  expect(mockedGetDrawResults).toHaveBeenCalledWith(
    1,
    50,
    expect.objectContaining({ lotteryCode: "dlt", sort: "latest" })
  );
  expect(mockedGetRecommendations).toHaveBeenCalledWith(
    1,
    1,
    expect.objectContaining({ lotteryCode: "dlt", sort: "latest" })
  );
});
