import { expect, test } from "@playwright/test";

test.describe("strategy backtesting", () => {
  test("runs a backtest from a preset and shows metrics + trades", async ({ page }) => {
    await page.goto("/strategies");

    await expect(page.getByText("RSI Dip Buyer").first()).toBeVisible({ timeout: 15_000 });
    await page.getByRole("button", { name: /^backtest$/i }).first().click();

    await expect(page.getByText("Backtest Strategy")).toBeVisible();

    await page.getByRole("button", { name: /run backtest/i }).click();

    const results = page.getByTestId("backtest-results");
    await expect(results).toBeVisible({ timeout: 15_000 });

    await expect(page.getByText("Total Return")).toBeVisible();
    await expect(page.getByText("Max Drawdown")).toBeVisible();
    await expect(page.getByText("Win Rate")).toBeVisible();

    await expect(
      results.getByText(/take profit|stop loss|signal|end of data/i).first()
    ).toBeVisible();
  });
});
