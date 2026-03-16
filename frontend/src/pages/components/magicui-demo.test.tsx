import { render, screen } from "@testing-library/react";
import MagicUiDemo from "./magicui-demo";

test("renders Magic UI demo heading", () => {
  render(<MagicUiDemo />);
  expect(screen.getByText("Magic UI")).toBeInTheDocument();
});
