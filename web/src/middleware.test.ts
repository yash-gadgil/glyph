import { describe, expect, it } from "vitest";
import { NextRequest } from "next/server";
import { middleware } from "./middleware";

function request(path: string, cookies: Record<string, string> = {}): NextRequest {
  const req = new NextRequest(`http://localhost:3000${path}`);
  for (const [name, value] of Object.entries(cookies)) {
    req.cookies.set(name, value);
  }
  return req;
}

function redirectTarget(res: Response): string | null {
  const location = res.headers.get("location");
  return location ? new URL(location).pathname : null;
}

describe("middleware (no session)", () => {
  it("lets anonymous users reach public routes", () => {
    for (const path of ["/", "/signin", "/signup", "/recovery", "/reset-password"]) {
      const res = middleware(request(path));
      expect(res.status, path).toBe(200);
    }
  });

  it("redirects anonymous users away from app routes", () => {
    for (const path of ["/dashboard", "/portfolio", "/orders", "/watchlist", "/settings"]) {
      const res = middleware(request(path));
      expect(res.status, path).toBe(307);
      expect(redirectTarget(res), path).toBe("/");
    }
  });

  it("ignores next internals and files", () => {
    expect(middleware(request("/_next/static/chunk.js")).status).toBe(200);
    expect(middleware(request("/favicon.ico")).status).toBe(200);
    expect(middleware(request("/api/health")).status).toBe(200);
  });
});

describe("middleware (with session)", () => {
  it("redirects signed-in users from public routes to the dashboard", () => {
    for (const cookie of ["accessToken", "refreshToken"]) {
      const res = middleware(request("/signin", { [cookie]: "tok" }));
      expect(res.status, cookie).toBe(307);
      expect(redirectTarget(res), cookie).toBe("/dashboard");
    }
  });

  it("lets signed-in users reach app routes", () => {
    const res = middleware(request("/dashboard", { accessToken: "tok" }));
    expect(res.status).toBe(200);
  });
});
