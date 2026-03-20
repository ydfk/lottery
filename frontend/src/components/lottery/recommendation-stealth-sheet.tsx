import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { Recommendation } from "@/types/lottery";

interface RecommendationStealthSheetProps {
  recommendation: Recommendation | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function parseNumbers(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

export function RecommendationStealthSheet(props: RecommendationStealthSheetProps) {
  const { recommendation, open, onOpenChange } = props;
  const [mounted, setMounted] = useState(false);
  const [exitCount, setExitCount] = useState(0);
  const resetTimerRef = useRef<number | null>(null);

  const rows = useMemo(() => {
    if (!recommendation) {
      return [];
    }
    return recommendation.entries.map((entry) => ({
      id: entry.id,
      redNumbers: parseNumbers(entry.redNumbers),
      blueNumbers: parseNumbers(entry.blueNumbers),
    }));
  }, [recommendation]);

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  useEffect(() => {
    if (!open) {
      setExitCount(0);
      if (resetTimerRef.current) {
        window.clearTimeout(resetTimerRef.current);
        resetTimerRef.current = null;
      }
    }
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [open]);

  if (!mounted || !open || !recommendation) {
    return null;
  }

  function handleExitTap() {
    const nextCount = exitCount + 1;
    setExitCount(nextCount);

    if (resetTimerRef.current) {
      window.clearTimeout(resetTimerRef.current);
    }

    if (nextCount >= 5) {
      setExitCount(0);
      onOpenChange(false);
      return;
    }

    resetTimerRef.current = window.setTimeout(() => {
      setExitCount(0);
      resetTimerRef.current = null;
    }, 1800);
  }

  return createPortal(
    <div className="fixed inset-0 z-[90] bg-white">
      <div className="relative flex h-full flex-col overflow-hidden px-3 pb-3 pt-3 sm:px-6">
        <div className="flex justify-center pb-3">
          <button
            type="button"
            aria-label="退出隐览"
            className="flex h-8 items-center justify-center px-4"
            onClick={handleExitTap}
          >
            <div className="flex items-center gap-1.5">
              {Array.from({ length: 5 }).map((_, index) => (
                <span
                  key={index}
                  className={`block size-1.5 rounded-full transition-colors ${
                    index < exitCount ? "bg-slate-500" : "bg-slate-200"
                  }`}
                />
              ))}
            </div>
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-auto">
          <div className="mx-auto w-full max-w-5xl">
            <div className="overflow-hidden border border-slate-300">
              <div className="divide-y divide-slate-300">
                {rows.map((row) => (
                  <div
                    key={row.id}
                    className="grid min-h-18 grid-cols-[repeat(7,minmax(0,1fr))] gap-px bg-slate-300 p-px md:min-h-20"
                  >
                    {row.redNumbers.map((number, index) => (
                      <div
                        key={`${row.id}-red-${index}`}
                        className="flex items-center justify-center bg-white text-base font-medium tracking-[0.16em] text-rose-700 md:text-xl"
                      >
                        {number}
                      </div>
                    ))}
                    {row.blueNumbers.map((number, index) => (
                      <div
                        key={`${row.id}-blue-${index}`}
                        className="flex items-center justify-center border-l border-slate-300 bg-sky-50 text-base font-medium tracking-[0.16em] text-sky-700 md:text-xl"
                      >
                        {number}
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
