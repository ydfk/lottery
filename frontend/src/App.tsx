import { useEffect, useState, type FormEvent } from "react";
import { History, LayoutDashboard, List, ReceiptText, Sparkles } from "lucide-react";
import { toast } from "sonner";
import { AuthPanel } from "@/components/lottery/auth-panel";
import { DashboardPanel } from "@/components/lottery/dashboard-panel";
import { DrawResultsPanel } from "@/components/lottery/draw-results-panel";
import {
  LotteryDisplayModeToggle,
  type LotteryDisplayMode,
} from "@/components/lottery/display-mode-toggle";
import { HistoryPanel } from "@/components/lottery/history-panel";
import { LotteryShell, type LotteryShellTab } from "@/components/lottery/lottery-shell";
import { RecommendationPanel } from "@/components/lottery/recommendation-panel";
import { RecordPanel } from "@/components/lottery/record-panel";
import { isTrendLotteryCode, useLotteryNumberTrends } from "@/hooks/use-lottery-number-trends";
import { getStoredToken } from "@/lib/api/client";
import { getProfile, loginAndStoreToken, logout } from "@/lib/api/methods/auth";
import {
  createTicket,
  deleteRecommendation,
  deleteTicket,
  formatParsedEntry,
  getDashboard,
  getDrawResults,
  importTickets,
  getRecommendationDetail,
  getRecommendations,
  generateRecommendation,
  getTicketHistory,
  recheckRecommendation,
  recheckTicket,
  recognizeTicket,
  syncDrawIssue,
  updateTicket,
  uploadTicketImage,
} from "@/lib/api/methods/lottery";
import { formatLotteryDrawDate } from "@/lib/lottery-display";
import { calculateEntriesCost } from "@/lib/lottery-cost";
import {
  buildDraftsFromParsedEntries,
  buildDraftsFromRecommendationEntries,
  buildDraftsFromTicketEntries,
  buildParsedEntriesFromDrafts,
  createEmptyEntryDraft,
  normalizeDraftsForLottery,
} from "@/lib/ticket-entry-drafts";
import type { AuthUser } from "@/types/auth";
import type {
  DashboardData,
  DrawResult as DrawResultItem,
  DrawResultFilters,
  Recommendation,
  RecommendationFilters,
  Ticket as TicketRecord,
  TicketEntryDraft,
  TicketHistoryFilters,
  TicketRecognitionDraft,
  TicketUpload,
} from "@/types/lottery";

type TabKey = "dashboard" | "recommendation" | "records" | "history" | "draws";

const DISPLAY_MODE_STORAGE_KEY = "lottery:display-mode";
const DRAW_PAGE_SIZE = 20;
const RECOMMENDATION_PAGE_SIZE = 20;
const HISTORY_PAGE_SIZE = 20;

const tabs: LotteryShellTab<TabKey>[] = [
  {
    key: "dashboard",
    label: "看板",
    description: "查看账户与统计概览",
    icon: LayoutDashboard,
  },
  {
    key: "recommendation",
    label: "推荐",
    description: "浏览推荐结果与购买动作",
    icon: Sparkles,
  },
  {
    key: "records",
    label: "记录",
    description: "手动录入号码，可选图片识别",
    icon: ReceiptText,
  },
  {
    key: "history",
    label: "历史",
    description: "追踪已购记录与中奖状态",
    icon: History,
  },
  {
    key: "draws",
    label: "开奖",
    description: "查看历史爬取到的开奖记录",
    icon: List,
  },
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

const DEFAULT_DRAW_FILTERS: DrawResultFilters = {
  lotteryCode: "",
  issue: "",
  drawDate: "",
  sort: "latest",
};

function replaceRecommendation(items: Recommendation[], nextItem: Recommendation) {
  const nextItems = items.slice();
  const index = nextItems.findIndex((item) => item.id === nextItem.id);
  if (index >= 0) {
    nextItems[index] = nextItem;
    return nextItems;
  }
  return [nextItem, ...nextItems];
}

function getStoredDisplayMode(): LotteryDisplayMode {
  if (typeof window === "undefined") {
    return "app";
  }

  return window.localStorage.getItem(DISPLAY_MODE_STORAGE_KEY) === "web" ? "web" : "app";
}

function formatDateTimeInput(value?: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const offsetDate = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return offsetDate.toISOString().slice(0, 16);
}

function toISODateTime(value: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}

export default function App() {
  const [displayMode, setDisplayMode] = useState<LotteryDisplayMode>(() => getStoredDisplayMode());
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
  const [recommendationRecheckPendingId, setRecommendationRecheckPendingId] = useState("");
  const [deletePending, setDeletePending] = useState(false);
  const [generatingLotteryCode, setGeneratingLotteryCode] = useState("");
  const [recommendationLoading, setRecommendationLoading] = useState(false);
  const [recommendationLoadingMore, setRecommendationLoadingMore] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyLoadingMore, setHistoryLoadingMore] = useState(false);
  const [drawLoading, setDrawLoading] = useState(false);
  const [drawCompletePendingId, setDrawCompletePendingId] = useState("");
  const [drawSyncPending, setDrawSyncPending] = useState(false);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [drawResults, setDrawResults] = useState<DrawResultItem[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [recommendationFilters, setRecommendationFilters] = useState<RecommendationFilters>(
    DEFAULT_RECOMMENDATION_FILTERS
  );
  const [recommendationPage, setRecommendationPage] = useState(1);
  const [recommendationHasMore, setRecommendationHasMore] = useState(false);
  const [recommendationTotal, setRecommendationTotal] = useState(0);
  const [tickets, setTickets] = useState<TicketRecord[]>([]);
  const [historyFilters, setHistoryFilters] =
    useState<TicketHistoryFilters>(DEFAULT_HISTORY_FILTERS);
  const [historyPage, setHistoryPage] = useState(1);
  const [historyHasMore, setHistoryHasMore] = useState(false);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [drawFilters, setDrawFilters] = useState<DrawResultFilters>(DEFAULT_DRAW_FILTERS);
  const [drawPage, setDrawPage] = useState(1);
  const [drawTotal, setDrawTotal] = useState(0);
  const [selectedRecommendation, setSelectedRecommendation] = useState<Recommendation | null>(null);
  const [selectedHistoryTicket, setSelectedHistoryTicket] = useState<TicketRecord | null>(null);
  const [editingTicket, setEditingTicket] = useState<TicketRecord | null>(null);
  const [purchaseRecommendation, setPurchaseRecommendation] = useState<Recommendation | null>(null);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [uploadedTicket, setUploadedTicket] = useState<TicketUpload | null>(null);
  const [recognitionDraft, setRecognitionDraft] = useState<TicketRecognitionDraft | null>(null);
  const [lotteryCode, setLotteryCode] = useState("");
  const [issue, setIssue] = useState("");
  const [drawDate, setDrawDate] = useState("");
  const [purchasedAt, setPurchasedAt] = useState("");
  const [costAmount, setCostAmount] = useState("");
  const [notes, setNotes] = useState("");
  const [entryDrafts, setEntryDrafts] = useState<TicketEntryDraft[]>([createEmptyEntryDraft()]);
  const webNavigationTabs = tabs.filter((item) => item.key !== "records");
  const appNavigationTabs = tabs.filter((item) => item.key !== "draws");
  const navigationTabs = displayMode === "web" ? webNavigationTabs : appNavigationTabs;
  const navigationActiveTab =
    displayMode === "web" && activeTab === "records"
      ? "history"
      : displayMode === "app" && activeTab === "draws"
        ? "dashboard"
        : activeTab;
  const { trends: numberTrends, reloadTrend } = useLotteryNumberTrends(
    Boolean(currentUser) && displayMode === "web" && activeTab === "draws"
  );

  function resetLotteryState() {
    setDashboard(null);
    setDrawResults([]);
    setDrawFilters(DEFAULT_DRAW_FILTERS);
    setDrawPage(1);
    setDrawTotal(0);
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
    setEditingTicket(null);
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
    setPurchasedAt("");
    setCostAmount("");
    setNotes("");
    setEntryDrafts([createEmptyEntryDraft()]);
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
      const [dashboardResult, drawResult, recommendationResult, historyResult] =
        await Promise.allSettled([
          getDashboard(),
          getDrawResults(1, DRAW_PAGE_SIZE, drawFilters),
          getRecommendations(1, RECOMMENDATION_PAGE_SIZE, recommendationFilters),
          getTicketHistory(1, HISTORY_PAGE_SIZE, historyFilters),
        ]);

      if (dashboardResult.status === "fulfilled") {
        setDashboard(dashboardResult.value);
      }
      if (drawResult.status === "fulfilled") {
        setDrawResults(drawResult.value.items);
        setDrawPage(drawResult.value.page);
        setDrawTotal(drawResult.value.total);
      }
      if (recommendationResult.status === "fulfilled") {
        setRecommendations(recommendationResult.value.items);
        setRecommendationPage(recommendationResult.value.page);
        setRecommendationHasMore(recommendationResult.value.hasMore);
        setRecommendationTotal(recommendationResult.value.total);
      }
      if (historyResult.status === "fulfilled") {
        setTickets(historyResult.value.items);
        setHistoryPage(historyResult.value.page);
        setHistoryHasMore(historyResult.value.hasMore);
        setHistoryTotal(historyResult.value.total);
      }

      const failedResult = [dashboardResult, drawResult, recommendationResult, historyResult].find(
        (result) => result.status === "rejected"
      );
      if (failedResult?.status === "rejected") {
        handleRequestError(failedResult.reason, "部分数据加载失败");
      }
    } catch (error) {
      handleRequestError(error, "加载数据失败");
    } finally {
      if (showLoading) {
        setLoading(false);
      }
    }
  }

  async function loadDrawResults(page: number, filters: DrawResultFilters) {
    setDrawLoading(true);

    try {
      const drawData = await getDrawResults(page, DRAW_PAGE_SIZE, filters);
      setDrawResults(drawData.items);
      setDrawPage(drawData.page);
      setDrawTotal(drawData.total);
    } catch (error) {
      handleRequestError(error, "加载开奖记录失败");
    } finally {
      setDrawLoading(false);
    }
  }

  async function loadRecommendations(
    page: number,
    append: boolean,
    filters: RecommendationFilters
  ) {
    if (append) {
      setRecommendationLoadingMore(true);
    } else {
      setRecommendationLoading(true);
    }

    try {
      const recommendationData = await getRecommendations(page, RECOMMENDATION_PAGE_SIZE, filters);
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
      const historyData = await getTicketHistory(page, HISTORY_PAGE_SIZE, filters);
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
    window.localStorage.setItem(DISPLAY_MODE_STORAGE_KEY, displayMode);
  }, [displayMode]);

  useEffect(() => {
    if (displayMode === "app" && activeTab === "draws") {
      setActiveTab("dashboard");
    }
  }, [activeTab, displayMode]);

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
    const entries = buildParsedEntriesFromDrafts(entryDrafts, lotteryCode);
    if (entries.length === 0) {
      setCostAmount("");
      return;
    }

    setCostAmount(calculateEntriesCost(entries).toFixed(2));
  }, [entryDrafts, lotteryCode]);

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
    setEditingTicket(null);
    setPurchaseRecommendation(recommendation);
    setLotteryCode(recommendation.lotteryCode || "");
    setIssue(recommendation.issue || "");
    setDrawDate(formatLotteryDrawDate(recommendation.drawDate));
    setEntryDrafts(buildDraftsFromRecommendationEntries(recommendation.entries));
    setActiveTab("records");
    toast.success("已切换到记录页，可手动确认购买信息后保存");
  }

  function handleEditTicket(ticket: TicketRecord) {
    setSelectedHistoryTicket(null);
    setPurchaseRecommendation(null);
    setEditingTicket(ticket);
    setSelectedImage(null);
    setPreviewUrl("");
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setLotteryCode(ticket.lotteryCode || "");
    setIssue(ticket.issue || "");
    setDrawDate(formatLotteryDrawDate(ticket.drawDate || ticket.manualDrawDate));
    setPurchasedAt(formatDateTimeInput(ticket.purchasedAt));
    setNotes(ticket.notes || "");
    setEntryDrafts(buildDraftsFromTicketEntries(ticket.entries));
    setActiveTab("records");
    toast.success("已进入历史记录编辑");
  }

  function handleCancelEditTicket() {
    setEditingTicket(null);
    resetTicketWorkflow();
    setActiveTab("history");
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
      setEntryDrafts(buildDraftsFromParsedEntries(draft.entries));
      toast.success("识别完成");
    } catch (error) {
      handleRequestError(error, "识别失败");
    } finally {
      setRecognizePending(false);
    }
  }

  async function handleCreateTicket() {
    const entries = buildParsedEntriesFromDrafts(entryDrafts, lotteryCode);
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
    const calculatedCostAmount = Number(costAmount);
    if (!(Number.isFinite(calculatedCostAmount) && calculatedCostAmount > 0)) {
      toast.error("请完善投注号码以计算金额");
      return;
    }

    setSubmitPending(true);
    try {
      const upload = uploadedTicket || (selectedImage ? await uploadSelectedTicketImage() : null);
      const payload = {
        lotteryCode: lotteryCode || undefined,
        recommendationId: editingTicket?.recommendationId || purchaseRecommendation?.id,
        issue: issue || undefined,
        drawDate: drawDate || undefined,
        purchasedAt: toISODateTime(purchasedAt) || new Date().toISOString(),
        costAmount: calculatedCostAmount,
        notes: notes || undefined,
        entries: entries.map((entry) => ({
          ...formatParsedEntry(entry),
          multiple: entry.multiple,
          isAdditional: entry.isAdditional,
        })),
      };
      if (editingTicket) {
        await updateTicket(editingTicket.id, payload);
        toast.success("历史记录已更新");
      } else {
        await createTicket({
          ...payload,
          uploadId: upload?.id,
        });
        toast.success("票据已入库并完成判奖检查");
      }
      resetTicketWorkflow();
      setEditingTicket(null);
      setPurchaseRecommendation(null);
      await loadData(false);
      if (editingTicket) {
        setActiveTab("history");
      }
    } catch (error) {
      handleRequestError(error, "入库失败");
    } finally {
      setSubmitPending(false);
    }
  }

  function handleChangeLotteryCode(value: string) {
    setLotteryCode(value);
    setEntryDrafts((current) => normalizeDraftsForLottery(current, value));
  }

  function handleChangeEntryField(
    index: number,
    field: "redNumbers" | "blueNumbers",
    value: string
  ) {
    setEntryDrafts((current) =>
      current.map((entry, entryIndex) =>
        entryIndex === index ? { ...entry, [field]: value } : entry
      )
    );
  }

  function handleToggleEntryAdditional(index: number) {
    if (lotteryCode !== "dlt") {
      return;
    }

    setEntryDrafts((current) =>
      current.map((entry, entryIndex) =>
        entryIndex === index ? { ...entry, isAdditional: !entry.isAdditional } : entry
      )
    );
  }

  function handleChangeEntryMultiple(index: number, nextMultiple: number) {
    setEntryDrafts((current) =>
      current.map((entry, entryIndex) =>
        entryIndex === index
          ? { ...entry, multiple: Math.max(1, Math.min(99, nextMultiple)) }
          : entry
      )
    );
  }

  function handleAddEntry() {
    setEntryDrafts((current) => [createEmptyEntryDraft(), ...current]);
  }

  function handleRemoveEntry(index: number) {
    setEntryDrafts((current) => {
      if (current.length <= 1) {
        return [createEmptyEntryDraft()];
      }
      return current.filter((_, entryIndex) => entryIndex !== index);
    });
  }

  async function handleRecheckTicket(ticketId: string) {
    setRecheckPending(true);
    try {
      const updatedTicket = await recheckTicket(ticketId);
      setTickets((current) =>
        current.map((item) => (item.id === updatedTicket.id ? updatedTicket : item))
      );
      setSelectedHistoryTicket(updatedTicket);
      toast.success("重新判奖完成");
    } catch (error) {
      handleRequestError(error, "重新判奖失败");
    } finally {
      setRecheckPending(false);
    }
  }

  async function handleDeleteTicket(ticket: TicketRecord) {
    if (
      !window.confirm(
        `确认删除第 ${ticket.issue} 期的这条记录吗？相关号码、上传图片记录会一并删除。`
      )
    ) {
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

  async function handleRecheckRecommendation(recommendation: Recommendation) {
    setRecommendationRecheckPendingId(recommendation.id);
    try {
      const updatedRecommendation = await recheckRecommendation(
        recommendation.lotteryCode,
        recommendation.id
      );
      setRecommendations((items) => replaceRecommendation(items, updatedRecommendation));
      if (selectedRecommendation?.id === updatedRecommendation.id) {
        setSelectedRecommendation(updatedRecommendation);
      }
      toast.success("推荐判奖已更新");
    } catch (error) {
      handleRequestError(error, "推荐重新判奖失败");
    } finally {
      setRecommendationRecheckPendingId("");
    }
  }

  async function handleCompleteDrawInfo(draw: DrawResultItem) {
    setDrawCompletePendingId(draw.id);
    try {
      const result = await syncDrawIssue(draw.lotteryCode, draw.issue);
      if (result.syncedCount <= 0) {
        throw new Error("第三方接口暂未返回当前期完整开奖信息");
      }
      await Promise.all([
        loadDrawResults(drawPage, drawFilters),
        isTrendLotteryCode(draw.lotteryCode) ? reloadTrend(draw.lotteryCode) : Promise.resolve(),
      ]);
      toast.success("开奖信息已补全");
    } catch (error) {
      handleRequestError(error, "补全开奖信息失败");
    } finally {
      setDrawCompletePendingId("");
    }
  }

  async function handleSyncDrawIssue(lotteryCode: string, issue: string) {
    setDrawSyncPending(true);
    try {
      const result = await syncDrawIssue(lotteryCode, issue);
      const nextFilters = {
        lotteryCode,
        issue,
        drawDate: "",
        sort: "latest",
      };
      setDrawFilters(nextFilters);
      await Promise.all([
        loadDrawResults(1, nextFilters),
        isTrendLotteryCode(lotteryCode) ? reloadTrend(lotteryCode) : Promise.resolve(),
      ]);
      if (result.syncedCount > 0) {
        toast.success("开奖数据已同步");
      } else {
        toast.info("第三方暂未返回该期开奖数据");
      }
    } catch (error) {
      handleRequestError(error, "同步开奖数据失败");
    } finally {
      setDrawSyncPending(false);
    }
  }

  async function handleImportTickets(workbook: File, imagesZip: File | null) {
    const formData = new FormData();
    formData.append("workbook", workbook);
    if (imagesZip) {
      formData.append("imagesZip", imagesZip);
    }

    try {
      const result = await importTickets(formData);
      setSelectedHistoryTicket(null);
      await loadData(false);
      toast.success(
        result.failedCount > 0
          ? `导入完成，成功 ${result.successCount} 行，失败 ${result.failedCount} 行`
          : `导入完成，共成功导入 ${result.successCount} 行`
      );
      return result;
    } catch (error) {
      handleRequestError(error, "导入 Excel 失败");
      throw error;
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

  function handleHistoryPageChange(nextPage: number) {
    if (nextPage === historyPage || historyLoading) {
      return;
    }
    setSelectedHistoryTicket(null);
    void loadHistoryTickets(nextPage, false, historyFilters);
  }

  function handleLoadMoreRecommendations() {
    if (!recommendationHasMore || recommendationLoadingMore) {
      return;
    }
    void loadRecommendations(recommendationPage + 1, true, recommendationFilters);
  }

  function handleRecommendationPageChange(nextPage: number) {
    if (nextPage === recommendationPage || recommendationLoading) {
      return;
    }
    setSelectedRecommendation(null);
    void loadRecommendations(nextPage, false, recommendationFilters);
  }

  function handleDrawFilterChange(nextFilters: DrawResultFilters) {
    setDrawFilters(nextFilters);
    void loadDrawResults(1, nextFilters);
  }

  function handleDrawPageChange(nextPage: number) {
    if (nextPage === drawPage || drawLoading) {
      return;
    }
    void loadDrawResults(nextPage, drawFilters);
  }

  function handleOpenRecordPanel() {
    setSelectedHistoryTicket(null);
    setEditingTicket(null);
    setActiveTab("records");
  }

  function renderActivePanel(user: AuthUser) {
    if (activeTab === "dashboard") {
      return (
        <DashboardPanel
          currentUser={user}
          dashboard={dashboard}
          displayMode={displayMode}
          generatingLotteryCode={generatingLotteryCode}
          onDisplayModeChange={setDisplayMode}
          onGenerateRecommendation={(lotteryCode) => void handleGenerateRecommendation(lotteryCode)}
          onLogout={handleLogout}
        />
      );
    }

    if (activeTab === "recommendation") {
      return (
        <RecommendationPanel
          displayMode={displayMode}
          recommendations={recommendations}
          filters={recommendationFilters}
          loading={recommendationLoading}
          loadingMore={recommendationLoadingMore}
          hasMore={recommendationHasMore}
          page={recommendationPage}
          pageSize={RECOMMENDATION_PAGE_SIZE}
          total={recommendationTotal}
          selectedRecommendation={selectedRecommendation}
          detailPending={detailPending}
          recheckPendingId={recommendationRecheckPendingId}
          deletePending={deletePending}
          onFiltersChange={handleRecommendationFilterChange}
          onLoadMore={handleLoadMoreRecommendations}
          onPageChange={handleRecommendationPageChange}
          onSelectRecommendation={(recommendationId) =>
            void handleOpenRecommendation(recommendationId)
          }
          onRecordPurchase={handleRecordPurchase}
          onRecheckRecommendation={(recommendation) =>
            void handleRecheckRecommendation(recommendation)
          }
          onDeleteRecommendation={(recommendation) =>
            void handleDeleteRecommendation(recommendation)
          }
        />
      );
    }

    if (activeTab === "records") {
      return (
        <RecordPanel
          displayMode={displayMode}
          selectedRecommendation={purchaseRecommendation}
          mode={editingTicket ? "edit" : "create"}
          previewUrl={previewUrl}
          existingImageUrl={editingTicket?.imageUrl}
          selectedImage={selectedImage}
          uploadPending={uploadPending}
          uploadedTicket={uploadedTicket}
          recognitionDraft={recognitionDraft}
          lotteryCode={lotteryCode}
          recognizePending={recognizePending}
          issue={issue}
          drawDate={drawDate}
          purchasedAt={purchasedAt}
          costAmount={costAmount}
          notes={notes}
          entryDrafts={entryDrafts}
          submitPending={submitPending}
          onSelectImage={handleSelectImage}
          onLotteryCodeChange={handleChangeLotteryCode}
          onRecognize={() => void handleRecognizeTicket()}
          onIssueChange={setIssue}
          onDrawDateChange={setDrawDate}
          onPurchasedAtChange={setPurchasedAt}
          onNotesChange={setNotes}
          onEntryFieldChange={handleChangeEntryField}
          onToggleEntryAdditional={handleToggleEntryAdditional}
          onChangeEntryMultiple={handleChangeEntryMultiple}
          onAddEntry={handleAddEntry}
          onRemoveEntry={handleRemoveEntry}
          onCreateTicket={() => void handleCreateTicket()}
          onCancelEdit={editingTicket ? handleCancelEditTicket : undefined}
          onClearRecommendation={() => setPurchaseRecommendation(null)}
        />
      );
    }

    if (activeTab === "draws") {
      return (
        <DrawResultsPanel
          displayMode={displayMode}
          items={drawResults}
          numberTrends={numberTrends}
          filters={drawFilters}
          loading={drawLoading}
          page={drawPage}
          pageSize={DRAW_PAGE_SIZE}
          total={drawTotal}
          completePendingId={drawCompletePendingId}
          syncPending={drawSyncPending}
          onFiltersChange={handleDrawFilterChange}
          onPageChange={handleDrawPageChange}
          onRetryTrend={(lotteryCode) => void reloadTrend(lotteryCode)}
          onSyncIssue={(lotteryCode, issue) => void handleSyncDrawIssue(lotteryCode, issue)}
          onCompleteDraw={(draw) => void handleCompleteDrawInfo(draw)}
        />
      );
    }

    return (
      <HistoryPanel
        displayMode={displayMode}
        tickets={tickets}
        filters={historyFilters}
        loading={historyLoading}
        loadingMore={historyLoadingMore}
        hasMore={historyHasMore}
        page={historyPage}
        pageSize={HISTORY_PAGE_SIZE}
        total={historyTotal}
        selectedTicket={selectedHistoryTicket}
        recheckPending={recheckPending}
        deletePending={deletePending}
        onFiltersChange={handleHistoryFilterChange}
        onLoadMore={handleLoadMoreHistory}
        onPageChange={handleHistoryPageChange}
        onImportTickets={handleImportTickets}
        onOpenRecord={handleOpenRecordPanel}
        onSelectTicket={setSelectedHistoryTicket}
        onEditTicket={handleEditTicket}
        onRecheckTicket={(ticketId) => void handleRecheckTicket(ticketId)}
        onDeleteTicket={(ticket) => void handleDeleteTicket(ticket)}
      />
    );
  }

  if (!currentUser) {
    if (loading) {
      return (
        <div
          className={`min-h-screen px-4 py-6 sm:px-6 ${
            displayMode === "web"
              ? "bg-[linear-gradient(180deg,rgba(248,250,252,0.96),rgba(226,232,240,0.88)),radial-gradient(circle_at_top_right,rgba(59,130,246,0.14),transparent_38%)]"
              : "bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.92),rgba(226,232,240,0.9)_45%,rgba(203,213,225,1))]"
          }`}
        >
          <div
            className={`mx-auto flex w-full ${
              displayMode === "web" ? "max-w-6xl" : "max-w-md"
            } justify-end`}
          >
            {displayMode === "web" ? (
              <LotteryDisplayModeToggle value={displayMode} onValueChange={setDisplayMode} />
            ) : (
              <LotteryDisplayModeToggle
                value={displayMode}
                compact
                className="bg-white/84"
                onValueChange={setDisplayMode}
              />
            )}
          </div>
          <div className="flex min-h-[calc(100vh-7rem)] items-center justify-center">
            <div className="rounded-3xl bg-white/80 px-6 py-4 text-sm text-slate-600 shadow-sm">
              正在检查登录状态...
            </div>
          </div>
        </div>
      );
    }

    return (
      <AuthPanel
        displayMode={displayMode}
        username={authUsername}
        password={authPassword}
        pending={authPending}
        onDisplayModeChange={setDisplayMode}
        onUsernameChange={setAuthUsername}
        onPasswordChange={setAuthPassword}
        onSubmit={handleAuthSubmit}
      />
    );
  }

  return (
    <LotteryShell
      displayMode={displayMode}
      activeTab={activeTab}
      navigationActiveTab={navigationActiveTab}
      navigationTabs={navigationTabs}
      tabs={tabs}
      currentUser={currentUser}
      onDisplayModeChange={setDisplayMode}
      onTabChange={setActiveTab}
    >
      {renderActivePanel(currentUser)}
    </LotteryShell>
  );
}
