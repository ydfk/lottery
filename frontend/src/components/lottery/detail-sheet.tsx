import * as DialogPrimitive from "@radix-ui/react-dialog";
import { ChevronLeft } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface DetailSheetProps {
  open: boolean;
  title: string;
  rightAction?: ReactNode;
  children: ReactNode;
  onOpenChange: (open: boolean) => void;
}

export function DetailSheet(props: DetailSheetProps) {
  const { open, title, rightAction, children, onOpenChange } = props;

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-slate-950/35 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <DialogPrimitive.Content className="fixed inset-0 z-50 outline-none">
          <div className="mx-auto flex h-full max-w-4xl flex-col bg-[rgba(248,250,252,0.98)] shadow-[0_20px_60px_rgba(15,23,42,0.18)]">
            <header className="sticky top-0 z-10 grid grid-cols-[56px_1fr_auto] items-center gap-2 border-b border-slate-200/80 bg-white/92 px-3 py-2 backdrop-blur">
              <Button
                type="button"
                variant="ghost"
                className="h-11 w-11 rounded-full p-0 text-slate-700"
                onClick={() => onOpenChange(false)}
              >
                <ChevronLeft className="size-5" />
                <span className="sr-only">返回</span>
              </Button>
              <DialogPrimitive.Title className="truncate px-2 text-center text-base font-semibold text-slate-950">
                {title}
              </DialogPrimitive.Title>
              <div className={cn("flex min-w-0 items-center justify-end gap-2", !rightAction && "opacity-0")}>
                {rightAction || (
                  <span className="inline-flex h-11 min-w-11 items-center justify-center">占位</span>
                )}
              </div>
            </header>
            <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-6 sm:py-5">
              {children}
            </div>
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
