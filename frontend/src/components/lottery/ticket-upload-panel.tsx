import { ImageUp, ScanLine } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { TicketUpload } from "@/types/lottery";

interface TicketUploadPanelProps {
  previewUrl: string;
  selectedImage: File | null;
  uploadPending: boolean;
  uploadedTicket: TicketUpload | null;
  onSelectImage: (file: File | null) => void;
  onUpload: () => void;
}

export function TicketUploadPanel(props: TicketUploadPanelProps) {
  const { previewUrl, selectedImage, uploadPending, uploadedTicket, onSelectImage, onUpload } = props;

  return (
    <Card className="border-white/60 bg-white/85 backdrop-blur">
      <CardHeader>
        <CardTitle className="text-slate-900">上传原图</CardTitle>
        <p className="text-sm text-slate-500">上传票据照片</p>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="grid gap-4 md:grid-cols-[1fr_0.9fr]">
          <label className="flex min-h-72 cursor-pointer flex-col items-center justify-center rounded-[1.75rem] border border-dashed border-slate-300 bg-slate-50 p-5 text-center">
            {previewUrl ? (
              <img src={previewUrl} alt="彩票预览" className="h-64 w-full rounded-2xl object-cover" />
            ) : (
              <>
                <ImageUp className="size-10 text-slate-400" />
                <p className="mt-4 text-base font-medium text-slate-700">点击拍照或上传彩票原图</p>
                <p className="mt-2 text-sm text-slate-500">手机端可直接调起后置摄像头</p>
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

          <div className="flex flex-col justify-between rounded-[1.75rem] bg-slate-50 p-5">
            <div className="space-y-4">
              {selectedImage && (
                <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700">
                  已选择：{selectedImage.name}
                </div>
              )}

              {uploadedTicket && (
                <div className="rounded-2xl border border-slate-200 bg-white p-4">
                  <p className="text-sm font-medium text-slate-700">已上传</p>
                  <p className="mt-2 break-all text-xs text-slate-500">uploadId: {uploadedTicket.id}</p>
                  <div className="mt-3 flex items-center gap-2 text-sm text-slate-600">
                    <ScanLine className="size-4" />
                    可继续识别
                  </div>
                </div>
              )}
            </div>

            <Button className="mt-5 h-12 rounded-2xl" disabled={!selectedImage || uploadPending} onClick={onUpload}>
              {uploadPending ? "上传中..." : "上传原图"}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
