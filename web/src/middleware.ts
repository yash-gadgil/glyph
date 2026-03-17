import { NextRequest, NextResponse } from "next/server";

const PUBLIC_ROUTES = [
  "/",
  "/signin",
  "/signup",
  "/verify",
  "/callback",
  "/recovery",
  "/reset-password"
];

const MOCK_MODE = process.env.NEXT_PUBLIC_MOCK_API === "true";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  const hasSession =
    MOCK_MODE ||
    req.cookies.has("refreshToken") ||
    req.cookies.has("accessToken");

  if (PUBLIC_ROUTES.includes(pathname) && hasSession) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  const isAppRoute =
    !PUBLIC_ROUTES.includes(pathname) &&
    !pathname.startsWith("/_next") &&
    !pathname.startsWith("/api") &&
    !pathname.includes(".");

  if (isAppRoute && !hasSession) {
    return NextResponse.redirect(new URL("/", req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
