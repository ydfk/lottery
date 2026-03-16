import { cn } from "@/lib/utils";

interface NumberBallsProps {
  redNumbers: string;
  blueNumbers: string;
  compact?: boolean;
}

function Ball({
  value,
  color,
  compact,
}: {
  value: string;
  color: "red" | "blue";
  compact?: boolean;
}) {
  return (
    <span
      className={cn(
        "flex items-center justify-center rounded-full font-semibold text-white shadow-sm",
        compact ? "size-8 text-sm" : "size-10 text-base",
        color === "red" ? "bg-[var(--lottery-red)]" : "bg-[var(--lottery-blue)]"
      )}
    >
      {value}
    </span>
  );
}

export function NumberBalls({ redNumbers, blueNumbers, compact }: NumberBallsProps) {
  const redList = redNumbers.split(",").filter(Boolean);
  const blueList = blueNumbers.split(",").filter(Boolean);

  return (
    <div className="flex flex-wrap items-center gap-2">
      {redList.map((value) => (
        <Ball key={`red-${value}`} value={value} color="red" compact={compact} />
      ))}
      {blueList.map((value) => (
        <Ball key={`blue-${value}`} value={value} color="blue" compact={compact} />
      ))}
    </div>
  );
}
