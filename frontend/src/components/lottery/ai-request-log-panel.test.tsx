import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  getAIRequestLog,
  listAIRequestLogs,
  type AIRequestLogSummary,
} from "@/lib/api/methods/ai-request-logs";
import { AIRequestLogPanel } from "./ai-request-log-panel";

vi.mock("@/lib/api/methods/ai-request-logs", () => ({
  listAIRequestLogs: vi.fn(),
  getAIRequestLog: vi.fn(),
}));

const entry: AIRequestLogSummary = {
  id: 1,
  startedAt: "2026-11-19T15:00:00Z",
  purpose: "recommendation",
  lotteryCode: "ssq",
  issue: "2026136",
  model: "test-model",
  provider: "openai-compatible",
  statusCode: 429,
  attempt: 2,
  durationMs: 83,
  success: false,
};

beforeEach(() => vi.resetAllMocks());

test("列表显示调用结果并按需加载请求详情", async () => {
  vi.mocked(listAIRequestLogs).mockResolvedValue({
    items: [entry],
    page: 1,
    pageSize: 20,
    total: 1,
    keepLimit: 1000,
  });
  vi.mocked(getAIRequestLog).mockResolvedValue({
    ...entry,
    method: "POST",
    endpoint: "https://model.example.com/chat/completions",
    contentType: "application/json",
    request: '{"model":"test-model"}',
    response: '{"error":"rate limit"}',
    error: "模型请求失败: HTTP 429",
  });

  const user = userEvent.setup();
  render(<AIRequestLogPanel />);
  expect(await screen.findByText("调用失败")).toBeInTheDocument();
  expect(screen.getByText(/系统保留最近 1000 条/)).toBeInTheDocument();
  expect(getAIRequestLog).not.toHaveBeenCalled();

  await user.click(screen.getByRole("button", { name: /查看详情.*双色球/ }));
  expect(await screen.findByText("模型请求失败: HTTP 429")).toBeInTheDocument();
  expect(screen.getByRole("dialog", { name: "请求详情 #1" })).toBeInTheDocument();
  expect(screen.getByText(/POST https:\/\/model.example.com/)).toBeInTheDocument();
  expect(getAIRequestLog).toHaveBeenCalledWith(1, expect.any(AbortSignal));
});

test("可以翻页并从读取错误中重试", async () => {
  const load = vi.mocked(listAIRequestLogs);
  load
    .mockResolvedValueOnce({ items: [entry], page: 1, pageSize: 20, total: 21, keepLimit: 1000 })
    .mockRejectedValueOnce(new Error("查询失败"))
    .mockResolvedValueOnce({ items: [entry], page: 2, pageSize: 20, total: 21, keepLimit: 1000 });

  const user = userEvent.setup();
  render(<AIRequestLogPanel />);
  await user.click(await screen.findByRole("button", { name: "下一页" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("查询失败");
  await user.click(screen.getByRole("button", { name: "刷新" }));
  expect(await screen.findByText("第 2 页")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "下一页" })).toBeDisabled();
  expect(load).toHaveBeenLastCalledWith(2, "", expect.any(AbortSignal));
});
