import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

export default function InputDemo() {
  return (
    <div className="grid w-full max-w-sm items-center gap-1.5 space-y-4">
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="email">Email</Label>
        <Input type="email" id="email" placeholder="Email" />
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="file">Picture</Label>
        <Input id="file" type="file" />
      </div>

      <div className="flex w-full max-w-sm items-center space-x-2">
        <Input type="email" placeholder="Email" />
        <Button type="submit">Subscribe</Button>
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="disabled">Disabled</Label>
        <Input disabled type="email" id="disabled" placeholder="Disabled Input" />
      </div>
    </div>
  );
}
