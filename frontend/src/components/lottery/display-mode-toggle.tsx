import { MonitorCog, Smartphone } from "lucide-react";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { cn } from "@/lib/utils";

export type LotteryDisplayMode = "app" | "web";

interface LotteryDisplayModeToggleProps {
  value: LotteryDisplayMode;
  onValueChange: (value: LotteryDisplayMode) => void;
  className?: string;
  compact?: boolean;
}

const displayModeOptions = [
  {
    value: "app" as const,
    label: "应用",
    icon: Smartphone,
  },
  {
    value: "web" as const,
    label: "Web",
    icon: MonitorCog,
  },
];

export function LotteryDisplayModeToggle(props: LotteryDisplayModeToggleProps) {
  const { value, onValueChange, className, compact = false } = props;

  return (
    <ToggleGroup
      type="single"
      value={value}
      className={cn(
        "h-auto rounded-full border border-white/70 bg-white/92 p-1 shadow-[0_18px_40px_rgba(15,23,42,0.12)] backdrop-blur",
        className
      )}
      aria-label="切换展示模式"
      onValueChange={(nextValue) => {
        if (nextValue === "app" || nextValue === "web") {
          onValueChange(nextValue);
        }
      }}
    >
      {displayModeOptions.map((item) => {
        const Icon = item.icon;

        return (
          <ToggleGroupItem
            key={item.value}
            value={item.value}
            className={cn(
              "rounded-full text-slate-600 data-[state=on]:bg-slate-900 data-[state=on]:text-white",
              compact ? "size-9 px-0" : "px-3"
            )}
            aria-label={`切换到${item.label}展示`}
          >
            <Icon data-icon="inline-start" />
            <span className={cn(compact && "hidden sm:inline")}>{item.label}</span>
          </ToggleGroupItem>
        );
      })}
    </ToggleGroup>
  );
}
