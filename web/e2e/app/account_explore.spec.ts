import { expect, test } from "@playwright/test";

test.describe("account", () => {
  test("shows the paper funds balance and reset control", async ({ page }) => {
    await page.goto("/account");

    await expect(page.getByText("Paper Funds")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("$100,000.00").first()).toBeVisible();
    await expect(page.getByRole("button", { name: "Reset" })).toBeVisible();
  });

  test("shows the user profile", async ({ page }) => {
    await page.goto("/account");

    await expect(page.getByText("mock@example.com")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Buying Power")).toBeVisible();
  });
});

test.describe("portfolio (real data)", () => {
  test("hero shows total value, allocation, and live-valued positions", async ({ page }) => {
    await page.goto("/portfolio");

    await expect(page.getByText("Total Account Value")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Allocation")).toBeVisible();

    for (const symbol of ["AAPL", "MSFT", "TSLA", "NVDA"]) {
      await expect(page.getByText(symbol).first()).toBeVisible();
    }
    await expect(page.getByText(/unrealized/i).first()).toBeVisible();
  });
});

test.describe("explore", () => {
  test("renders news and movers from the API", async ({ page }) => {
    await page.goto("/explore");

    await expect(page.getByText("Tech Giants Rally on New AI Breakthroughs")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("NVDA").first()).toBeVisible();
    await expect(page.getByText("RIVN").first()).toBeVisible();
  });
});

test.describe("strategies", () => {
  test("lists templates and persists a custom strategy", async ({ page }) => {
    await page.goto("/strategies");

    await expect(page.getByText("RSI Dip Buyer").first()).toBeVisible({ timeout: 15_000 });

    await page.goto("/strategies/new");
    await expect(page.getByPlaceholder("My RSI Reversal")).toBeVisible({ timeout: 15_000 });

    await page.getByPlaceholder("My RSI Reversal").fill("E2E Momentum");
    await page.getByRole("button", { name: /save strategy/i }).click();

    await expect(page).toHaveURL(/\/strategies$/, { timeout: 10_000 });
    await expect(page.getByText("E2E Momentum").first()).toBeVisible({ timeout: 10_000 });

    await page.reload();
    await expect(page.getByText("E2E Momentum").first()).toBeVisible({ timeout: 10_000 });
  });
});

test.describe("dashboard greeting", () => {
  test("greets the signed-in user by name", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page.getByText("mock_user").first()).toBeVisible({ timeout: 15_000 });
  });
});
