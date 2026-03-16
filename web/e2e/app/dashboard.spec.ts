import { expect, test } from "@playwright/test";

test.describe("session routing in mock mode", () => {
  test("the landing page redirects to the dashboard", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/dashboard$/);
  });

  test("auth pages redirect to the dashboard too", async ({ page }) => {
    await page.goto("/signin");
    await expect(page).toHaveURL(/\/dashboard$/);
  });
});

test.describe("dashboard", () => {
  test("renders the mock user's data", async ({ page }) => {
    await page.goto("/dashboard");

    await expect(page.getByText("mock_user").first()).toBeVisible({ timeout: 15_000 });
  });
});
