import { Loader2 } from "lucide-react";

export function LoadingSkeleton() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <Loader2 className="h-12 w-12 animate-spin text-primary mb-4" />
      <p className="text-muted-foreground">加载中...</p>
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="w-full animate-pulse space-y-3">
      <div className="h-10 bg-gray-200 rounded-t-md" />
      <div className="flex flex-wrap gap-1 justify-center">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="w-8 h-8 bg-gray-200 rounded-full" />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-2">
        <div className="h-6 bg-gray-200 rounded" />
        <div className="h-6 bg-gray-200 rounded" />
      </div>
      <div className="h-10 bg-gray-200 rounded-b-md" />
    </div>
  );
}

export function CardsLoadingSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <CardSkeleton key={i} />
      ))}
    </div>
  );
}