import { ScanSearch } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { NumberBalls } from "@/components/lottery/number-balls";
import type { TicketRecognitionDraft, TicketUpload } from "@/types/lottery";

interface TicketRecognitionPanelProps {
  uploadedTicket: TicketUpload | null;
  recognitionDraft: TicketRecognitionDraft | null;
  ocrText: string;
  recognizePending: boolean;
  onOCRTextChange: (value: string) => void;
  onRecognize: () => void;
}

function formatNumbers(values: number[]) {
  return values.map((value) => value.toString().padStart(2, "0")).join(",");
}

export function TicketRecognitionPanel(props: TicketRecognitionPanelProps) {
  const { uploadedTicket, recognitionDraft, ocrText, recognizePending, onOCRTextChange, onRecognize } = props;

  return (
    <Card className="border-white/60 bg-white/85 backdrop-blur">
      <CardHeader>
        <CardTitle className="text-slate-900">识别校对</CardTitle>
        <p className="text-sm text-slate-500">第二步只做 OCR 识别和结果校对，不会入库。</p>
      </CardHeader>
      <CardContent className="space-y-5">
        {!uploadedTicket ? (
          <div className="rounded-[1.5rem] border border-dashed border-slate-300 bg-slate-50 p-8 text-center text-sm text-slate-500">
            请先在“上传原图”页上传一张彩票图片。
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="space-y-4">
              <div className="overflow-hidden rounded-[1.5rem] border border-slate-200 bg-slate-100">
                <img src={uploadedTicket.imageUrl} alt="已上传彩票原图" className="h-72 w-full object-cover" />
              </div>
              <div className="rounded-[1.5rem] bg-slate-50 p-4">
                <p className="text-sm font-medium text-slate-700">OCR 降级文本</p>
                <Textarea
                  className="mt-3 min-h-40 bg-white"
                  placeholder="可选。填写后会优先使用这段文本识别，便于调试解析逻辑。"
                  value={ocrText}
                  onChange={(event) => onOCRTextChange(event.target.value)}
                />
                <Button className="mt-4 h-11 w-full rounded-2xl" disabled={recognizePending} onClick={onRecognize}>
                  {recognizePending ? "识别中..." : "开始识别"}
                </Button>
              </div>
            </div>

            <div className="space-y-4">
              {recognitionDraft ? (
                <>
                  <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="text-sm font-medium text-slate-700">识别结果概览</p>
                        <p className="mt-1 text-xs text-slate-500">期号 {recognitionDraft.issue || "待补充"}</p>
                      </div>
                      <Badge variant="secondary">
                        置信度 {(recognitionDraft.confidence * 100).toFixed(0)}%
                      </Badge>
                    </div>
                    <p className="mt-3 rounded-2xl bg-white p-3 text-sm leading-6 text-slate-600">
                      {recognitionDraft.rawText || "暂无原始文本"}
                    </p>
                  </div>

                  <div className="space-y-3">
                    {recognitionDraft.entries.map((entry, index) => (
                      <div key={`${formatNumbers(entry.red)}-${formatNumbers(entry.blue)}-${index}`} className="rounded-[1.5rem] border border-slate-200 bg-white p-4">
                        <div className="flex items-center justify-between gap-3">
                          <span className="text-sm font-medium text-slate-700">识别注 {index + 1}</span>
                          <ScanSearch className="size-4 text-slate-400" />
                        </div>
                        <div className="mt-3">
                          <NumberBalls redNumbers={formatNumbers(entry.red)} blueNumbers={formatNumbers(entry.blue)} />
                        </div>
                      </div>
                    ))}
                  </div>
                </>
              ) : (
                <div className="rounded-[1.5rem] border border-dashed border-slate-300 bg-slate-50 p-8 text-center text-sm text-slate-500">
                  上传完成后，在这里查看 OCR 识别出的期号、原始文本和号码注单。
                </div>
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
