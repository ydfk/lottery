import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { BrandLogo } from "@/components/lottery/brand-logo";
import {
  LotteryDisplayModeToggle,
  type LotteryDisplayMode,
} from "@/components/lottery/display-mode-toggle";
import { APP_VERSION_LABEL } from "@/lib/app-meta";
import { cn } from "@/lib/utils";
import type { AuthUser } from "@/types/auth";

export interface LotteryShellTab<Key extends string = string> {
  key: Key;
  label: string;
  description: string;
  icon: LucideIcon;
}

interface LotteryShellProps<Key extends string = string> {
  displayMode: LotteryDisplayMode;
  activeTab: Key;
  tabs: readonly LotteryShellTab<Key>[];
  navigationTabs?: readonly LotteryShellTab<Key>[];
  navigationActiveTab?: Key;
  currentUser: AuthUser;
  children: ReactNode;
  onTabChange: (tab: Key) => void;
  onDisplayModeChange: (mode: LotteryDisplayMode) => void;
}

export function LotteryShell<Key extends string = string>(props: LotteryShellProps<Key>) {
  const {
    displayMode,
    activeTab,
    tabs,
    navigationTabs,
    navigationActiveTab,
    currentUser,
    children,
    onTabChange,
    onDisplayModeChange,
  } = props;

  if (displayMode === "web") {
    return (
      <WebLotteryShell
        activeTab={activeTab}
        currentUser={currentUser}
        navigationActiveTab={navigationActiveTab}
        navigationTabs={navigationTabs ?? tabs}
        tabs={tabs}
        onDisplayModeChange={onDisplayModeChange}
        onTabChange={onTabChange}
      >
        {children}
      </WebLotteryShell>
    );
  }

  return (
    <AppLotteryShell
      activeTab={activeTab}
      navigationActiveTab={navigationActiveTab}
      navigationTabs={navigationTabs ?? tabs}
      tabs={tabs}
      onTabChange={onTabChange}
    >
      {children}
    </AppLotteryShell>
  );
}

interface ShellChildrenProps {
  children: ReactNode;
}

interface AppLotteryShellProps<Key extends string = string> extends ShellChildrenProps {
  activeTab: Key;
  navigationTabs: readonly LotteryShellTab<Key>[];
  navigationActiveTab?: Key;
  tabs: readonly LotteryShellTab<Key>[];
  onTabChange: (tab: Key) => void;
}

function AppLotteryShell<Key extends string = string>(props: AppLotteryShellProps<Key>) {
  const { activeTab, navigationTabs, navigationActiveTab, children, onTabChange } = props;
  const currentNavigationTab = navigationActiveTab ?? activeTab;

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.84),rgba(241,245,249,0.9)_40%,rgba(226,232,240,1))] pb-28">
      <div className="mx-auto w-full max-w-6xl px-4 pb-10 pt-4 sm:px-6">{children}</div>

      <nav className="fixed inset-x-0 bottom-4 z-50 mx-auto flex w-[calc(100%-1.5rem)] max-w-5xl items-center gap-2 rounded-[2rem] border border-white/70 bg-white/90 p-2 shadow-[0_18px_40px_rgba(15,23,42,0.14)] backdrop-blur">
        {navigationTabs.map((item) => {
          const Icon = item.icon;
          const active = currentNavigationTab === item.key;

          return (
            <button
              key={item.key}
              type="button"
              className={cn(
                "flex flex-1 items-center justify-center gap-2 rounded-[1.25rem] px-3 py-3 text-sm transition",
                active
                  ? "bg-slate-900 text-white shadow-sm"
                  : "text-slate-500 hover:bg-slate-100 hover:text-slate-900"
              )}
              onClick={() => onTabChange(item.key)}
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

interface WebLotteryShellProps<Key extends string = string> extends ShellChildrenProps {
  activeTab: Key;
  currentUser: AuthUser;
  navigationTabs: readonly LotteryShellTab<Key>[];
  navigationActiveTab?: Key;
  tabs: readonly LotteryShellTab<Key>[];
  onTabChange: (tab: Key) => void;
  onDisplayModeChange: (mode: LotteryDisplayMode) => void;
}

function WebLotteryShell<Key extends string = string>(props: WebLotteryShellProps<Key>) {
  const {
    activeTab,
    currentUser,
    navigationTabs,
    navigationActiveTab,
    tabs,
    children,
    onTabChange,
    onDisplayModeChange,
  } = props;
  const activeTabMeta = tabs.find((item) => item.key === activeTab) ?? tabs[0];
  const currentNavigationTab = navigationActiveTab ?? activeTab;

  return (
    <div className="min-h-screen bg-[linear-gradient(180deg,rgba(248,250,252,0.96),rgba(226,232,240,0.88)),radial-gradient(circle_at_top_right,rgba(59,130,246,0.14),transparent_40%)] px-4 py-4 sm:px-6 lg:px-8">
      <div className="mx-auto flex w-full max-w-[1500px] flex-col gap-6">
        <header className="relative overflow-hidden rounded-[2rem] border border-white/70 bg-[linear-gradient(135deg,rgba(255,255,255,0.94),rgba(241,245,249,0.92)_52%,rgba(226,232,240,0.88))] p-5 shadow-[0_22px_60px_rgba(15,23,42,0.14)]">
          <div className="absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_top,rgba(37,99,235,0.16),transparent_62%)]" />
          <div className="relative flex flex-col gap-5 xl:flex-row xl:items-center xl:justify-between">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:gap-5">
              <BrandLogo subtitle="Web 工作台视图" />
              <div className="flex flex-wrap items-center gap-2">
                <Badge className="bg-slate-900 text-white hover:bg-slate-900">桌面布局</Badge>
                <Badge variant="secondary">{activeTabMeta.label}</Badge>
                <span className="text-sm text-slate-500">当前用户 {currentUser.username}</span>
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <LotteryDisplayModeToggle value="web" onValueChange={onDisplayModeChange} />
              <div className="rounded-[1.4rem] border border-white/70 bg-white/72 px-4 py-3 shadow-sm backdrop-blur">
                <p className="text-xs font-medium tracking-[0.18em] text-slate-400 uppercase">
                  当前视图
                </p>
                <p className="mt-1 text-sm font-semibold text-slate-900">
                  {activeTabMeta.label} · {activeTabMeta.description}
                </p>
              </div>
            </div>
          </div>
        </header>

        <div className="grid gap-6 xl:grid-cols-[280px_minmax(0,1fr)]">
          <aside className="xl:sticky xl:top-4 xl:self-start">
            <Card className="border-white/70 bg-white/82 shadow-[0_20px_55px_rgba(15,23,42,0.08)] backdrop-blur">
              <CardHeader className="gap-3">
                <Badge variant="secondary" className="w-fit">
                  Web 导航
                </Badge>
                <CardTitle>工作台</CardTitle>
                <CardDescription>更适合桌面端浏览推荐、录入票据和查看历史记录。</CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                {navigationTabs.map((item) => {
                  const Icon = item.icon;
                  const active = currentNavigationTab === item.key;

                  return (
                    <button
                      key={item.key}
                      type="button"
                      className={cn(
                        "flex items-start gap-3 rounded-[1.4rem] border px-4 py-4 text-left transition",
                        active
                          ? "border-slate-900 bg-slate-900 text-white shadow-[0_18px_36px_rgba(15,23,42,0.18)]"
                          : "border-transparent bg-slate-50 text-slate-700 hover:border-slate-200 hover:bg-white"
                      )}
                      onClick={() => onTabChange(item.key)}
                    >
                      <div
                        className={cn(
                          "flex size-10 shrink-0 items-center justify-center rounded-2xl",
                          active ? "bg-white/10" : "bg-slate-900 text-white"
                        )}
                      >
                        <Icon className="size-4" />
                      </div>
                      <div className="min-w-0">
                        <p className="text-sm font-semibold">{item.label}</p>
                        <p
                          className={cn(
                            "mt-1 text-xs leading-5",
                            active ? "text-white/70" : "text-slate-500"
                          )}
                        >
                          {item.description}
                        </p>
                      </div>
                    </button>
                  );
                })}

                <Separator />

                <div className="rounded-[1.4rem] bg-slate-900 px-4 py-4 text-white">
                  <p className="text-xs font-medium tracking-[0.16em] text-white/55 uppercase">
                    版本
                  </p>
                  <p className="mt-2 text-sm font-semibold">{APP_VERSION_LABEL}</p>
                </div>
              </CardContent>
            </Card>
          </aside>

          <main className="min-w-0">{children}</main>
        </div>
      </div>
    </div>
  );
}
