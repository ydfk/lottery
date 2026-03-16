import { Slider } from "@/components/ui/slider";

type SliderProps = React.ComponentProps<typeof Slider>;

export default function SliderDemo({ className, ...props }: SliderProps) {
  return (
    <div className="w-[60%]">
      <Slider defaultValue={[50]} max={100} step={1} className={className} {...props} />
    </div>
  );
}
