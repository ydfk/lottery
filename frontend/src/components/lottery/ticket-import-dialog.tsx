import { useMemo, useState, type ChangeEvent } from "react";
import { FileArchive, FileSpreadsheet, UploadCloud } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { TicketImportResult } from "@/types/lottery";

interface TicketImportDialogProps {
  onImport: (workbook: File, imagesZip: File | null) => Promise<TicketImportResult>;
}

const workbookAccept = ".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet";
const imagesZipAccept = ".zip,application/zip,application/x-zip-compressed";

function formatFileSize(size: number) {
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function FilePickerCard(props: {
  title: string;
  description: string;
  accept: string;
  icon: typeof FileSpreadsheet;
  file: File | null;
  required?: boolean;
  onChange: (file: File | null) => void;
}) {
  const { title, description, accept, icon: Icon, file, required, onChange } = props;

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0] ?? null;
    onChange(nextFile);
    event.currentTarget.value = "";
  }

  return (
    <label className="group block cursor-pointer">
      <input type="file" accept={accept} className="sr-only" onChange={handleChange} />
      <div className="rounded-[1.4rem] border border-dashed border-slate-300 bg-slate-50/90 p-4 transition group-hover:border-slate-400 group-hover:bg-slate-50">
        <div className="flex items-start gap-3">
          <div className="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-white text-slate-600 shadow-sm">
            <Icon className="size-5" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm font-semibold text-slate-900">{title}</p>
              {required ? (
                <Badge className="bg-slate-900 text-white hover:bg-slate-900">必填</Badge>
              ) : (
                <Badge variant="secondary">可选</Badge>
              )}
            </div>
            <p className="mt-1 text-xs leading-5 text-slate-500">{description}</p>
            <div className="mt-3 rounded-2xl border border-white/80 bg-white px-3 py-2">
              {file ? (
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-slate-900">{file.name}</p>
                    <p className="text-xs text-slate-500">{formatFileSize(file.size)}</p>
                  </div>
                  <span className="text-xs font-medium text-slate-500">点击更换</span>
                </div>
              ) : (
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm text-slate-500">点击选择文件</p>
                  <span className="text-xs font-medium text-slate-500">上传</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </label>
  );
}

export function TicketImportDialog(props: TicketImportDialogProps) {
  const { onImport } = props;
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [workbookFile, setWorkbookFile] = useState<File | null>(null);
  const [imagesZipFile, setImagesZipFile] = useState<File | null>(null);
  const [result, setResult] = useState<TicketImportResult | null>(null);

  const failedRows = useMemo(
    () => result?.rows.filter((item) => item.status === "failed") ?? [],
    [result]
  );

  function resetState() {
    setWorkbookFile(null);
    setImagesZipFile(null);
    setResult(null);
  }

  function handleOpenChange(nextOpen: boolean) {
    if (pending && !nextOpen) {
      return;
    }

    setOpen(nextOpen);
    if (!nextOpen) {
      resetState();
    }
  }

  async function handleImport() {
    if (!workbookFile) {
      toast.error("请先选择 Excel 文件");
      return;
    }

    setPending(true);
    try {
      const importResult = await onImport(workbookFile, imagesZipFile);
      setResult(importResult);
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="h-10 rounded-full px-4"
        onClick={() => setOpen(true)}
      >
        <UploadCloud className="size-4" />
        导入 Excel
      </Button>

      <DialogContent className="max-w-3xl gap-0 rounded-[1.8rem] border-slate-200 bg-white p-0 shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
        <DialogHeader className="border-b border-slate-100 px-6 pb-4 pt-6">
          <DialogTitle className="text-xl text-slate-950">批量导入历史记录</DialogTitle>
          <DialogDescription className="mt-2 text-sm leading-6 text-slate-500">
            上传 Excel
            后会按一行一注导入，同彩种同一期号会自动合并为同一次购买记录；如果同时上传图片 ZIP，会按
            Excel 里的图片名自动匹配原图。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 px-6 py-5">
          <div className="grid gap-4 md:grid-cols-2">
            <FilePickerCard
              title="Excel 工作簿"
              description="支持 .xlsx，至少需要包含彩票类型、期号、开奖日期、购买时间、红球、蓝球、倍数、追加、金额等列。"
              accept={workbookAccept}
              icon={FileSpreadsheet}
              file={workbookFile}
              required
              onChange={(file) => {
                setWorkbookFile(file);
                setResult(null);
              }}
            />

            <FilePickerCard
              title="图片压缩包"
              description="可选上传 .zip；Excel 里的“图片名”列会按文件名匹配 ZIP 中图片，适合补齐票据原图。"
              accept={imagesZipAccept}
              icon={FileArchive}
              file={imagesZipFile}
              onChange={(file) => {
                setImagesZipFile(file);
                setResult(null);
              }}
            />
          </div>

          <div className="rounded-[1.4rem] border border-slate-200 bg-slate-50/90 p-4">
            <div className="flex flex-wrap gap-2">
              <Badge variant="secondary">一行一注号码</Badge>
              <Badge variant="secondary">同一期自动合并</Badge>
              <Badge variant="secondary">支持 双色球 / 大乐透</Badge>
              <Badge variant="secondary">可自动关联推荐</Badge>
            </div>
            <p className="mt-3 text-sm leading-6 text-slate-600">
              图片 ZIP
              不是必填。如果当前历史列表有筛选条件，导入成功后列表会自动刷新，但新记录可能因为筛选条件暂时不显示。
            </p>
          </div>

          {result ? (
            <div className="rounded-[1.4rem] border border-slate-200 bg-white">
              <div className="border-b border-slate-100 px-4 py-4">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge className="bg-slate-900 text-white hover:bg-slate-900">导入结果</Badge>
                  <Badge variant="secondary">共 {result.totalCount} 行</Badge>
                  <Badge className="bg-emerald-50 text-emerald-700 hover:bg-emerald-50">
                    成功 {result.successCount} 行
                  </Badge>
                  <Badge className="bg-rose-50 text-rose-700 hover:bg-rose-50">
                    失败 {result.failedCount} 行
                  </Badge>
                </div>
                <p className="mt-3 text-sm text-slate-600">
                  {result.failedCount > 0
                    ? "失败行已经列在下方，修正后可以直接重新选择文件再次导入。"
                    : "所有行都已导入完成，历史数据和看板统计已经自动刷新。"}
                </p>
              </div>

              {failedRows.length > 0 ? (
                <ScrollArea className="max-h-72">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-slate-50/80">
                        <TableHead className="pl-4">行号</TableHead>
                        <TableHead>彩种</TableHead>
                        <TableHead>期号</TableHead>
                        <TableHead className="pr-4">失败原因</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {failedRows.map((item) => (
                        <TableRow key={`${item.row}-${item.issue}-${item.message}`}>
                          <TableCell className="pl-4 font-medium text-slate-900">
                            第 {item.row} 行
                          </TableCell>
                          <TableCell className="text-slate-600">
                            {item.lotteryCode || "-"}
                          </TableCell>
                          <TableCell className="text-slate-600">{item.issue || "-"}</TableCell>
                          <TableCell className="pr-4 text-sm text-rose-600">
                            {item.message || "导入失败"}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </ScrollArea>
              ) : null}
            </div>
          ) : null}
        </div>

        <DialogFooter className="border-t border-slate-100 px-6 py-4 sm:justify-between sm:space-x-0">
          <p className="text-sm text-slate-500">
            {workbookFile ? `已选择 ${workbookFile.name}` : "请选择要导入的 Excel 文件"}
          </p>
          <div className="flex flex-col-reverse gap-2 sm:flex-row">
            <Button
              type="button"
              variant="ghost"
              className="rounded-full"
              disabled={pending}
              onClick={() => handleOpenChange(false)}
            >
              关闭
            </Button>
            <Button
              type="button"
              className="rounded-full"
              disabled={pending || !workbookFile}
              onClick={() => void handleImport()}
            >
              {pending ? "导入中..." : result ? "重新导入" : "开始导入"}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
