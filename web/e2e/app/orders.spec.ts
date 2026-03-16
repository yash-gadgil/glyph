import { expect, Page, test } from "@playwright/test";

async function openOrders(page: Page) {
  await page.goto("/orders");
  await expect(page.getByRole("heading", { name: "Order Management" })).toBeVisible({
    timeout: 15_000,
  });
}

function orderForm(page: Page) {
  return {
    symbol: page.getByPlaceholder("AAPL"),
    quantity: page.getByPlaceholder("0"),
    price: page.getByPlaceholder(/15025/),
    stopPrice: page.getByPlaceholder(/14900/),
    submitBuy: page.getByRole("button", { name: /place buy order/i }),
    submitSell: page.getByRole("button", { name: /place sell order/i }),
  };
}

test.describe("order management", () => {
  test("places a market buy order and sees it fill", async ({ page }) => {
    await openOrders(page);
    const form = orderForm(page);

    await form.symbol.fill("NVDA");
    await form.quantity.fill("3");
    await form.submitBuy.click();

    await expect(page.getByText("NVDA").first()).toBeVisible();
    await expect(page.getByText(/pending/i).first()).toBeVisible();

    await expect(page.getByText(/filled/i).first()).toBeVisible({ timeout: 10_000 });
  });

  test("validates the form before submitting", async ({ page }) => {
    await openOrders(page);
    const form = orderForm(page);

    await form.submitBuy.click();
    await expect(page.getByText("Symbol is required")).toBeVisible();

    await form.symbol.fill("AAPL");
    await form.submitBuy.click();
    await expect(page.getByText(/quantity must be a positive integer/i)).toBeVisible();
  });

  test("requires a price for limit orders", async ({ page }) => {
    await openOrders(page);
    const form = orderForm(page);

    await form.symbol.fill("AAPL");
    await form.quantity.fill("2");
    await page.locator("select").first().selectOption("limit");

    await form.submitBuy.click();
    await expect(page.getByText(/price is required/i)).toBeVisible();

    await form.price.fill("18000");
    await form.submitBuy.click();

    await expect(page.getByText(/price is required/i)).not.toBeVisible();
    await expect(page.getByText("AAPL").first()).toBeVisible();
  });

  test("switches to sell mode", async ({ page }) => {
    await openOrders(page);

    await page.getByRole("button", { name: "Sell", exact: true }).click();
    await expect(orderForm(page).submitSell).toBeVisible();
  });

  test("cancels an order from the list", async ({ page }) => {
    await openOrders(page);
    const form = orderForm(page);

    await form.symbol.fill("TSLA");
    await form.quantity.fill("1");
    await form.submitBuy.click();
    await expect(page.getByText("TSLA").first()).toBeVisible();

    page.on("dialog", (dialog) => dialog.accept());

    const row = page.locator("tr", { hasText: "TSLA" }).first();
    await row.hover();
    await row.getByTitle("Cancel Order").click();

    await expect(page.getByText(/cancelled/i).first()).toBeVisible({ timeout: 10_000 });
  });

  test("filters orders with the status tabs", async ({ page }) => {
    await openOrders(page);

    await page.getByRole("button", { name: "Cancelled", exact: true }).click();
    await expect(page.getByText(/pending/i)).toHaveCount(0);
  });
});
