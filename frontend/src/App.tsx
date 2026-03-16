import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  ClipboardCheck,
  ImageUp,
  LogOut,
  RefreshCw,
  ScanSearch,
  Sparkles,
  Ticket,
  Trophy,
} from "lucide-react";
import { toast } from "sonner";
import { AuthPanel, type AuthMode } from "@/components/lottery/auth-panel";
import { TicketConfirmPanel } from "@/components/lottery/ticket-confirm-panel";
import { TicketRecognitionPanel } from "@/components/lottery/ticket-recognition-panel";
import { TicketUploadPanel } from "@/components/lottery/ticket-upload-panel";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { NumberBalls } from "@/components/lottery/number-balls";
import { SummaryCard } from "@/components/lottery/summary-card";
import { TicketCard } from "@/components/lottery/ticket-card";
import { getStoredToken } from "@/lib/api/client";
import { getProfile, loginAndStoreToken, logout, register } from "@/lib/api/methods/auth";
import {
  createTicket,
  formatParsedEntry,
  generateRecommendation,
  getDashboard,
  getTickets,
  recognizeTicket,
  syncDraws,
  uploadTicketImage,
} from "@/lib/api/methods/lottery";
import type { AuthUser } from "@/types/auth";
import type {
  DashboardData,
  ParsedEntry,
  Recommendation,
  Ticket as TicketRecord,
  TicketRecognitionDraft,
  TicketUpload,
} from "@/types/lottery";

type TabKey = "dashboard" | "upload" | "recognize" | "confirm" | "tickets" | "recommendation";

const tabs: { key: TabKey; label: string; icon: typeof Sparkles }[] = [
  { key: "dashboard", label: "看板", icon: Trophy },
  { key: "upload", label: "上传原图", icon: ImageUp },
  { key: "recognize", label: "识别校对", icon: ScanSearch },
  { key: "confirm", label: "确认入库", icon: ClipboardCheck },
  { key: "tickets", label: "票据记录", icon: Ticket },
  { key: "recommendation", label: "推荐中心", icon: Sparkles },
];

function formatCurrency(value: number) {
  return `¥ ${value.toFixed(2)}`;
}

function buildEntryText(entries: ParsedEntry[]) {
  return entries
    .map((entry) => {
      const { redNumbers, blueNumbers } = formatParsedEntry(entry);
      return `${redNumbers}+${blueNumbers}`;
    })
    .join("\n");
}

function parseEntryText(value: string) {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [redPart, bluePart] = line.split("+");
      const redNumbers = redPart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      const blueNumbers = bluePart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      return { red: redNumbers, blue: blueNumbers };
    })
    .filter((entry) => entry.red.length > 0 && entry.blue.length > 0);
}

export default function App() {
  const [activeTab, setActiveTab] = useState<TabKey>("dashboard");
  const [authMode, setAuthMode] = useState<AuthMode>("login");
  const [currentUser, setCurrentUser] = useState<AuthUser | null>(null);
  const [authUsername, setAuthUsername] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [authPending, setAuthPending] = useState(false);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [tickets, setTickets] = useState<TicketRecord[]>([]);
  const [latestRecommendation, setLatestRecommendation] = useState<Recommendation | null>(null);
  const [loading, setLoading] = useState(true);
  const [uploadPending, setUploadPending] = useState(false);
  const [recognizePending, setRecognizePending] = useState(false);
  const [submitPending, setSubmitPending] = useState(false);
  const [actionPending, setActionPending] = useState(false);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [uploadedTicket, setUploadedTicket] = useState<TicketUpload | null>(null);
  const [recognitionDraft, setRecognitionDraft] = useState<TicketRecognitionDraft | null>(null);
  const [issue, setIssue] = useState("");
  const [ocrText, setOcrText] = useState("");
  const [notes, setNotes] = useState("");
  const [entryText, setEntryText] = useState("");
  const [lastScanned, setLastScanned] = useState<TicketRecord | null>(null);

  function resetLotteryState() {
    setDashboard(null);
    setTickets([]);
    setLatestRecommendation(null);
    setLastScanned(null);
    setActiveTab("dashboard");
  }

  function resetTicketWorkflow() {
    setSelectedImage(null);
    setPreviewUrl("");
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setIssue("");
    setOcrText("");
    setNotes("");
    setEntryText("");
  }

  function handleRequestError(error: unknown, fallbackMessage: string) {
    if (!getStoredToken()) {
      setCurrentUser(null);
      setAuthMode("login");
      resetLotteryState();
    }
    toast.error(error instanceof Error ? error.message : fallbackMessage);
  }

  async function loadData(showLoading = true) {
    if (showLoading) {
      setLoading(true);
    }
    try {
      const [dashboardData, ticketData] = await Promise.all([getDashboard(), getTickets()]);
      setDashboard(dashboardData);
      setLatestRecommendation(dashboardData.latestRecommendation || null);
      setTickets(ticketData);
    } catch (error) {
      handleRequestError(error, "加载数据失败");
    } finally {
      if (showLoading) {
        setLoading(false);
      }
    }
  }

  useEffect(() => {
    async function bootstrap() {
      if (!getStoredToken()) {
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const profile = await getProfile();
        setCurrentUser(profile);
        await loadData(false);
      } catch (error) {
        logout();
        resetLotteryState();
        toast.error(error instanceof Error ? error.message : "登录已失效，请重新登录");
      } finally {
        setLoading(false);
      }
    }

    void bootstrap();
  }, []);

  useEffect(() => {
    if (!selectedImage) {
      setPreviewUrl("");
      return;
    }

    const objectUrl = URL.createObjectURL(selectedImage);
    setPreviewUrl(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [selectedImage]);

  const recommendationEntries = useMemo(
    () => latestRecommendation?.entries ?? dashboard?.latestRecommendation?.entries ?? [],
    [dashboard?.latestRecommendation?.entries, latestRecommendation?.entries]
  );

  const workflowStatus = useMemo(() => {
    if (recognitionDraft) {
      return { label: "待确认入库", hint: "识别结果已生成，可进入下一步确认号码" };
    }
    if (uploadedTicket) {
      return { label: "待识别", hint: "原图已保存，下一步执行 OCR 识别" };
    }
    return { label: "待上传", hint: "从上传原图开始新的录票流程" };
  }, [recognitionDraft, uploadedTicket]);

  async function handleAuthSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!authUsername.trim() || !authPassword.trim()) {
      toast.error("请输入用户名和密码");
      return;
    }

    setAuthPending(true);
    try {
      if (authMode === "register") {
        await register(authUsername.trim(), authPassword);
      }
      await loginAndStoreToken(authUsername.trim(), authPassword);
      const profile = await getProfile();
      setCurrentUser(profile);
      toast.success(authMode === "login" ? "登录成功" : "注册并登录成功");
      await loadData();
    } catch (error) {
      handleRequestError(error, "登录失败");
    } finally {
      setAuthPending(false);
    }
  }

  function handleLogout() {
    logout();
    setCurrentUser(null);
    resetLotteryState();
    resetTicketWorkflow();
    toast.success("已退出登录");
  }

  async function handleGenerateRecommendation() {
    setActionPending(true);
    try {
      const result = await generateRecommendation(dashboard?.lottery.recommendationCount || 5);
      setLatestRecommendation(result);
      toast.success("推荐已生成");
      await loadData();
      setActiveTab("recommendation");
    } catch (error) {
      handleRequestError(error, "生成推荐失败");
    } finally {
      setActionPending(false);
    }
  }

  async function handleSyncDraws() {
    setActionPending(true);
    try {
      const result = await syncDraws();
      toast.success(`开奖同步完成，更新 ${result.syncedCount} 期`);
      await loadData();
    } catch (error) {
      handleRequestError(error, "开奖同步失败");
    } finally {
      setActionPending(false);
    }
  }

  function handleSelectImage(file: File | null) {
    setSelectedImage(file);
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setIssue("");
    setOcrText("");
    setNotes("");
    setEntryText("");
  }

  async function handleUploadTicketImage() {
    if (!selectedImage) {
      toast.error("请先上传彩票照片");
      return;
    }

    const formData = new FormData();
    formData.append("image", selectedImage);

    setUploadPending(true);
    try {
      const upload = await uploadTicketImage(formData);
      setUploadedTicket(upload);
      setRecognitionDraft(null);
      toast.success("原图上传成功");
      setActiveTab("recognize");
    } catch (error) {
      handleRequestError(error, "上传原图失败");
    } finally {
      setUploadPending(false);
    }
  }

  async function handleRecognizeTicket() {
    if (!uploadedTicket) {
      toast.error("请先上传原图");
      return;
    }

    setRecognizePending(true);
    try {
      const draft = await recognizeTicket({
        uploadId: uploadedTicket.id,
        ocrText: ocrText || undefined,
      });
      setUploadedTicket(draft.upload);
      setRecognitionDraft(draft);
      setIssue(draft.issue || "");
      setEntryText(buildEntryText(draft.entries));
      toast.success("识别完成，请校对号码");
      setActiveTab("confirm");
    } catch (error) {
      handleRequestError(error, "识别失败");
    } finally {
      setRecognizePending(false);
    }
  }

  async function handleCreateTicket() {
    if (!uploadedTicket) {
      toast.error("请先上传原图");
      return;
    }

    const entries = parseEntryText(entryText);
    if (entries.length === 0) {
      toast.error("请至少确认一注号码");
      return;
    }

    setSubmitPending(true);
    try {
      const ticket = await createTicket({
        uploadId: uploadedTicket.id,
        issue: issue || undefined,
        purchasedAt: new Date().toISOString(),
        notes: notes || undefined,
        entries: entries.map((entry) => formatParsedEntry(entry)),
      });
      setLastScanned(ticket);
      toast.success("票据已入库并完成判奖检查");
      resetTicketWorkflow();
      await loadData();
      setActiveTab("tickets");
    } catch (error) {
      handleRequestError(error, "入库失败");
    } finally {
      setSubmitPending(false);
    }
  }

  if (!currentUser) {
    if (loading) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.92),rgba(226,232,240,0.9)_45%,rgba(203,213,225,1))]">
          <div className="rounded-3xl bg-white/80 px-6 py-4 text-sm text-slate-600 shadow-sm">
            正在检查登录状态...
          </div>
        </div>
      );
    }

    return (
      <AuthPanel
        mode={authMode}
        username={authUsername}
        password={authPassword}
        pending={authPending}
        onModeChange={setAuthMode}
        onUsernameChange={setAuthUsername}
        onPasswordChange={setAuthPassword}
        onSubmit={handleAuthSubmit}
      />
    );
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.85),rgba(243,244,246,0.9)_45%,rgba(226,232,240,1))] pb-28">
      <div className="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-4 pb-12 pt-4 sm:px-6">
        <section className="overflow-hidden rounded-[2rem] border border-white/60 bg-[linear-gradient(135deg,rgba(15,23,42,0.95),rgba(37,99,235,0.9)_58%,rgba(224,36,94,0.82))] p-6 text-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
          <div className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
            <div className="max-w-2xl">
              <Badge className="bg-white/15 text-white hover:bg-white/20">双色球 MVP</Badge>
              <h1 className="mt-4 text-3xl font-semibold tracking-tight sm:text-4xl">
                彩票管理、AI 推荐与拍照录票一体化
              </h1>
              <p className="mt-3 text-sm leading-6 text-white/80 sm:text-base">
                先围绕福彩双色球建立完整闭环：推荐、购买记录、拍照识别、开奖同步和自动判奖，后续再平滑扩展更多彩种。
              </p>
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <Button
                variant="secondary"
                className="h-11 rounded-2xl bg-white text-slate-900 hover:bg-white/90"
                onClick={() => void handleGenerateRecommendation()}
                disabled={actionPending}
              >
                <Sparkles className="mr-2 size-4" />
                生成推荐
              </Button>
              <Button
                className="h-11 rounded-2xl border border-white/20 bg-white/10 text-white hover:bg-white/15"
                onClick={() => void handleSyncDraws()}
                disabled={actionPending}
              >
                <RefreshCw className="mr-2 size-4" />
                同步开奖
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="h-11 rounded-2xl border border-white/20 bg-white/10 text-white hover:bg-white/15"
                onClick={handleLogout}
              >
                <LogOut className="mr-2 size-4" />
                {currentUser.username}
              </Button>
            </div>
          </div>

          <div className="mt-8 grid gap-4 md:grid-cols-[1.2fr_0.8fr]">
            <Card className="border-white/10 bg-white/10 text-white backdrop-blur">
              <CardHeader>
                <CardTitle className="text-sm font-medium text-white/75">最近一期开奖</CardTitle>
              </CardHeader>
              <CardContent>
                {dashboard?.latestDraw ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-2xl font-semibold">第 {dashboard.latestDraw.issue} 期</p>
                        <p className="text-sm text-white/70">
                          奖池 {formatCurrency(dashboard.latestDraw.prizePoolAmount)}
                        </p>
                      </div>
                      <Badge className="bg-white/15 text-white">已同步</Badge>
                    </div>
                    <NumberBalls
                      redNumbers={dashboard.latestDraw.redNumbers}
                      blueNumbers={dashboard.latestDraw.blueNumbers}
                    />
                  </div>
                ) : (
                  <p className="text-sm text-white/75">
                    还没有同步到开奖数据，先点击右上角“同步开奖”。
                  </p>
                )}
              </CardContent>
            </Card>

            <div className="grid gap-4 sm:grid-cols-2">
              <SummaryCard
                title="已录入票据"
                value={String(dashboard?.stats.totalTickets ?? 0)}
                hint="包含拍照识别入库的购彩记录"
              />
              <SummaryCard
                title="中奖票据"
                value={String(dashboard?.stats.wonTickets ?? 0)}
                hint="系统会在开奖同步后自动补判"
              />
              <SummaryCard
                title="累计投入"
                value={formatCurrency(dashboard?.stats.totalCost ?? 0)}
                hint="当前只统计双色球"
              />
              <SummaryCard
                title="累计奖金"
                value={formatCurrency(dashboard?.stats.totalPrize ?? 0)}
                hint="优先使用同步到库的奖级奖金"
              />
            </div>
          </div>
        </section>

        <div className="mt-6 grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
          <main className="space-y-6">
            {activeTab === "dashboard" && (
              <>
                <Card className="border-white/60 bg-white/80 backdrop-blur">
                  <CardHeader className="flex flex-row items-center justify-between">
                    <div>
                      <CardTitle className="text-slate-900">最新推荐</CardTitle>
                      <p className="mt-1 text-sm text-slate-500">
                        {latestRecommendation?.summary || "当前还没有推荐记录"}
                      </p>
                    </div>
                    <Badge variant="secondary">
                      {latestRecommendation?.provider ||
                        dashboard?.lottery.recommendationProvider ||
                        "mock"}
                    </Badge>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {recommendationEntries.length > 0 ? (
                      recommendationEntries.slice(0, 3).map((entry) => (
                        <div key={entry.id} className="rounded-2xl bg-slate-50 p-4">
                          <div className="flex items-center justify-between gap-3">
                            <span className="text-sm font-medium text-slate-600">
                              推荐 {entry.sequence}
                            </span>
                            <span className="text-xs text-slate-500">
                              置信度 {(entry.confidence * 100).toFixed(0)}%
                            </span>
                          </div>
                          <div className="mt-3">
                            <NumberBalls
                              redNumbers={entry.redNumbers}
                              blueNumbers={entry.blueNumbers}
                              compact
                            />
                          </div>
                          <p className="mt-3 text-sm leading-6 text-slate-600">{entry.reason}</p>
                        </div>
                      ))
                    ) : (
                      <p className="text-sm text-slate-500">
                        还没有推荐结果，点击“生成推荐”即可创建一批号码。
                      </p>
                    )}
                  </CardContent>
                </Card>

                <Card className="border-white/60 bg-white/80 backdrop-blur">
                  <CardHeader>
                    <CardTitle className="text-slate-900">最近票据</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {tickets.slice(0, 3).map((ticket) => (
                      <TicketCard key={ticket.id} ticket={ticket} />
                    ))}
                    {!loading && tickets.length === 0 && (
                      <p className="text-sm text-slate-500">
                    还没有票据记录，可以先去“上传原图”开始流程。
                      </p>
                    )}
                  </CardContent>
                </Card>
              </>
            )}

            {activeTab === "upload" && (
              <TicketUploadPanel
                previewUrl={previewUrl}
                selectedImage={selectedImage}
                uploadPending={uploadPending}
                uploadedTicket={uploadedTicket}
                onSelectImage={handleSelectImage}
                onUpload={() => void handleUploadTicketImage()}
              />
            )}

            {activeTab === "recognize" && (
              <TicketRecognitionPanel
                uploadedTicket={uploadedTicket}
                recognitionDraft={recognitionDraft}
                ocrText={ocrText}
                recognizePending={recognizePending}
                onOCRTextChange={setOcrText}
                onRecognize={() => void handleRecognizeTicket()}
              />
            )}

            {activeTab === "confirm" && (
              <TicketConfirmPanel
                recognitionDraft={recognitionDraft}
                issue={issue}
                notes={notes}
                entryText={entryText}
                submitPending={submitPending}
                onIssueChange={setIssue}
                onNotesChange={setNotes}
                onEntryTextChange={setEntryText}
                onCreateTicket={() => void handleCreateTicket()}
              />
            )}

            {activeTab === "recommendation" && (
              <Card className="border-white/60 bg-white/80 backdrop-blur">
                <CardHeader className="flex flex-row items-center justify-between">
                  <div>
                    <CardTitle className="text-slate-900">推荐中心</CardTitle>
                    <p className="mt-1 text-sm text-slate-500">
                      {latestRecommendation?.basis || "可按当前彩种默认数量生成推荐"}
                    </p>
                  </div>
                  <Button
                    onClick={() => void handleGenerateRecommendation()}
                    disabled={actionPending}
                  >
                    <Sparkles className="mr-2 size-4" />
                    重新生成
                  </Button>
                </CardHeader>
                <CardContent className="space-y-4">
                  {recommendationEntries.length > 0 ? (
                    recommendationEntries.map((entry) => (
                      <div
                        key={entry.id}
                        className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4"
                      >
                        <div className="flex items-center justify-between gap-4">
                          <div>
                            <p className="text-sm font-medium text-slate-700">
                              推荐 {entry.sequence}
                            </p>
                            <p className="text-xs text-slate-500">
                              目标期号 {latestRecommendation?.issue || "待定"}
                            </p>
                          </div>
                          <Badge variant="secondary">{(entry.confidence * 100).toFixed(0)}%</Badge>
                        </div>
                        <div className="mt-4">
                          <NumberBalls
                            redNumbers={entry.redNumbers}
                            blueNumbers={entry.blueNumbers}
                          />
                        </div>
                        <p className="mt-4 text-sm leading-6 text-slate-600">{entry.reason}</p>
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-slate-500">还没有推荐记录。</p>
                  )}
                </CardContent>
              </Card>
            )}

            {activeTab === "tickets" && (
              <div className="space-y-4">
                {tickets.map((ticket) => (
                  <TicketCard key={ticket.id} ticket={ticket} />
                ))}
                {!loading && tickets.length === 0 && (
                  <Card className="border-white/60 bg-white/80 backdrop-blur">
                    <CardContent className="py-10 text-center text-sm text-slate-500">
                      当前还没有票据记录。
                    </CardContent>
                  </Card>
                )}
              </div>
            )}
          </main>

          <aside className="space-y-6">
            <Card className="border-white/60 bg-white/80 backdrop-blur">
              <CardHeader>
                <CardTitle className="text-slate-900">当前实现状态</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm leading-6 text-slate-600">
                <p>
                  已支持双色球开奖号同步、鉴权访问、推荐扩展点、拍照录票接口、自动判奖和移动端录入界面。
                </p>
                <p>录票流程已经拆成上传原图、识别校对、确认入库三个独立页面，更适合逐步调试。</p>
                <p>推荐默认走 mock 提供者，配置 OpenAI 兼容接口后可切到真实模型。</p>
              </CardContent>
            </Card>

            <Card className="border-white/60 bg-white/80 backdrop-blur">
              <CardHeader>
                <CardTitle className="text-slate-900">当前录票进度</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-sm font-medium text-slate-700">{workflowStatus.label}</p>
                  <p className="mt-2 text-sm leading-6 text-slate-500">{workflowStatus.hint}</p>
                </div>
                {uploadedTicket && (
                  <div className="rounded-2xl border border-slate-200 bg-white p-4">
                    <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Upload ID</p>
                    <p className="mt-2 break-all text-sm text-slate-700">{uploadedTicket.id}</p>
                  </div>
                )}
              </CardContent>
            </Card>

            {lastScanned && (
              <Card className="border-white/60 bg-white/80 backdrop-blur">
                <CardHeader>
                  <CardTitle className="text-slate-900">最近识别结果</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-slate-500">期号</span>
                    <span className="font-semibold text-slate-900">{lastScanned.issue}</span>
                  </div>
                  {lastScanned.entries.map((entry) => (
                    <div key={entry.id} className="rounded-2xl bg-slate-50 p-3">
                      <NumberBalls
                        redNumbers={entry.redNumbers}
                        blueNumbers={entry.blueNumbers}
                        compact
                      />
                      <div className="mt-3 flex items-center justify-between text-sm">
                        <span className="text-slate-500">{entry.matchSummary}</span>
                        <span className="font-semibold text-slate-900">
                          {formatCurrency(entry.prizeAmount)}
                        </span>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}
          </aside>
        </div>
      </div>

      <nav className="fixed inset-x-0 bottom-4 z-50 mx-auto flex w-[calc(100%-1.5rem)] max-w-5xl items-center gap-2 overflow-x-auto rounded-[2rem] border border-white/70 bg-white/90 p-2 shadow-[0_18px_40px_rgba(15,23,42,0.14)] backdrop-blur">
        {tabs.map((item) => {
          const Icon = item.icon;
          const active = activeTab === item.key;
          return (
            <button
              key={item.key}
              type="button"
              className={`flex min-w-[112px] flex-1 items-center justify-center gap-2 rounded-[1.25rem] px-3 py-3 text-sm transition ${
                active
                  ? "bg-slate-900 text-white shadow-sm"
                  : "text-slate-500 hover:bg-slate-100 hover:text-slate-900"
              }`}
              onClick={() => setActiveTab(item.key)}
            >
              <Icon className="size-4" />
              <span>{item.label}</span>
            </button>
          );
        })}
      </nav>
    </div>
  );
}
