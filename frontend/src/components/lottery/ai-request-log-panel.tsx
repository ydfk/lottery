import { useEffect, useState } from "react";
import { RefreshCw } from "lucide-react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  getAIRequestLog,
  listAIRequestLogs,
  type AIRequestLog,
  type AIRequestLogSummary,
  type AIRequestPurpose,
} from "@/lib/api/methods/ai-request-logs";

const dateTime = new Intl.DateTimeFormat("zh-CN", {
  dateStyle: "medium",
  timeStyle: "medium",
  hour12: false,
});

const purposeLabels: Record<AIRequestPurpose, string> = {
  recommendation: "号码推荐",
  ocr: "图片识别",
};

function requestContext(entry: AIRequestLogSummary) {
  const lottery =
    entry.lotteryCode === "ssq" ? "双色球" : entry.lotteryCode === "dlt" ? "大乐透" : "";
  return [purposeLabels[entry.purpose], lottery, entry.issue && `第 ${entry.issue} 期`]
    .filter(Boolean)
    .join(" · ");
}

function formatPayload(text: string) {
  try {
    return JSON.stringify(JSON.parse(text), null, 2);
  } catch {
    return text;
  }
}

function modelOutput(response: string) {
  try {
    const payload = JSON.parse(response) as { choices?: { message?: { content?: unknown } }[] };
    const content = payload.choices?.[0]?.message?.content;
    return typeof content === "string" ? content : "";
  } catch {
    return "";
  }
}

function PayloadSection({ title, value }: { title: string; value: string }) {
  return (
    <section className="min-w-0">
      <h3 className="mb-2 text-sm font-semibold">{title}</h3>
      <pre className="whitespace-pre-wrap break-all rounded-md border bg-muted p-3 text-xs leading-relaxed">
        {value ? formatPayload(value) : "无"}
      </pre>
    </section>
  );
}

function AIRequestLogDetails({ id }: { id: number }) {
  const [entry, setEntry] = useState<AIRequestLog | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void getAIRequestLog(id, controller.signal).then(
      (response) => {
        if (!controller.signal.aborted) setEntry(response);
      },
      (failure: unknown) => {
        if (!controller.signal.aborted) {
          setError(failure instanceof Error ? failure.message : "读取日志详情失败");
        }
      }
    );
    return () => controller.abort();
  }, [id]);

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }
  if (!entry) return <p role="status">正在读取日志详情…</p>;
  const output = modelOutput(entry.response);

  return (
    <div className="flex min-w-0 flex-col gap-4">
      <dl className="grid gap-3 text-sm sm:grid-cols-2">
        <div className="min-w-0">
          <dt className="text-muted-foreground">请求地址</dt>
          <dd className="break-all">{[entry.method, entry.endpoint].filter(Boolean).join(" ")}</dd>
        </div>
        <div className="min-w-0">
          <dt className="text-muted-foreground">响应</dt>
          <dd className="break-all">
            {entry.statusCode ? `HTTP ${entry.statusCode}` : "未收到 HTTP 响应"}
            {entry.contentType && ` · ${entry.contentType}`}
          </dd>
        </div>
      </dl>
      {entry.error && <PayloadSection title="错误详情" value={entry.error} />}
      {output && <PayloadSection title="模型输出" value={output} />}
      <PayloadSection title="请求内容" value={entry.request} />
      <PayloadSection title="响应内容" value={entry.response} />
    </div>
  );
}

function AIRequestLogRow({ entry }: { entry: AIRequestLogSummary }) {
  const title = requestContext(entry);
  const startedAt = dateTime.format(new Date(entry.startedAt));

  return (
    <li className="grid items-center gap-2 py-3 sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:gap-4">
      <div className="min-w-0">
        <p className="font-medium">{title}</p>
        <p className="break-words text-sm text-muted-foreground">
          <time dateTime={entry.startedAt}>{startedAt}</time> · {entry.model || entry.provider}
          {entry.attempt > 1 && ` · 第 ${entry.attempt} 次调用`}
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
        <Badge variant={entry.success ? "secondary" : "destructive"}>
          {entry.success ? "调用成功" : "调用失败"}
        </Badge>
        <span>{entry.statusCode ? `HTTP ${entry.statusCode}` : "无 HTTP 响应"}</span>
        <span>{entry.durationMs.toLocaleString("zh-CN")} ms</span>
      </div>
      <Dialog>
        <DialogTrigger asChild>
          <Button className="min-h-11 justify-self-start" size="sm" variant="outline">
            查看详情
            <span className="sr-only">
              {" "}
              {title} {startedAt}
            </span>
          </Button>
        </DialogTrigger>
        <DialogContent className="max-h-[85dvh] w-[calc(100%-2rem)] overflow-y-auto sm:max-w-3xl">
          <DialogHeader className="pr-6">
            <DialogTitle>请求详情 #{entry.id}</DialogTitle>
            <DialogDescription>
              {title} · {startedAt} · 第 {entry.attempt} 次调用
            </DialogDescription>
          </DialogHeader>
          <AIRequestLogDetails id={entry.id} />
        </DialogContent>
      </Dialog>
    </li>
  );
}

type LogPage = Awaited<ReturnType<typeof listAIRequestLogs>>;

export function AIRequestLogPanel() {
  const [page, setPage] = useState(1);
  const [purpose, setPurpose] = useState<AIRequestPurpose | "">("");
  const [revision, setRevision] = useState(0);
  const [logs, setLogs] = useState<LogPage | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    void listAIRequestLogs(page, purpose, controller.signal).then(
      (response) => {
        if (controller.signal.aborted) return;
        setLogs(response);
        setLoading(false);
      },
      (failure: unknown) => {
        if (controller.signal.aborted) return;
        setError(failure instanceof Error ? failure.message : "读取 AI 请求日志失败");
        setLoading(false);
      }
    );
    return () => controller.abort();
  }, [page, purpose, revision]);

  function prepareLoad() {
    setLoading(true);
    setError("");
  }

  return (
    <Card>
      <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <CardTitle>AI 请求日志</CardTitle>
          <CardDescription>
            {logs
              ? `当前筛选共 ${logs.total} 条，系统保留最近 ${logs.keepLimit} 条。`
              : "查看当前账号的号码推荐和图片识别请求。"}
          </CardDescription>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Select
            value={purpose || "all"}
            onValueChange={(value: AIRequestPurpose | "all") => {
              prepareLoad();
              setPurpose(value === "all" ? "" : value);
              setPage(1);
            }}
          >
            <SelectTrigger aria-label="调用类型" className="min-h-11 w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="all">全部类型</SelectItem>
                <SelectItem value="recommendation">号码推荐</SelectItem>
                <SelectItem value="ocr">图片识别</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          <Button
            className="min-h-11"
            variant="outline"
            disabled={loading}
            onClick={() => {
              prepareLoad();
              setRevision((current) => current + 1);
            }}
          >
            <RefreshCw data-icon="inline-start" aria-hidden="true" />
            刷新
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {loading && <p role="status">正在读取请求日志…</p>}
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {!loading &&
          !error &&
          logs &&
          (logs.items.length ? (
            <>
              <ul className="divide-y border-y">
                {logs.items.map((entry) => (
                  <AIRequestLogRow key={entry.id} entry={entry} />
                ))}
              </ul>
              <nav
                aria-label="请求日志分页"
                className="mt-4 flex flex-wrap items-center justify-end gap-2"
              >
                <Button
                  className="min-h-11"
                  size="sm"
                  variant="outline"
                  disabled={logs.page <= 1}
                  onClick={() => {
                    prepareLoad();
                    setPage(logs.page - 1);
                  }}
                >
                  上一页
                </Button>
                <span className="text-sm text-muted-foreground">第 {logs.page} 页</span>
                <Button
                  className="min-h-11"
                  size="sm"
                  variant="outline"
                  disabled={logs.page * logs.pageSize >= logs.total}
                  onClick={() => {
                    prepareLoad();
                    setPage(logs.page + 1);
                  }}
                >
                  下一页
                </Button>
              </nav>
            </>
          ) : (
            <p className="py-4 text-sm text-muted-foreground">没有符合条件的请求日志。</p>
          ))}
      </CardContent>
    </Card>
  );
}
