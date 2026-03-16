import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Input } from "./Input";

describe("Input password visibility toggle", () => {
  it("renders a toggle button for password inputs that flips the input type", async () => {
    render(<Input placeholder="Password" type="password" />);

    const input = screen.getByPlaceholderText("Password");
    expect(input).toHaveAttribute("type", "password");

    const toggle = screen.getByRole("button", { name: /show password/i });
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    await userEvent.click(toggle);
    expect(input).toHaveAttribute("type", "text");
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: /hide password/i })).toBe(toggle);

    await userEvent.click(toggle);
    expect(input).toHaveAttribute("type", "password");
    expect(toggle).toHaveAttribute("aria-pressed", "false");
  });

  it("does not render a toggle button for non-password inputs", () => {
    render(<Input placeholder="Email" type="email" />);

    expect(screen.getByPlaceholderText("Email")).toHaveAttribute("type", "email");
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("keeps focus on the input when toggling (button is not a tab stop trap)", async () => {
    render(<Input placeholder="Password" type="password" />);

    const input = screen.getByPlaceholderText("Password");
    await userEvent.click(input);
    expect(input).toHaveFocus();

    await userEvent.click(screen.getByRole("button", { name: /show password/i }));
    expect(input).toHaveFocus();
  });
});
