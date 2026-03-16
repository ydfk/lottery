import { useMemo } from "react";
import { ClipboardCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { TicketRecognitionDraft } from "@/types/lottery";

interface TicketConfirmPanelProps {
  recognitionDraft: TicketRecognitionDraft | null;
  issue: string;
  notes: string;
  entryText: string;
  submitPending: boolean;
  onIssueChange: (value: string) => void;
  onNotesChange: (value: string) => void;
  onEntryTextChange: (value: string) => void;
  onCreateTicket: () => void;
}

function buildEntryText(recognitionDraft: TicketRecognitionDraft | null) {
  if (!recognitionDraft) {
    return "";
  }
  return recognitionDraft.entries
    .map((entry) => {
      const red = entry.red.map((value) => value.toString().padStart(2, "0")).join(",");
      const blue = entry.blue.map((value) => value.toString().padStart(2, "0")).join(",");
      return `${red}+${blue}`;
    })
    .join("\n");
}

export function TicketConfirmPanel(props: TicketConfirmPanelProps) {
  const {
    recognitionDraft,
    issue,
    notes,
    entryText,
    submitPending,
    onIssueChange,
    onNotesChange,
    onEntryTextChange,
    onCreateTicket,
  } = props;

  const placeholder = useMemo(() => buildEntryText(recognitionDraft), [recognitionDraft]);

  return (
    <Card className="border-white/60 bg-white/85 backdrop-blur">
      <CardHeader>
        <CardTitle className="text-slate-900">确认入库</CardTitle>
        <p className="text-sm text-slate-500">第三步确认期号、号码和备注，提交后正式入库并自动判奖。</p>
      </CardHeader>
      <CardContent className="space-y-5">
        {!recognitionDraft ? (
          <div className="rounded-[1.5rem] border border-dashed border-slate-300 bg-slate-50 p-8 text-center text-sm text-slate-500">
            请先完成“识别校对”，确认有可用号码后再入库。
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="space-y-4 rounded-[1.75rem] bg-slate-50 p-5">
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">期号</label>
                <Input
                  placeholder="可手动修正期号"
                  value={issue}
                  onChange={(event) => onIssueChange(event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">备注</label>
                <Textarea
                  placeholder="例如购买门店、机选/自选说明"
                  className="min-h-24 bg-white"
                  value={notes}
                  onChange={(event) => onNotesChange(event.target.value)}
                />
              </div>
              <div className="rounded-2xl border border-slate-200 bg-white p-4 text-sm leading-6 text-slate-600">
                提交后会把原图、识别文本和最终确认的号码一并入库；如果该期已开奖，会立即判奖。
              </div>
            </div>

            <div className="space-y-4">
              <div className="rounded-[1.75rem] border border-slate-200 bg-white p-5">
                <div className="flex items-center gap-2 text-slate-800">
                  <ClipboardCheck className="size-4" />
                  <p className="text-sm font-medium">确认号码</p>
                </div>
                <p className="mt-2 text-xs text-slate-500">
                  每行一注，格式：`01,02,03,04,05,06+07`
                </p>
                <Textarea
                  className="mt-3 min-h-56"
                  value={entryText}
                  placeholder={placeholder}
                  onChange={(event) => onEntryTextChange(event.target.value)}
                />
              </div>

              <Button className="h-12 w-full rounded-2xl" disabled={submitPending} onClick={onCreateTicket}>
                {submitPending ? "入库中..." : "确认入库并判奖"}
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
