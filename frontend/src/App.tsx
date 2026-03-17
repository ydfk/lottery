import { useEffect, useState, type FormEvent } from "react";
import { History, LayoutDashboard, ReceiptText, Sparkles } from "lucide-react";
import { toast } from "sonner";
import { AuthPanel, type AuthMode } from "@/components/lottery/auth-panel";
import { DashboardPanel } from "@/components/lottery/dashboard-panel";
import { HistoryPanel } from "@/components/lottery/history-panel";
import { RecommendationPanel } from "@/components/lottery/recommendation-panel";
import { RecordPanel } from "@/components/lottery/record-panel";
import { getStoredToken } from "@/lib/api/client";
import { getProfile, loginAndStoreToken, logout, register } from "@/lib/api/methods/auth";
import {
  createTicket,
  formatParsedEntry,
  getAllTickets,
  getDashboard,
  getRecommendationDetail,
  getRecommendations,
  recognizeTicket,
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

type TabKey = "dashboard" | "recommendation" | "records" | "history";

const tabs: { key: TabKey; label: string; icon: typeof LayoutDashboard }[] = [
  { key: "dashboard", label: "看板", icon: LayoutDashboard },
  { key: "recommendation", label: "推荐", icon: Sparkles },
  { key: "records", label: "记录", icon: ReceiptText },
  { key: "history", label: "历史", icon: History },
];

function buildParsedEntryText(entries: ParsedEntry[]) {
  return entries
    .map((entry) => {
      const { redNumbers, blueNumbers } = formatParsedEntry(entry);
      const multiple = entry.multiple > 1 ? ` (${entry.multiple})` : "";
      return `${redNumbers}+${blueNumbers}${multiple}`;
    })
    .join("\n");
}

function buildRecommendationEntryText(recommendation: Recommendation) {
  return recommendation.entries
    .map((entry) => `${entry.redNumbers}+${entry.blueNumbers}`)
    .join("\n");
}

function parseEntryText(value: string) {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const multipleMatch = line.match(/[（(]\s*(\d+)\s*[)）]\s*$/);
      const multiple = multipleMatch ? Number(multipleMatch[1]) : 1;
      const normalizedLine = line.replace(/[（(]\s*\d+\s*[)）]\s*$/, "").trim();
      const [redPart, bluePart] = normalizedLine.split("+");
      const redNumbers = redPart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      const blueNumbers = bluePart
        ?.split(",")
        .map((item) => Number(item.trim()))
        .filter((item) => Number.isFinite(item)) ?? [];
      return { red: redNumbers, blue: blueNumbers, multiple: multiple > 0 ? multiple : 1 };
    })
    .filter((entry) => entry.red.length > 0 && entry.blue.length > 0);
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
  const [authMode, setAuthMode] = useState<AuthMode>("login");
  const [currentUser, setCurrentUser] = useState<AuthUser | null>(null);
  const [authUsername, setAuthUsername] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [authPending, setAuthPending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [detailPending, setDetailPending] = useState(false);
  const [uploadPending, setUploadPending] = useState(false);
  const [recognizePending, setRecognizePending] = useState(false);
  const [submitPending, setSubmitPending] = useState(false);
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [tickets, setTickets] = useState<TicketRecord[]>([]);
  const [selectedRecommendation, setSelectedRecommendation] = useState<Recommendation | null>(null);
  const [purchaseRecommendation, setPurchaseRecommendation] = useState<Recommendation | null>(null);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [uploadedTicket, setUploadedTicket] = useState<TicketUpload | null>(null);
  const [recognitionDraft, setRecognitionDraft] = useState<TicketRecognitionDraft | null>(null);
  const [issue, setIssue] = useState("");
  const [ocrText, setOCRText] = useState("");
  const [notes, setNotes] = useState("");
  const [entryText, setEntryText] = useState("");

  function resetLotteryState() {
    setDashboard(null);
    setRecommendations([]);
    setTickets([]);
    setSelectedRecommendation(null);
    setPurchaseRecommendation(null);
    setActiveTab("dashboard");
  }

  function resetTicketWorkflow() {
    setSelectedImage(null);
    setPreviewUrl("");
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setIssue("");
    setOCRText("");
    setNotes("");
    setEntryText("");
  }

  function handleRequestError(error: unknown, fallbackMessage: string) {
    if (!getStoredToken()) {
      setCurrentUser(null);
      setAuthMode("login");
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
      const [dashboardData, recommendationData, ticketData] = await Promise.all([
        getDashboard(),
        getRecommendations(),
        getAllTickets(),
      ]);
      setDashboard(dashboardData);
      setRecommendations(recommendationData);
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

  async function handleOpenRecommendation(recommendationId: string | null) {
    if (!recommendationId) {
      setSelectedRecommendation(null);
      return;
    }

    const currentItem = recommendations.find((item) => item.id === recommendationId) || null;
    setSelectedRecommendation(currentItem);
    setDetailPending(true);

    try {
      const detail = await getRecommendationDetail(recommendationId);
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
    setIssue(recommendation.issue || "");
    setEntryText(buildRecommendationEntryText(recommendation));
    setActiveTab("records");
    toast.success("已切换到记录页，接下来上传这条推荐的购买票据即可");
  }

  function handleSelectImage(file: File | null) {
    setSelectedImage(file);
    setUploadedTicket(null);
    setRecognitionDraft(null);
    setOCRText("");
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
      setIssue(draft.issue || issue);
      setEntryText(buildParsedEntryText(draft.entries));
      toast.success("识别完成，请继续校对并入库");
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
      await createTicket({
        uploadId: uploadedTicket.id,
        recommendationId: purchaseRecommendation?.id,
        issue: issue || undefined,
        purchasedAt: new Date().toISOString(),
        notes: notes || undefined,
        entries: entries.map((entry) => ({
          ...formatParsedEntry(entry),
          multiple: entry.multiple,
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
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.84),rgba(241,245,249,0.9)_40%,rgba(226,232,240,1))] pb-28">
      <div className="mx-auto w-full max-w-6xl px-4 pb-10 pt-4 sm:px-6">
        {activeTab === "dashboard" && (
          <DashboardPanel currentUser={currentUser} dashboard={dashboard} onLogout={handleLogout} />
        )}

        {activeTab === "recommendation" && (
          <RecommendationPanel
            recommendations={recommendations}
            selectedRecommendation={selectedRecommendation}
            detailPending={detailPending}
            onSelectRecommendation={(recommendationId) => void handleOpenRecommendation(recommendationId)}
            onRecordPurchase={handleRecordPurchase}
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
            ocrText={ocrText}
            recognizePending={recognizePending}
            issue={issue}
            notes={notes}
            entryText={entryText}
            submitPending={submitPending}
            onSelectImage={handleSelectImage}
            onUpload={() => void handleUploadTicketImage()}
            onOCRTextChange={setOCRText}
            onRecognize={() => void handleRecognizeTicket()}
            onIssueChange={setIssue}
            onNotesChange={setNotes}
            onEntryTextChange={setEntryText}
            onCreateTicket={() => void handleCreateTicket()}
            onClearRecommendation={() => setPurchaseRecommendation(null)}
          />
        )}

        {activeTab === "history" && <HistoryPanel tickets={tickets} />}
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
