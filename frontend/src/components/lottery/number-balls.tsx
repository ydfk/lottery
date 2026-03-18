import { cn } from "@/lib/utils";

interface NumberBallsProps {
  redNumbers: string;
  blueNumbers: string;
  compact?: boolean;
}

interface HitNumberBallsProps extends NumberBallsProps {
  drawRedNumbers?: string;
  drawBlueNumbers?: string;
}

function parseNumberList(value: string | undefined) {
  if (!value) {
    return [];
  }

  return value.split(",").map((item) => item.trim()).filter(Boolean);
}

function Ball({
  value,
  color,
  compact,
  active = true,
}: {
  value: string;
  color: "red" | "blue";
  compact?: boolean;
  active?: boolean;
}) {
  return (
    <span
      className={cn(
        "flex items-center justify-center rounded-full border font-semibold shadow-sm transition-colors",
        compact ? "size-8 text-sm" : "size-10 text-base",
        active
          ? color === "red"
            ? "border-[var(--lottery-red)] bg-[var(--lottery-red)] text-white"
            : "border-[var(--lottery-blue)] bg-[var(--lottery-blue)] text-white"
          : color === "red"
            ? "border-rose-200 bg-white text-rose-500"
            : "border-sky-200 bg-white text-sky-500"
      )}
    >
      {value}
    </span>
  );
}

export function NumberBalls({ redNumbers, blueNumbers, compact }: NumberBallsProps) {
  const redList = parseNumberList(redNumbers);
  const blueList = parseNumberList(blueNumbers);

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

export function HitNumberBalls(props: HitNumberBallsProps) {
  const { redNumbers, blueNumbers, drawRedNumbers, drawBlueNumbers, compact } = props;
  const redList = parseNumberList(redNumbers);
  const blueList = parseNumberList(blueNumbers);
  const drawRedSet = new Set(parseNumberList(drawRedNumbers));
  const drawBlueSet = new Set(parseNumberList(drawBlueNumbers));

  return (
    <div className="flex flex-wrap items-center gap-2">
      {redList.map((value) => (
        <Ball
          key={`hit-red-${value}`}
          value={value}
          color="red"
          compact={compact}
          active={drawRedSet.has(value)}
        />
      ))}
      {blueList.map((value) => (
        <Ball
          key={`hit-blue-${value}`}
          value={value}
          color="blue"
          compact={compact}
          active={drawBlueSet.has(value)}
        />
      ))}
    </div>
  );
}
