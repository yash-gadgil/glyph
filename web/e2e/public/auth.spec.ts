import { expect, test } from "@playwright/test";

test.describe("middleware guards (no session)", () => {
  test("app routes bounce anonymous visitors to the landing page", async ({ page }) => {
    for (const route of ["/dashboard", "/orders", "/portfolio"]) {
      await page.goto(route);
      await expect(page, route).toHaveURL(/\/$/);
    }
  });

  test("public pages stay reachable", async ({ page }) => {
    await page.goto("/signin");
    await expect(page).toHaveURL(/\/signin$/);

    await page.goto("/signup");
    await expect(page).toHaveURL(/\/signup$/);
  });
});

test.describe("sign in page", () => {
  test("renders the form with provider and recovery links", async ({ page }) => {
    await page.goto("/signin");

    await expect(page.getByText("Welcome back")).toBeVisible({ timeout: 15_000 });
    await expect(page.getByPlaceholder("you@example.com")).toBeVisible();
    await expect(page.getByRole("button", { name: /sign in/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /forgot password/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /don't have an account/i })).toBeVisible();
  });

  test("client-side validation rejects bad input", async ({ page }) => {
    await page.goto("/signin");
    await expect(page.getByPlaceholder("you@example.com")).toBeVisible({ timeout: 15_000 });

    await page.getByPlaceholder("you@example.com").fill("not-an-email");
    await page.getByPlaceholder("••••••••").fill("short");
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page.getByText(/invalid email/i)).toBeVisible();
    await expect(page.getByText(/at least 8 characters/i)).toBeVisible();
  });

  test("navigates to the signup page", async ({ page }) => {
    await page.goto("/signin");
    await page.getByRole("link", { name: /don't have an account/i }).click();
    await expect(page).toHaveURL(/\/signup$/);
  });
});

test.describe("sign up page", () => {
  test("rejects mismatched passwords client-side", async ({ page }) => {
    await page.goto("/signup");

    const password = page.getByPlaceholder("••••••••").first();
    await expect(password).toBeVisible({ timeout: 15_000 });

    const inputs = page.locator("input");
    const count = await inputs.count();
    for (let i = 0; i < count; i++) {
      const placeholder = (await inputs.nth(i).getAttribute("placeholder")) ?? "";
      if (placeholder.includes("@")) await inputs.nth(i).fill("e2e@example.com");
      else if (placeholder.includes("•")) await inputs.nth(i).fill("Passw0rd!");
      else await inputs.nth(i).fill("E2E User");
    }
    await page.locator('input[type="password"]').last().fill("Different1!");

    await page.getByRole("button", { name: /sign up|create/i }).click();
    await expect(page.getByText(/passwords do not match/i)).toBeVisible();
  });
});
