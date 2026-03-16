import { ShimmerButton } from "@/components/magicui/shimmer-button";
import { RainbowButton } from "@/components/ui/rainbow-button";

export default function MagicUiDemo() {
  return (
    <div className="flex justify-center">
      <div className="w-full max-w-lg rounded-md border p-6">
        <h1 className="text-lg font-semibold">Magic UI</h1>
        <div className="mt-4 flex flex-wrap gap-3">
          <ShimmerButton>Get Started</ShimmerButton>
          <ShimmerButton className="text-xs px-3 py-1">Small</ShimmerButton>
          <RainbowButton>Get Started</RainbowButton>
        </div>
      </div>
    </div>
  );
}
