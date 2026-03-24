import { useEffect, useState, type FormEvent } from "react";
import { History, LayoutDashboard, ReceiptText, Sparkles } from "lucide-react";
import { toast } from "sonner";
import { AuthPanel } from "@/components/lottery/auth-panel";
import { DashboardPanel } from "@/components/lottery/dashboard-panel";
import { HistoryPanel } from "@/components/lottery/history-panel";
import { RecommendationPanel } from "@/components/lottery/recommendation-panel";
import { RecordPanel } from "@/components/lottery/record-panel";
import { getStoredToken } from "@/lib/api/client";
import { getProfile, loginAndStoreToken, logout } from "@/lib/api/methods/auth";
import {
  createTicket,
  deleteRecommendation,
  deleteTicket,
  formatParsedEntry,
  getDashboard,
  getRecommendationDetail,
  getRecommendations,
  generateRecommendation,
  getTicketHistory,
  recheckTicket,
  recognizeTicket,
  uploadTicketImage,
} from "@/lib/api/methods/lottery";
import { formatLotteryDrawDate } from "@/lib/lottery-display";
import type { AuthUser } from "@/types/auth";
import type {
  DashboardData,
  ParsedEntry,
  Recommendation,
  RecommendationFilters,
  Ticket as TicketRecord,
  TicketHistoryFilters,
  TicketRecognitionDraft,
  TicketUpload,
} from "@/types/lottery";

type TabKey = "dashboard" | "recommendation" | "records" | "history";

const tabs: { key: TabKey; label: string; icon: typeof LayoutDashboard }[] = [
  { key: "dashboard", label: "看板", icon: LayoutDashboard },
  { key: "recommendation", label: "推荐", icon: Sparkles },
  { key: "records", label: "记录", icon: ReceiptText },
  { key: "history", label: "历史", icon: History },
];

const DEFAULT_HISTORY_FILTERS: TicketHistoryFilters = {
  lotteryCode: "",
  status: "",
  sort: "latest",
};

const DEFAULT_RECOMMENDATION_FILTERS: RecommendationFilters = {
  lotteryCode: "",
  status: "",
  sort: "latest",
};

function buildParsedEntryText(entries: ParsedEntry[], lotteryCode?: string) {
  return entries
    .map((entry) => {
      const { redNumbers, blueNumbers } = formatParsedEntry(entry);
      const additional = lotteryCode === "dlt" && entry.isAdditional ? " 追加" : "";
      const multiple = entry.multiple > 1 ? ` (${entry.multiple})` : "";
      return `${redNumbers}+${blueNumbers}${additional}${multiple}`;
    })
    .join("\n");
}

function buildRecommendationEntryText(recommendation: Recommendation) {
  return recommendation.entries
    .map((entry) => `${entry.redNumbers}+${entry.blueNumbers}`)
    .join("\n");
}

function calculateEntriesCost(entries: ParsedEntry[]) {
  return entries.reduce((total, entry) => {
    const perBetCost = entry.isAdditional ? 3 : 2;
    return total + Math.max(1, entry.multiple) * perBetCost;
  }, 0);
}

function parseEntryText(value: string, lotteryCode: string) {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const isAdditional = line.includes("追加");
      const sourceLine = line.replace(/追加/g, "").trim();
      const multipleMatch = sourceLine.match(/[（(]\s*(\d+)\s*[)）]\s*$/);
      const multiple = multipleMatch ? Number(multipleMatch[1]) : 1;
      const normalizedLine = sourceLine.replace(/[（(]\s*\d+\s*[)）]\s*$/, "").trim();
      const [redPart, bluePart] = normalizedLine.split("+");
      const redNumbers = redPart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      const blueNumbers = bluePart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      return {
        red: redNumbers,
        blue: blueNumbers,
        multiple: multiple > 0 ? multiple : 1,
        isAdditional: lotteryCode === "dlt" ? isAdditional : false,
      };
    })
    .filter((entry) => entry.red.length > 0 && entry.blue.length > 0);
}

function toggleEntryAdditionalText(value: string, index: number, lotteryCode: string) {
  if (lotteryCode !== "dlt") {
    return value;
  }

  const lines = value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

  return lines
    .map((line, lineIndex) => {
      if (lineIndex !== index) {
        return line;
      }

      const hasAdditional = line.includes("追加");
      const sourceLine = line.replace(/追加/g, "").trim();
      if (hasAdditional) {
        return sourceLine;
      }

      const multipleMatch = sourceLine.match(/[（(]\s*\d+\s*[)）]\s*$/);
      if (!multipleMatch) {
        return `${sourceLine} 追加`;
      }

      const multiplePart = multipleMatch[0];
      const numbersPart = sourceLine.slice(0, sourceLine.length - multiplePart.length).trim();
      return `${numbersPart} 追加 ${multiplePart}`.trim();
    })
    .join("\n");
}

function changeEntryMultipleText(value: string, index: number, nextMultiple: number) {
  const lines = value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

  return lines
    .map((line, lineIndex) => {
      if (lineIndex !== index) {
        return line;
      }

      const baseLine = line.replace(/[（(]\s*\d+\s*[)）]\s*$/, "").trim();
      return nextMultiple > 1 ? `${baseLine} (${nextMultiple})` : baseLine;
    })
    .join("\n");
}

function replaceRecommendation(
  items: Recommendation[],
  nextItem: Recommendation
) {
  const nextItems = items.slice();
  const index = nextItems.findIndex((item) => item.id === nextItem.id);
  if (index >= 0) {
    nextItems[index] = nextItem;
    return nextItems;
  }
  return [nextItem, ...nextItems];
}

export default function App() {
  const [activeTab, setActiveTab] = useState<TabKey>("dashboard");
  const [currentUser, setCurrentUser] = useState<AuthUser | null>(null);
  const [authUsername, setAuthUsername] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [authPending, setAuthPending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [detailPending, setDetailPending] = useState(false);
  const [uploadPending, setUploadPending] = useState(false);
  const [recognizePending, setRecognizePending] = useState(false);
  const [submitPending, setSubmitPending] = useState(false);
  const [recheckPending, setRecheckPending] = useState(false);
  const [deletePending, setDeletePending] = useState(false);
  const [generatingLotteryCode, setGeneratingLotteryCode] = useState("");
  const [recommendationLoading, setRecommendationLoading] = useState(false);
  const [recommendationLoadingMore, setRecommendationLoadingMore] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyLoadingMore, setHistoryLoadingMore] = useState(false);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [recommendationFilters, setRecommendationFilters] = useState<RecommendationFilters>(DEFAULT_RECOMMENDATION_FILTERS);
  const [recommendationPage, setRecommendationPage] = useState(1);
  const [recommendationHasMore, setRecommendationHasMore] = useState(false);
  const [recommendationTotal, setRecommendationTotal] = useState(0);
  const [tickets, setTickets] = useState<TicketRecord[]>([]);
  const [historyFilters, setHistoryFilters] = useState<TicketHistoryFilters>(DEFAULT_HISTORY_FILTERS);
  const [historyPage, setHistoryPage] = useState(1);
  const [historyHasMore, setHistoryHasMore] = useState(false);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [selectedRecommendation, setSelectedRecommendation] = useState<Recommendation | null>(null);
  const [selectedHistoryTicket, setSelectedHistoryTicket] = useState<TicketRecord | null>(null);
  const [purchaseRecommendation, setPurchaseRecommendation] = useState<Recommendation | null>(null);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [uploadedTicket, setUploadedTicket] = useState<TicketUpload | null>(null);
  const [recognitionDraft, setRecognitionDraft] = useState<TicketRecognitionDraft | null>(null);
  const [lotteryCode, setLotteryCode] = useState("");
  const [issue, setIssue] = useState("");
  const [drawDate, setDrawDate] = useState("");
  const [costAmount, setCostAmount] = useState("");
  const [costAmountEdited, setCostAmountEdited] = useState(false);
  const [notes, setNotes] = useState("");
  const [entryText, setEntryText] = useState("");

  function resetLotteryState() {
    setDashboard(null);
    setRecommendations([]);
    setRecommendationFilters(DEFAULT_RECOMMENDATION_FILTERS);
    setRecommendationPage(1);
    setRecommendationHasMore(false);
    setRecommendationTotal(0);
    setTickets([]);
    setHistoryFilters(DEFAULT_HISTORY_FILTERS);
    setHistoryPage(1);
    setHistoryHasMore(false);
    setHistoryTotal(0);
    setSelectedRecommendation(null);
    setSelectedHistoryTicket(null);
    setPurchaseRecommendation(null);
    setActiveTab("dashboard");
  }

  function resetTicketWorkflow() {
    setSelectedImage(null);
    setPreviewUrl("");
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setLotteryCode("");
    setIssue("");
    setDrawDate("");
    setCostAmount("");
    setCostAmountEdited(false);
    setNotes("");
    setEntryText("");
  }

  function handleRequestError(error: unknown, fallbackMessage: string) {
    if (!getStoredToken()) {
      setCurrentUser(null);
      resetLotteryState();
      resetTicketWorkflow();
    }
    toast.error(error instanceof Error ? error.message : fallbackMessage);
  }

  async function loadData(showLoading = true) {
    if (showLoading) {
      setLoading(true);
    }

    try {
      const [dashboardData, recommendationData, historyData] = await Promise.all([
        getDashboard(),
        getRecommendations(1, 10, recommendationFilters),
        getTicketHistory(1, 10, historyFilters),
      ]);
      setDashboard(dashboardData);
      setRecommendations(recommendationData.items);
      setRecommendationPage(recommendationData.page);
      setRecommendationHasMore(recommendationData.hasMore);
      setRecommendationTotal(recommendationData.total);
      setTickets(historyData.items);
      setHistoryPage(historyData.page);
      setHistoryHasMore(historyData.hasMore);
      setHistoryTotal(historyData.total);
    } catch (error) {
      handleRequestError(error, "加载数据失败");
    } finally {
      if (showLoading) {
        setLoading(false);
      }
    }
  }

  async function loadRecommendations(page: number, append: boolean, filters: RecommendationFilters) {
    if (append) {
      setRecommendationLoadingMore(true);
    } else {
      setRecommendationLoading(true);
    }

    try {
      const recommendationData = await getRecommendations(page, 10, filters);
      setRecommendations((current) =>
        append ? [...current, ...recommendationData.items] : recommendationData.items
      );
      setRecommendationPage(recommendationData.page);
      setRecommendationHasMore(recommendationData.hasMore);
      setRecommendationTotal(recommendationData.total);
    } catch (error) {
      handleRequestError(error, "加载推荐失败");
    } finally {
      if (append) {
        setRecommendationLoadingMore(false);
      } else {
        setRecommendationLoading(false);
      }
    }
  }

  async function loadHistoryTickets(page: number, append: boolean, filters: TicketHistoryFilters) {
    if (append) {
      setHistoryLoadingMore(true);
    } else {
      setHistoryLoading(true);
    }

    try {
      const historyData = await getTicketHistory(page, 10, filters);
      setTickets((current) => (append ? [...current, ...historyData.items] : historyData.items));
      setHistoryPage(historyData.page);
      setHistoryHasMore(historyData.hasMore);
      setHistoryTotal(historyData.total);
    } catch (error) {
      handleRequestError(error, "加载历史记录失败");
    } finally {
      if (append) {
        setHistoryLoadingMore(false);
      } else {
        setHistoryLoading(false);
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
        resetTicketWorkflow();
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

  useEffect(() => {
    if (costAmountEdited || !lotteryCode) {
      return;
    }

    const entries = parseEntryText(entryText, lotteryCode);
    if (entries.length === 0) {
      setCostAmount("");
      return;
    }

    setCostAmount(calculateEntriesCost(entries).toFixed(2));
  }, [costAmountEdited, entryText, lotteryCode]);

  async function handleAuthSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!authUsername.trim() || !authPassword.trim()) {
      toast.error("请输入用户名和密码");
      return;
    }

    setAuthPending(true);
    try {
      await loginAndStoreToken(authUsername.trim(), authPassword);
      const profile = await getProfile();
      setCurrentUser(profile);
      toast.success("登录成功");
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

  async function handleOpenRecommendation(recommendationId: string | null) {
    if (!recommendationId) {
      setSelectedRecommendation(null);
      return;
    }

    const currentItem = recommendations.find((item) => item.id === recommendationId) || null;
    if (!currentItem?.lotteryCode) {
      toast.error("未找到推荐对应的彩票类型");
      return;
    }
    setSelectedRecommendation(currentItem);
    setDetailPending(true);

    try {
      const detail = await getRecommendationDetail(currentItem.lotteryCode, recommendationId);
      setSelectedRecommendation(detail);
      setRecommendations((items) => replaceRecommendation(items, detail));
    } catch (error) {
      handleRequestError(error, "加载推荐详情失败");
    } finally {
      setDetailPending(false);
    }
  }

  function handleRecordPurchase(recommendation: Recommendation) {
    setSelectedRecommendation(null);
    setPurchaseRecommendation(recommendation);
    setLotteryCode(recommendation.lotteryCode || "");
    setIssue(recommendation.issue || "");
    setDrawDate(formatLotteryDrawDate(recommendation.drawDate));
    setEntryText(buildRecommendationEntryText(recommendation));
    setCostAmountEdited(false);
    setActiveTab("records");
    toast.success("已切换到记录页，接下来上传这条推荐的购买票据即可");
  }

  function handleSelectImage(file: File | null) {
    setSelectedImage(file);
    setUploadedTicket(null);
    setRecognitionDraft(null);
  }

  async function uploadSelectedTicketImage() {
    if (!selectedImage) {
      throw new Error("请先选择彩票照片");
    }
    const formData = new FormData();
    formData.append("image", selectedImage);

    setUploadPending(true);
    try {
      const upload = await uploadTicketImage(formData);
      setUploadedTicket(upload);
      return upload;
    } catch (error) {
      handleRequestError(error, "上传原图失败");
      throw error;
    } finally {
      setUploadPending(false);
    }
  }

  async function handleRecognizeTicket() {
    if (!selectedImage && !uploadedTicket) {
      toast.error("请先选择彩票照片");
      return;
    }

    setRecognizePending(true);
    try {
      const upload = uploadedTicket || (await uploadSelectedTicketImage());
      const draft = await recognizeTicket({
        uploadId: upload.id,
      });
      setUploadedTicket(draft.upload);
      setRecognitionDraft(draft);
      setLotteryCode(draft.lotteryCode || lotteryCode);
      setIssue(draft.issue || "");
      setDrawDate(draft.drawDate || "");
      setCostAmount(draft.costAmount > 0 ? draft.costAmount.toFixed(2) : "");
      setCostAmountEdited(false);
      setEntryText(buildParsedEntryText(draft.entries, draft.lotteryCode || lotteryCode));
      toast.success("识别完成");
    } catch (error) {
      handleRequestError(error, "识别失败");
    } finally {
      setRecognizePending(false);
    }
  }

  async function handleCreateTicket() {
    if (!selectedImage && !uploadedTicket) {
      toast.error("请先选择彩票照片");
      return;
    }

    const entries = parseEntryText(entryText, lotteryCode);
    if (!lotteryCode.trim()) {
      toast.error("请确认彩票类型");
      return;
    }
    if (!issue.trim()) {
      toast.error("请填写期号");
      return;
    }
    if (!drawDate.trim()) {
      toast.error("请填写开奖日期");
      return;
    }
    if (entries.length === 0) {
      toast.error("请至少填写一注号码");
      return;
    }
    const manualCostAmount = Number(costAmount);
    if (!(Number.isFinite(manualCostAmount) && manualCostAmount > 0)) {
      toast.error("请填写正确的金额");
      return;
    }

    setSubmitPending(true);
    try {
      const upload = uploadedTicket || (await uploadSelectedTicketImage());
      await createTicket({
        lotteryCode: lotteryCode || undefined,
        uploadId: upload.id,
        recommendationId: purchaseRecommendation?.id,
        issue: issue || undefined,
        drawDate: drawDate || undefined,
        purchasedAt: new Date().toISOString(),
        costAmount: Number.isFinite(manualCostAmount) && manualCostAmount > 0 ? manualCostAmount : undefined,
        notes: notes || undefined,
        entries: entries.map((entry) => ({
          ...formatParsedEntry(entry),
          multiple: entry.multiple,
          isAdditional: entry.isAdditional,
        })),
      });
      toast.success("票据已入库并完成判奖检查");
      resetTicketWorkflow();
      setPurchaseRecommendation(null);
      await loadData(false);
    } catch (error) {
      handleRequestError(error, "入库失败");
    } finally {
      setSubmitPending(false);
    }
  }

  function handleToggleEntryAdditional(index: number) {
    setEntryText((current) => toggleEntryAdditionalText(current, index, lotteryCode));
  }

  function handleChangeEntryMultiple(index: number, nextMultiple: number) {
    setEntryText((current) => changeEntryMultipleText(current, index, nextMultiple));
  }

  function handleChangeLotteryCode(value: string) {
    setLotteryCode(value);
    if (!costAmountEdited) {
      const entries = parseEntryText(entryText, value);
      setCostAmount(entries.length > 0 ? calculateEntriesCost(entries).toFixed(2) : "");
    }
  }

  function handleChangeCostAmount(value: string) {
    setCostAmount(value);
    setCostAmountEdited(true);
  }

  async function handleRecheckTicket(ticketId: string) {
    setRecheckPending(true);
    try {
      const updatedTicket = await recheckTicket(ticketId);
      setTickets((current) => current.map((item) => (item.id === updatedTicket.id ? updatedTicket : item)));
      setSelectedHistoryTicket(updatedTicket);
      toast.success("重新判奖完成");
    } catch (error) {
      handleRequestError(error, "重新判奖失败");
    } finally {
      setRecheckPending(false);
    }
  }

  async function handleDeleteTicket(ticket: TicketRecord) {
    if (!window.confirm(`确认删除第 ${ticket.issue} 期的这条记录吗？相关号码、上传图片记录会一并删除。`)) {
      return;
    }

    setDeletePending(true);
    try {
      await deleteTicket(ticket.id);
      setSelectedHistoryTicket(null);
      if (selectedRecommendation?.id && ticket.recommendationId === selectedRecommendation.id) {
        setSelectedRecommendation(null);
      }
      toast.success("记录已删除");
      await loadData(false);
    } catch (error) {
      handleRequestError(error, "删除记录失败");
    } finally {
      setDeletePending(false);
    }
  }

  async function handleDeleteRecommendation(recommendation: Recommendation) {
    if (
      !window.confirm(
        `确认删除第 ${recommendation.issue} 期推荐吗？已录入的购买记录会保留，但不再关联这条推荐。`
      )
    ) {
      return;
    }

    setDeletePending(true);
    try {
      await deleteRecommendation(recommendation.lotteryCode, recommendation.id);
      setSelectedRecommendation(null);
      if (purchaseRecommendation?.id === recommendation.id) {
        setPurchaseRecommendation(null);
      }
      toast.success("推荐已删除，购买记录已保留");
      await loadData(false);
    } catch (error) {
      handleRequestError(error, "删除推荐失败");
    } finally {
      setDeletePending(false);
    }
  }

  async function handleGenerateRecommendation(lotteryCode: string) {
    setGeneratingLotteryCode(lotteryCode);
    try {
      const recommendation = await generateRecommendation(lotteryCode);
      const nextFilters = DEFAULT_RECOMMENDATION_FILTERS;
      const [dashboardData, detail] = await Promise.all([
        getDashboard(),
        getRecommendationDetail(lotteryCode, recommendation.id),
      ]);
      setDashboard(dashboardData);
      setRecommendationFilters(nextFilters);
      await loadRecommendations(1, false, nextFilters);
      setSelectedRecommendation(detail);
      setActiveTab("recommendation");
      toast.success(`${recommendation.lotteryCode === "dlt" ? "大乐透" : "双色球"}推荐已生成`);
    } catch (error) {
      handleRequestError(error, "生成推荐失败");
    } finally {
      setGeneratingLotteryCode("");
    }
  }

  function handleHistoryFilterChange(nextFilters: TicketHistoryFilters) {
    setHistoryFilters(nextFilters);
    setSelectedHistoryTicket(null);
    void loadHistoryTickets(1, false, nextFilters);
  }

  function handleRecommendationFilterChange(nextFilters: RecommendationFilters) {
    setRecommendationFilters(nextFilters);
    setSelectedRecommendation(null);
    void loadRecommendations(1, false, nextFilters);
  }

  function handleLoadMoreHistory() {
    if (!historyHasMore || historyLoadingMore) {
      return;
    }
    void loadHistoryTickets(historyPage + 1, true, historyFilters);
  }

  function handleLoadMoreRecommendations() {
    if (!recommendationHasMore || recommendationLoadingMore) {
      return;
    }
    void loadRecommendations(recommendationPage + 1, true, recommendationFilters);
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
        username={authUsername}
        password={authPassword}
        pending={authPending}
        onUsernameChange={setAuthUsername}
        onPasswordChange={setAuthPassword}
        onSubmit={handleAuthSubmit}
      />
    );
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.84),rgba(241,245,249,0.9)_40%,rgba(226,232,240,1))] pb-28">
      <div className="mx-auto w-full max-w-6xl px-4 pb-10 pt-4 sm:px-6">
        {activeTab === "dashboard" && (
          <DashboardPanel
            currentUser={currentUser}
            dashboard={dashboard}
            generatingLotteryCode={generatingLotteryCode}
            onGenerateRecommendation={(lotteryCode) => void handleGenerateRecommendation(lotteryCode)}
            onLogout={handleLogout}
          />
        )}

        {activeTab === "recommendation" && (
          <RecommendationPanel
            recommendations={recommendations}
            filters={recommendationFilters}
            loading={recommendationLoading}
            loadingMore={recommendationLoadingMore}
            hasMore={recommendationHasMore}
            total={recommendationTotal}
            selectedRecommendation={selectedRecommendation}
            detailPending={detailPending}
            deletePending={deletePending}
            onFiltersChange={handleRecommendationFilterChange}
            onLoadMore={handleLoadMoreRecommendations}
            onSelectRecommendation={(recommendationId) => void handleOpenRecommendation(recommendationId)}
            onRecordPurchase={handleRecordPurchase}
            onDeleteRecommendation={(recommendation) => void handleDeleteRecommendation(recommendation)}
          />
        )}

        {activeTab === "records" && (
          <RecordPanel
            selectedRecommendation={purchaseRecommendation}
            previewUrl={previewUrl}
            selectedImage={selectedImage}
            uploadPending={uploadPending}
            uploadedTicket={uploadedTicket}
            recognitionDraft={recognitionDraft}
            lotteryCode={lotteryCode}
            recognizePending={recognizePending}
            issue={issue}
            drawDate={drawDate}
            costAmount={costAmount}
            notes={notes}
            entryText={entryText}
            submitPending={submitPending}
            onSelectImage={handleSelectImage}
            onLotteryCodeChange={handleChangeLotteryCode}
            onRecognize={() => void handleRecognizeTicket()}
            onIssueChange={setIssue}
            onDrawDateChange={setDrawDate}
            onCostAmountChange={handleChangeCostAmount}
            onNotesChange={setNotes}
            onEntryTextChange={setEntryText}
            onToggleEntryAdditional={handleToggleEntryAdditional}
            onChangeEntryMultiple={handleChangeEntryMultiple}
            onCreateTicket={() => void handleCreateTicket()}
            onClearRecommendation={() => setPurchaseRecommendation(null)}
          />
        )}

        {activeTab === "history" && (
          <HistoryPanel
            tickets={tickets}
            filters={historyFilters}
            loading={historyLoading}
            loadingMore={historyLoadingMore}
            hasMore={historyHasMore}
            total={historyTotal}
            selectedTicket={selectedHistoryTicket}
            recheckPending={recheckPending}
            deletePending={deletePending}
            onFiltersChange={handleHistoryFilterChange}
            onLoadMore={handleLoadMoreHistory}
            onSelectTicket={setSelectedHistoryTicket}
            onRecheckTicket={(ticketId) => void handleRecheckTicket(ticketId)}
            onDeleteTicket={(ticket) => void handleDeleteTicket(ticket)}
          />
        )}
      </div>

      <nav className="fixed inset-x-0 bottom-4 z-50 mx-auto flex w-[calc(100%-1.5rem)] max-w-5xl items-center gap-2 rounded-[2rem] border border-white/70 bg-white/90 p-2 shadow-[0_18px_40px_rgba(15,23,42,0.14)] backdrop-blur">
        {tabs.map((item) => {
          const Icon = item.icon;
          const active = activeTab === item.key;
          return (
            <button
              key={item.key}
              type="button"
              className={`flex flex-1 items-center justify-center gap-2 rounded-[1.25rem] px-3 py-3 text-sm transition ${
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
