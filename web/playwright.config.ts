import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    trace: "on-first-retry",
    ...devices["Desktop Chrome"],
  },
  projects: [
    {
      name: "app",
      testMatch: /app\/.*\.spec\.ts/,
      use: { baseURL: "http://localhost:3100" },
    },
    {
      name: "public",
      testMatch: /public\/.*\.spec\.ts/,
      use: { baseURL: "http://localhost:3101" },
    },
  ],
  webServer: [
    {
      command: "NEXT_PUBLIC_MOCK_API=true PORT=3100 pnpm dev",
      url: "http://localhost:3100",
      reuseExistingServer: !process.env.CI,
      timeout: 180_000,
    },
    {
      command: "NEXT_PUBLIC_MOCK_API=false PORT=3101 pnpm dev",
      url: "http://localhost:3101",
      reuseExistingServer: !process.env.CI,
      timeout: 180_000,
    },
  ],
});
