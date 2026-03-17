import { expect, test } from "@playwright/test";

test.describe("watchlist", () => {
  test("shows the seeded watchlist and its symbols", async ({ page }) => {
    await page.goto("/watchlist");

    await expect(page.getByText("My Watchlist").first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(/assets tracked/i)).toBeVisible();

    for (const symbol of ["AAPL", "MSFT", "NVDA"]) {
      await expect(page.getByText(symbol).first()).toBeVisible();
    }
  });

  test("creates a new watchlist", async ({ page }) => {
    await page.goto("/watchlist");
    await expect(page.getByText("My Watchlist").first()).toBeVisible({ timeout: 15_000 });

    await page.locator("button:has(svg.lucide-plus)").first().click();
    const input = page.getByPlaceholder("watchlist name");
    await expect(input).toBeVisible();
    await input.fill("E2E List");
    await input.press("Enter");

    await expect(page.getByText("E2E List").first()).toBeVisible({ timeout: 10_000 });
  });
});

test.describe("portfolio", () => {
  test("shows seeded positions", async ({ page }) => {
    await page.goto("/portfolio");

    for (const symbol of ["AAPL", "MSFT", "TSLA", "NVDA"]) {
      await expect(page.getByText(symbol).first()).toBeVisible({ timeout: 15_000 });
    }
  });
});
