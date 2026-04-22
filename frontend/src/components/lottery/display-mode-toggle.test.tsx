import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { LotteryDisplayModeToggle } from "./display-mode-toggle";

test("renders display mode options", () => {
  render(<LotteryDisplayModeToggle value="app" onValueChange={() => undefined} />);

  expect(screen.getByRole("radio", { name: "切换到应用展示" })).toBeInTheDocument();
  expect(screen.getByRole("radio", { name: "切换到Web展示" })).toBeInTheDocument();
});

test("calls onValueChange when switching mode", async () => {
  const user = userEvent.setup();
  const onValueChange = vi.fn();

  render(<LotteryDisplayModeToggle value="app" onValueChange={onValueChange} />);

  await user.click(screen.getByRole("radio", { name: "切换到Web展示" }));

  expect(onValueChange).toHaveBeenCalledWith("web");
});
