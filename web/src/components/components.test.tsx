import { describe, expect, it, vi } from "vitest";
import React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import GlassButton from "./ui/GlassButton";
import { Input } from "./primitives/Input";
import { Label } from "./primitives/Label";
import { Skeleton } from "./primitives/Skeleton";

describe("GlassButton", () => {
  it("renders a button and fires onClick", async () => {
    const onClick = vi.fn();
    render(<GlassButton text="Place order" onClick={onClick} />);

    const button = screen.getByRole("button", { name: /place order/i });
    await userEvent.click(button);

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("renders a link when href is given", () => {
    render(<GlassButton text="Dashboard" href="/dashboard" />);

    const link = screen.getByRole("link", { name: /dashboard/i });
    expect(link).toHaveAttribute("href", "/dashboard");
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("does not fire onClick when disabled", async () => {
    const onClick = vi.fn();
    render(<GlassButton text="Nope" onClick={onClick} disabled />);

    const button = screen.getByRole("button", { name: /nope/i });
    expect(button).toBeDisabled();
    await userEvent.click(button);

    expect(onClick).not.toHaveBeenCalled();
  });

  it("renders an icon alongside the text", () => {
    render(<GlassButton text="With icon" icon={<svg data-testid="icon" />} />);
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });
});

describe("Input", () => {
  it("accepts typed text", async () => {
    render(<Input placeholder="Email" />);

    const input = screen.getByPlaceholderText("Email");
    await userEvent.type(input, "user@example.com");

    expect(input).toHaveValue("user@example.com");
  });

  it("respects type and disabled props", () => {
    render(<Input placeholder="Password" type="password" disabled />);

    const input = screen.getByPlaceholderText("Password");
    expect(input).toHaveAttribute("type", "password");
    expect(input).toBeDisabled();
  });
});

describe("Label", () => {
  it("associates with its input via htmlFor", () => {
    render(
      <>
        <Label htmlFor="email">Email address</Label>
        <input id="email" />
      </>
    );

    expect(screen.getByLabelText("Email address")).toBeInTheDocument();
  });
});

describe("Skeleton", () => {
  it("renders a pulsing placeholder with merged classes", () => {
    const { container } = render(<Skeleton className="h-4 w-32" />);
    const el = container.querySelector('[data-slot="skeleton"]');

    expect(el).not.toBeNull();
    expect(el!.className).toContain("animate-pulse");
    expect(el!.className).toContain("h-4");
  });
});
