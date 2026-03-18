import { ImageUp, RotateCcw, Save, ScanSearch, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { NumberBalls } from "@/components/lottery/number-balls";
import { formatLotteryIssue, getLotteryDisplayName, lotteryDisplayOptions } from "@/lib/lottery-display";
import type { Recommendation, TicketRecognitionDraft, TicketUpload } from "@/types/lottery";

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
  entryText: string;
  submitPending: boolean;
  onSelectImage: (file: File | null) => void;
  onLotteryCodeChange: (value: string) => void;
  onRecognize: () => void;
  onIssueChange: (value: string) => void;
  onDrawDateChange: (value: string) => void;
  onCostAmountChange: (value: string) => void;
  onNotesChange: (value: string) => void;
  onEntryTextChange: (value: string) => void;
  onToggleEntryAdditional: (index: number) => void;
  onCreateTicket: () => void;
  onClearRecommendation: () => void;
}

function buildPreviewEntries(value: string) {
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
      return {
        redNumbers:
          redPart
            ?.split(",")
            .map((item) => item.trim())
            .filter(Boolean)
            .join(",") || "",
        blueNumbers:
          bluePart
            ?.split(",")
            .map((item) => item.trim())
            .filter(Boolean)
            .join(",") || "",
        multiple: multiple > 0 ? multiple : 1,
        isAdditional,
      };
    })
    .filter((entry) => entry.redNumbers && entry.blueNumbers);
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
    entryText,
    submitPending,
    onSelectImage,
    onLotteryCodeChange,
    onRecognize,
    onIssueChange,
    onDrawDateChange,
    onCostAmountChange,
    onNotesChange,
    onEntryTextChange,
    onToggleEntryAdditional,
    onCreateTicket,
    onClearRecommendation,
  } = props;

  const previewEntries = buildPreviewEntries(entryText);
  const showAdditionalToggle = lotteryCode === "dlt";
  const recognizeLabel = recognitionDraft ? "重新识别" : "开始识别";
  const recognizeBusy = uploadPending || recognizePending;

  return (
    <div className="space-y-6">
      {selectedRecommendation && (
        <Card className="border-amber-200 bg-amber-50/90">
          <CardHeader className="flex flex-row items-start justify-between gap-4 pb-3">
            <div className="space-y-2">
              <CardTitle className="text-slate-900">关联推荐</CardTitle>
              <div className="flex items-center gap-2">
                <Badge variant="secondary">{getLotteryDisplayName(selectedRecommendation.lotteryCode)}</Badge>
                <Badge variant="secondary">
                  第 {formatLotteryIssue(selectedRecommendation.lotteryCode, selectedRecommendation.issue)} 期
                </Badge>
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
        <CardHeader className="pb-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <CardTitle className="text-slate-900">录入票据</CardTitle>
              <p className="mt-1 text-sm text-slate-500">选图后识别，确认无误再保存</p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {recognitionDraft && (
                <Badge variant="secondary">
                  识别 {(recognitionDraft.confidence * 100).toFixed(0)}%
                </Badge>
              )}
              {uploadedTicket && <Badge variant="secondary">图片已上传</Badge>}
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
            <label className="flex min-h-80 cursor-pointer flex-col items-center justify-center rounded-[1.75rem] border border-dashed border-slate-300 bg-slate-50 p-5 text-center">
              {previewUrl ? (
                <img src={previewUrl} alt="彩票预览" className="h-72 w-full rounded-2xl object-cover" />
              ) : (
                <>
                  <ImageUp className="size-10 text-slate-400" />
                  <p className="mt-4 text-base font-medium text-slate-700">点击拍照或上传彩票原图</p>
                </>
              )}
              <input
                className="hidden"
                type="file"
                accept="image/*"
                capture="environment"
                onChange={(event) => onSelectImage(event.target.files?.[0] || null)}
              />
            </label>

            <div className="space-y-4 rounded-[1.75rem] bg-slate-50 p-5">
              {selectedImage && (
                <div className="rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600">
                  {selectedImage.name}
                </div>
              )}

              <div className="grid gap-3 sm:grid-cols-2">
                <Button
                  type="button"
                  className="h-12 rounded-2xl"
                  disabled={(!selectedImage && !uploadedTicket) || recognizeBusy}
                  onClick={onRecognize}
                >
                  {recognizeBusy ? (
                    uploadPending ? "上传中..." : "识别中..."
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

          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">彩票类型</label>
              <select
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm outline-none transition focus:border-slate-400"
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

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">期号</label>
              <Input value={issue} onChange={(event) => onIssueChange(event.target.value)} />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">开奖日期</label>
              <Input type="date" value={drawDate} onChange={(event) => onDrawDateChange(event.target.value)} />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">金额</label>
              <Input
                type="number"
                step="0.01"
                value={costAmount}
                onChange={(event) => onCostAmountChange(event.target.value)}
              />
            </div>
          </div>

          <div className="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-700">号码</label>
              <Textarea
                className="min-h-56 bg-white"
                value={entryText}
                placeholder="每行一注，例如 01,02,03,04,05,06+07 (2)"
                onChange={(event) => onEntryTextChange(event.target.value)}
              />
            </div>

            <div className="space-y-4 rounded-[1.75rem] bg-slate-50 p-5">
              <div className="space-y-3">
                {previewEntries.length > 0 ? (
                  previewEntries.map((entry, index) => (
                    <div key={`${entry.redNumbers}-${entry.blueNumbers}-${index}`} className="rounded-[1.25rem] border border-slate-200 bg-white p-4">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-sm font-medium text-slate-700">号码 {index + 1}</span>
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-slate-500">{entry.multiple} 倍</span>
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
                        </div>
                      </div>
                      <div className="mt-3">
                        <NumberBalls redNumbers={entry.redNumbers} blueNumbers={entry.blueNumbers} compact />
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="rounded-[1.25rem] border border-dashed border-slate-300 bg-white px-4 py-6 text-center text-sm text-slate-500">
                    识别后在这里预览号码
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-700">备注</label>
            <Textarea
              className="min-h-24 bg-white"
              value={notes}
              placeholder="备注"
              onChange={(event) => onNotesChange(event.target.value)}
            />
          </div>

          <Button
            type="button"
            variant="secondary"
            className="h-12 w-full rounded-2xl"
            disabled={submitPending || !uploadedTicket || !recognitionDraft}
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
