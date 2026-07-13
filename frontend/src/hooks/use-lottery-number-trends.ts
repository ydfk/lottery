import { useCallback, useEffect, useState } from "react";
import { getDrawResults, getRecommendations } from "@/lib/api/methods/lottery";
import type { DrawResult, Recommendation } from "@/types/lottery";

export type TrendLotteryCode = "ssq" | "dlt";

export interface LotteryNumberTrendData {
  draws: DrawResult[];
  latestRecommendation: Recommendation | null;
  loading: boolean;
  error: string;
}

export type LotteryNumberTrendMap = Record<TrendLotteryCode, LotteryNumberTrendData>;

const LOTTERY_CODES: TrendLotteryCode[] = ["ssq", "dlt"];
const TREND_DRAW_COUNT = 50;

export function isTrendLotteryCode(value: string): value is TrendLotteryCode {
  return LOTTERY_CODES.includes(value as TrendLotteryCode);
}

function createEmptyTrendData(): LotteryNumberTrendData {
  return {
    draws: [],
    latestRecommendation: null,
    loading: false,
    error: "",
  };
}

export function useLotteryNumberTrends(active: boolean) {
  const [trends, setTrends] = useState<LotteryNumberTrendMap>(() => ({
    ssq: createEmptyTrendData(),
    dlt: createEmptyTrendData(),
  }));

  const loadTrend = useCallback(async (lotteryCode: TrendLotteryCode) => {
    setTrends((current) => ({
      ...current,
      [lotteryCode]: {
        ...current[lotteryCode],
        loading: true,
        error: "",
      },
    }));

    try {
      const [drawPage, recommendationPage] = await Promise.all([
        getDrawResults(1, TREND_DRAW_COUNT, {
          lotteryCode,
          issue: "",
          drawDate: "",
          sort: "latest",
        }),
        getRecommendations(1, 1, {
          lotteryCode,
          status: "",
          sort: "latest",
        }),
      ]);

      setTrends((current) => ({
        ...current,
        [lotteryCode]: {
          draws: drawPage.items.slice(0, TREND_DRAW_COUNT),
          latestRecommendation: recommendationPage.items[0] ?? null,
          loading: false,
          error: "",
        },
      }));
    } catch (error) {
      setTrends((current) => ({
        ...current,
        [lotteryCode]: {
          ...current[lotteryCode],
          loading: false,
          error: error instanceof Error ? error.message : "号码走势加载失败",
        },
      }));
    }
  }, []);

  const loadAllTrends = useCallback(async () => {
    await Promise.allSettled(LOTTERY_CODES.map((lotteryCode) => loadTrend(lotteryCode)));
  }, [loadTrend]);

  useEffect(() => {
    if (!active) {
      return;
    }

    void loadAllTrends();
  }, [active, loadAllTrends]);

  return {
    trends,
    reloadTrend: loadTrend,
  };
}
