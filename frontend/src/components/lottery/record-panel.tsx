import { ImageUp, Plus, RotateCcw, Save, ScanSearch, Trash2, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { NumberBalls } from "@/components/lottery/number-balls";
import { formatLotteryDrawDate, formatLotteryIssue, getLotteryDisplayName, lotteryDisplayOptions } from "@/lib/lottery-display";
import { buildParsedEntriesFromDrafts } from "@/lib/ticket-entry-drafts";
import type { Recommendation, TicketEntryDraft, TicketRecognitionDraft, TicketUpload } from "@/types/lottery";

interface RecordPanelProps {
  selectedRecommendation: Recommendation | null;
  previewUrl: string;
  selectedImage: File | null;
  uploadPending: boolean;
  uploadedTicket: TicketUpload | null;
  recognitionDraft: TicketRecognitionDraft | null;
  lotteryCode: string;
  recognizePending: boolean;
  issue: string;
  drawDate: string;
  costAmount: string;
  notes: string;
  entryDrafts: TicketEntryDraft[];
  submitPending: boolean;
  onSelectImage: (file: File | null) => void;
  onLotteryCodeChange: (value: string) => void;
  onRecognize: () => void;
  onIssueChange: (value: string) => void;
  onDrawDateChange: (value: string) => void;
  onCostAmountChange: (value: string) => void;
  onNotesChange: (value: string) => void;
  onEntryFieldChange: (index: number, field: "redNumbers" | "blueNumbers", value: string) => void;
  onToggleEntryAdditional: (index: number) => void;
  onChangeEntryMultiple: (index: number, nextMultiple: number) => void;
  onAddEntry: () => void;
  onRemoveEntry: (index: number) => void;
  onCreateTicket: () => void;
  onClearRecommendation: () => void;
}

export function RecordPanel(props: RecordPanelProps) {
  const {
    selectedRecommendation,
    previewUrl,
    selectedImage,
    uploadPending,
    uploadedTicket,
    recognitionDraft,
    lotteryCode,
    recognizePending,
    issue,
    drawDate,
    costAmount,
    notes,
    entryDrafts,
    submitPending,
    onSelectImage,
    onLotteryCodeChange,
    onRecognize,
    onIssueChange,
    onDrawDateChange,
    onCostAmountChange,
    onNotesChange,
    onEntryFieldChange,
    onToggleEntryAdditional,
    onChangeEntryMultiple,
    onAddEntry,
    onRemoveEntry,
    onCreateTicket,
    onClearRecommendation,
  } = props;

  const showAdditionalToggle = lotteryCode === "dlt";
  const recognizeLabel = recognitionDraft ? "重新识别" : "开始识别";
  const recognizeBusy = uploadPending || recognizePending;

  return (
    <div className="space-y-6">
      <section className="rounded-[1.6rem] border border-white/60 bg-white/88 p-4 shadow-[0_16px_40px_rgba(15,23,42,0.08)] backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Badge className="bg-slate-900 text-white hover:bg-slate-900">记录</Badge>
            <h2 className="text-base font-semibold text-slate-950">录入</h2>
          </div>
          <div className="flex items-center gap-2">
            {recognitionDraft ? <Badge variant="secondary">识别 {(recognitionDraft.confidence * 100).toFixed(0)}%</Badge> : null}
            {uploadedTicket ? <Badge variant="secondary">图片已上传</Badge> : null}
          </div>
        </div>
      </section>

      {selectedRecommendation && (
        <Card className="border-amber-200 bg-amber-50/90">
          <CardHeader className="flex flex-row items-start justify-between gap-4 pb-3">
            <div className="space-y-2">
              <CardTitle className="text-slate-900">关联推荐</CardTitle>
              <div className="flex items-center gap-2">
                <Badge variant="secondary">{getLotteryDisplayName(selectedRecommendation.lotteryCode)}</Badge>
                <Badge variant="secondary">第 {formatLotteryIssue(selectedRecommendation.lotteryCode, selectedRecommendation.issue)} 期</Badge>
                {selectedRecommendation.drawDate ? (
                  <Badge variant="secondary">{formatLotteryDrawDate(selectedRecommendation.drawDate)} 开奖</Badge>
                ) : null}
              </div>
            </div>
            <Button type="button" variant="ghost" size="icon" className="rounded-2xl" onClick={onClearRecommendation}>
              <X className="size-4" />
            </Button>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            {selectedRecommendation.entries.map((entry) => (
              <div key={entry.id} className="rounded-[1.25rem] bg-white p-4">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm font-medium text-slate-700">推荐 {entry.sequence}</span>
                  <span className="text-xs text-slate-500">{(entry.confidence * 100).toFixed(0)}%</span>
                </div>
                <div className="mt-3">
                  <NumberBalls redNumbers={entry.redNumbers} blueNumbers={entry.blueNumbers} compact />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <Card className="border-white/60 bg-white/85 shadow-[0_20px_50px_rgba(15,23,42,0.08)] backdrop-blur">
        <CardContent className="space-y-5">
          <div className="grid gap-3 lg:grid-cols-[0.88fr_1.12fr]">
            <label className="flex min-h-40 cursor-pointer flex-col items-center justify-center rounded-[1.45rem] border border-dashed border-slate-300 bg-slate-50 p-3 text-center">
              {previewUrl ? (
                <img src={previewUrl} alt="彩票预览" className="h-32 w-full rounded-[1.1rem] object-cover" />
              ) : (
                <>
                  <ImageUp className="size-7 text-slate-400" />
                  <p className="mt-1.5 text-sm font-medium text-slate-700">选择图片</p>
                </>
              )}
              <input className="hidden" type="file" accept="image/*" onChange={(event) => onSelectImage(event.target.files?.[0] || null)} />
            </label>

            <div className="space-y-3 rounded-[1.45rem] bg-slate-50 p-3">
              {selectedImage && (
                <div className="rounded-[1rem] border border-slate-200 bg-white px-3 py-2 text-sm text-slate-600">{selectedImage.name}</div>
              )}

              <div className="grid gap-3 sm:grid-cols-1">
                <Button
                  type="button"
                  className="h-10 rounded-[1rem]"
                  disabled={(!selectedImage && !uploadedTicket) || recognizeBusy}
                  onClick={onRecognize}
                >
                  {recognizeBusy ? (
                    uploadPending ? (
                      "上传中..."
                    ) : (
                      "识别中..."
                    )
                  ) : recognitionDraft ? (
                    <>
                      <RotateCcw className="mr-2 size-4" />
                      {recognizeLabel}
                    </>
                  ) : (
                    <>
                      <ScanSearch className="mr-2 size-4" />
                      {recognizeLabel}
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-2.5">
              <div className="min-w-0 space-y-1.5">
                <label className="text-sm font-medium text-slate-700">彩票类型</label>
                <select
                  className="flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm outline-none transition focus:border-slate-400"
                  value={lotteryCode}
                  onChange={(event) => onLotteryCodeChange(event.target.value)}
                >
                  <option value="">请选择</option>
                  {lotteryDisplayOptions.map((item) => (
                    <option key={item.code} value={item.code}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="min-w-0 space-y-1.5">
                <label className="text-sm font-medium text-slate-700">期号</label>
                <Input className="h-9" value={issue} onChange={(event) => onIssueChange(event.target.value)} />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-2.5">
              <div className="min-w-0 space-y-1.5">
                <label className="text-sm font-medium text-slate-700">开奖日期</label>
                <Input className="h-9 min-w-0" type="date" value={drawDate} onChange={(event) => onDrawDateChange(event.target.value)} />
              </div>

              <div className="min-w-0 space-y-1.5">
                <label className="text-sm font-medium text-slate-700">金额</label>
                <Input
                  className="h-9 min-w-0"
                  type="number"
                  step="0.01"
                  value={costAmount}
                  onChange={(event) => onCostAmountChange(event.target.value)}
                />
              </div>
            </div>
            <p className="text-[11px] leading-4 text-slate-400">金额会按号码、倍数和追加自动计算，也可手动修改。</p>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between gap-3">
              <label className="text-sm font-medium text-slate-700">号码</label>
              <Button type="button" variant="secondary" size="sm" className="rounded-full" onClick={onAddEntry}>
                <Plus className="size-3.5" />
                新增一注
              </Button>
            </div>

            <div className="space-y-3 rounded-[1.45rem] bg-slate-50 p-3">
              {entryDrafts.map((entry, index) => {
                const previewEntry = buildParsedEntriesFromDrafts([entry], lotteryCode)[0];

                return (
                  <div key={`entry-${index}`} className="rounded-[1.1rem] border border-slate-200 bg-white p-3">
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-sm font-medium text-slate-700">号码 {index + 1}</span>
                      <div className="flex items-center gap-2">
                        <div className="flex items-center rounded-full border border-slate-200 bg-slate-50">
                          <button
                            type="button"
                            className="h-7 w-7 text-sm text-slate-500 transition hover:text-slate-900"
                            onClick={() => onChangeEntryMultiple(index, Math.max(1, entry.multiple - 1))}
                          >
                            -
                          </button>
                          <span className="min-w-10 text-center text-xs text-slate-600">{entry.multiple} 倍</span>
                          <button
                            type="button"
                            className="h-7 w-7 text-sm text-slate-500 transition hover:text-slate-900"
                            onClick={() => onChangeEntryMultiple(index, entry.multiple + 1)}
                          >
                            +
                          </button>
                        </div>
                        {showAdditionalToggle && (
                          <Button
                            type="button"
                            variant={entry.isAdditional ? "default" : "secondary"}
                            size="sm"
                            className="h-7 rounded-full px-3 text-xs"
                            onClick={() => onToggleEntryAdditional(index)}
                          >
                            追加
                          </Button>
                        )}
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 rounded-full text-slate-500"
                          onClick={() => onRemoveEntry(index)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      </div>
                    </div>

                    <div className="mt-3 grid grid-cols-[minmax(0,1fr)_6rem] gap-2">
                      <div className="space-y-1.5">
                        <span className="text-xs text-slate-500">红球</span>
                        <Input
                          className="h-9 bg-white text-sm"
                          inputMode="numeric"
                          value={entry.redNumbers}
                          placeholder={lotteryCode === "dlt" ? "03 11 18 26 32" : "01 02 03 04 05 06"}
                          onChange={(event) => onEntryFieldChange(index, "redNumbers", event.target.value)}
                        />
                      </div>
                      <div className="space-y-1.5">
                        <span className="text-xs text-slate-500">蓝球</span>
                        <Input
                          className="h-9 bg-white text-sm"
                          inputMode="numeric"
                          value={entry.blueNumbers}
                          placeholder={lotteryCode === "dlt" ? "04 09" : "07"}
                          onChange={(event) => onEntryFieldChange(index, "blueNumbers", event.target.value)}
                        />
                      </div>
                    </div>

                    <div className="mt-2 text-[11px] leading-4 text-slate-400">支持空格、逗号或连续数字，例如 010203040506</div>

                    {previewEntry ? (
                      <div className="mt-3">
                        <NumberBalls
                          redNumbers={previewEntry.red.map((value) => value.toString().padStart(2, "0")).join(",")}
                          blueNumbers={previewEntry.blue.map((value) => value.toString().padStart(2, "0")).join(",")}
                          compact
                        />
                      </div>
                    ) : null}
                  </div>
                );
              })}
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">备注</label>
            <Textarea className="min-h-24 bg-white" value={notes} placeholder="备注" onChange={(event) => onNotesChange(event.target.value)} />
          </div>

          <Button
            type="button"
            variant="secondary"
            className="h-12 w-full rounded-2xl"
            disabled={submitPending || (!uploadedTicket && !selectedImage)}
            onClick={onCreateTicket}
          >
            <Save className="mr-2 size-4" />
            {submitPending ? "保存中..." : "保存票据"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
