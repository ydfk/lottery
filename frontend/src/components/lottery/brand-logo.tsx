import { APP_BRAND_NAME, APP_BRAND_SUBTITLE } from "@/lib/brand";

interface BrandLogoProps {
  subtitle?: string;
}

export function BrandLogo({ subtitle = APP_BRAND_SUBTITLE }: BrandLogoProps) {
  return (
    <div className="flex items-center gap-3">
      <img alt={APP_BRAND_NAME} className="size-12 rounded-2xl shadow-[0_10px_30px_rgba(7,17,31,0.18)]" src="/lottery-logo.svg" />
      <div className="space-y-1">
        <p className="text-lg font-semibold tracking-[0.16em] text-slate-950">{APP_BRAND_NAME}</p>
        <p className="text-sm text-slate-500">{subtitle}</p>
      </div>
    </div>
  );
}
