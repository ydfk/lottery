import { toast } from "sonner";
import { Button } from "@/components/ui/button";

export default function SonnerDemo() {
  return (
    <div className="flex flex-wrap gap-4">
      <Button
        variant="outline"
        onClick={() =>
          toast("Event has been created", {
            description: "Sunday, December 03, 2023 at 9:00 AM",
            action: {
              label: "Undo",
              onClick: () => console.log("Undo"),
            },
          })
        }
      >
        Show Toast
      </Button>

      <Button variant="outline" onClick={() => toast.success("Event has been created")}>
        Success
      </Button>

      <Button variant="outline" onClick={() => toast.error("Something went wrong")}>
        Error
      </Button>

      <Button variant="outline" onClick={() => toast.info("Event details updated")}>
        Info
      </Button>

      <Button variant="outline" onClick={() => toast.warning("Event date is approaching")}>
        Warning
      </Button>
    </div>
  );
}
