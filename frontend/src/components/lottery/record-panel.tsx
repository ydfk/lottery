import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { NumberBalls } from "@/components/lottery/number-balls";
import { TicketConfirmPanel } from "@/components/lottery/ticket-confirm-panel";
import { TicketRecognitionPanel } from "@/components/lottery/ticket-recognition-panel";
import { TicketUploadPanel } from "@/components/lottery/ticket-upload-panel";
import { getLotteryDisplayName } from "@/lib/lottery-display";
import type { Recommendation, TicketRecognitionDraft, TicketUpload } from "@/types/lottery";

interface RecordPanelProps {
  selectedRecommendation: Recommendation | null;
  previewUrl: string;
  selectedImage: File | null;
  uploadPending: boolean;
  uploadedTicket: TicketUpload | null;
  recognitionDraft: TicketRecognitionDraft | null;
  ocrText: string;
  recognizePending: boolean;
  issue: string;
  notes: string;
  entryText: string;
  submitPending: boolean;
  onSelectImage: (file: File | null) => void;
  onUpload: () => void;
  onOCRTextChange: (value: string) => void;
  onRecognize: () => void;
  onIssueChange: (value: string) => void;
  onNotesChange: (value: string) => void;
  onEntryTextChange: (value: string) => void;
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
    ocrText,
    recognizePending,
    issue,
    notes,
    entryText,
    submitPending,
    onSelectImage,
    onUpload,
    onOCRTextChange,
    onRecognize,
    onIssueChange,
    onNotesChange,
    onEntryTextChange,
    onCreateTicket,
    onClearRecommendation,
  } = props;

  return (
    <div className="space-y-6">
      {selectedRecommendation && (
        <Card className="border-amber-200 bg-amber-50/90">
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <CardTitle className="text-slate-900">关联推荐</CardTitle>
            </div>
            <Button type="button" variant="ghost" size="icon" className="rounded-2xl" onClick={onClearRecommendation}>
              <X className="size-4" />
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <Badge variant="secondary">{getLotteryDisplayName(selectedRecommendation.lotteryCode)}</Badge>
              <Badge variant="secondary">第 {selectedRecommendation.issue} 期</Badge>
            </div>
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

      <TicketUploadPanel
        previewUrl={previewUrl}
        selectedImage={selectedImage}
        uploadPending={uploadPending}
        uploadedTicket={uploadedTicket}
        onSelectImage={onSelectImage}
        onUpload={onUpload}
      />

      <TicketRecognitionPanel
        uploadedTicket={uploadedTicket}
        recognitionDraft={recognitionDraft}
        ocrText={ocrText}
        recognizePending={recognizePending}
        onOCRTextChange={onOCRTextChange}
        onRecognize={onRecognize}
      />

      <TicketConfirmPanel
        recognitionDraft={recognitionDraft}
        issue={issue}
        notes={notes}
        entryText={entryText}
        submitPending={submitPending}
        onIssueChange={onIssueChange}
        onNotesChange={onNotesChange}
        onEntryTextChange={onEntryTextChange}
        onCreateTicket={onCreateTicket}
      />
    </div>
  );
}
