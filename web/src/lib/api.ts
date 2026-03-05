import { isMockMode, mockApi } from "./mock";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type ApiErrorType = "http" | "parse" | "network";

export class ApiError extends Error {
  constructor(
    message: string,
    public type: ApiErrorType = "http",
    public statusCode?: number
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export async function api(
  path: string,
  options: RequestInit = {},
  retry = true
) {
  if (isMockMode()) {
    try {
      return await mockApi(path, options);
    } catch (err: any) {
      const status = typeof err?.statusCode === "number" ? err.statusCode : 500;
      throw new ApiError(err?.message ?? "Mock request failed", "http", status);
    }
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };

  let res: Response;

  try {
    res = await fetch(`${API_BASE_URL}/${path}`, {
      credentials: "include",
      headers,
      ...options,
    });
  } catch (error) {
    throw new ApiError("Unable to connect to server", "network");
  }

  if (res.status === 401 && retry) {
    const refreshed = await fetch(`${API_BASE_URL}/auth/refresh`, {
      credentials: "include",
    });

    if (refreshed.ok) {
      return api(path, options, false);
    } else {
      throw new ApiError("Session expired, Log in again", "http", 401);
    }
  }

  if (!res.ok) {
    let errorMessage = `Request failed with status ${res.status}`;

    try {
      const errorData = await res.json();
      errorMessage = errorData.message || errorMessage;
    } catch {}

    throw new ApiError(errorMessage, "http", res.status);
  }

  const contentType = res.headers.get("content-type");
  if (!contentType || !contentType.includes("application/json")) {
    return null;
  }

  try {
    return await res.json();
  } catch (error) {
    throw new ApiError("Invalid response format", "parse");
  }
}
