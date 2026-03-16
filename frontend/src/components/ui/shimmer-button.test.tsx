import { render, screen } from "@testing-library/react";
import { ShimmerButton } from "./shimmer-button";

test("renders button text", () => {
  render(<ShimmerButton>Get Started</ShimmerButton>);
  expect(screen.getByRole("button", { name: "Get Started" })).toBeInTheDocument();
});

test("respects disabled state", () => {
  render(<ShimmerButton disabled>Disabled</ShimmerButton>);
  expect(screen.getByRole("button", { name: "Disabled" })).toBeDisabled();
});
