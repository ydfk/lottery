import { useState } from "react";
import { KeyRound, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { deletePasskey, getPasskeys, registerPasskey } from "@/lib/api/methods/auth";
import type { PasskeyItem } from "@/types/auth";

export function PasskeyManagementDialog() {
  const [open, setOpen] = useState(false);
  const [passkeys, setPasskeys] = useState<PasskeyItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [deletingId, setDeletingId] = useState("");
  const [confirmingId, setConfirmingId] = useState("");

  async function refreshPasskeys() {
    setLoading(true);
    try {
      setPasskeys(await getPasskeys());
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取 Passkey 失败");
    } finally {
      setLoading(false);
    }
  }

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    setConfirmingId("");
    if (nextOpen) {
      void refreshPasskeys();
    }
  }

  async function handleRegister() {
    setRegistering(true);
    try {
      await registerPasskey();
      toast.success("Passkey 已添加");
      await refreshPasskeys();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "添加 Passkey 失败");
    } finally {
      setRegistering(false);
    }
  }

  async function handleDelete(id: string) {
    setDeletingId(id);
    try {
      await deletePasskey(id);
      setPasskeys((items) => items.filter((item) => item.id !== id));
      setConfirmingId("");
      toast.success("Passkey 已删除");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除 Passkey 失败");
    } finally {
      setDeletingId("");
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          className="h-10 shrink-0 rounded-2xl border border-white/20 bg-white/10 px-4 text-white hover:bg-white/15"
        >
          <KeyRound className="mr-2 size-4" />
          Passkey
        </Button>
      </DialogTrigger>
      <DialogContent className="rounded-[1.6rem] sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>管理 Passkey</DialogTitle>
          <DialogDescription>登录时无需输入用户名，浏览器会让你选择可用的账户。</DialogDescription>
        </DialogHeader>

        <div className="flex justify-end">
          <Button
            type="button"
            className="rounded-xl"
            disabled={registering || Boolean(deletingId)}
            onClick={() => void handleRegister()}
          >
            <Plus className="mr-2 size-4" />
            {registering ? "添加中..." : "添加 Passkey"}
          </Button>
        </div>

        <div className="space-y-2">
          {loading ? (
            <p className="rounded-2xl bg-slate-50 px-4 py-5 text-sm text-slate-500">
              正在读取 Passkey...
            </p>
          ) : passkeys.length === 0 ? (
            <p className="rounded-2xl bg-slate-50 px-4 py-5 text-sm text-slate-500">
              尚未添加 Passkey。
            </p>
          ) : (
            passkeys.map((passkey) => (
              <div
                key={passkey.id}
                className="flex items-center gap-3 rounded-2xl border border-slate-200 px-4 py-3"
              >
                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-blue-600">
                  <KeyRound className="size-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-semibold text-slate-900">Passkey</p>
                  <p className="mt-1 text-xs text-slate-500">
                    创建于 {new Date(passkey.createdAt).toLocaleString()}
                  </p>
                  <p className="mt-1 font-mono text-[11px] text-slate-400">
                    {passkey.id.slice(-8)}
                  </p>
                </div>
                {confirmingId === passkey.id ? (
                  <div className="flex shrink-0 gap-1">
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      disabled={deletingId === passkey.id}
                      onClick={() => setConfirmingId("")}
                    >
                      取消
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="destructive"
                      disabled={deletingId === passkey.id}
                      onClick={() => void handleDelete(passkey.id)}
                    >
                      {deletingId === passkey.id ? "删除中" : "确认删除"}
                    </Button>
                  </div>
                ) : (
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    className="shrink-0 text-slate-400 hover:text-red-600"
                    aria-label="删除 Passkey"
                    onClick={() => setConfirmingId(passkey.id)}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                )}
              </div>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
