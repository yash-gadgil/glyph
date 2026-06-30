"use client";

import { api } from "@/lib/api";
import { clearChatSession } from "@/services/advisor/chat";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

export function useSignin() {
  const router = useRouter();

  return useMutation({
    mutationFn: async (data: { email: string; password: string }) =>
      api("auth/signin", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    onSuccess: () => {
      router.push("dashboard");
    },
  });
}

export function useSignout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      try {
        await clearChatSession();
      } catch {
      }
      try {
        await api("auth/signout", { method: "POST" });
      } catch {
      }
      try {
        await fetch("/api/signout", { method: "POST" });
      } catch {
      }
    },

    onSettled: () => {
      window.dispatchEvent(new Event("kenaz:clear"));
      queryClient.clear();
      window.location.assign("/");
    },
  });
}

export function useSignup() {
  return useMutation({
    mutationFn: async (data: {
      name: string;
      email: string;
      password: string;
    }) =>
      api("auth/signup", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  });
}

export function useForgotPassword() {
  return useMutation({
    mutationFn: async (data: { email: string }) =>
      api("auth/forgot-password", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  });
}

export function useResetPassword() {
  const router = useRouter();

  return useMutation({
    mutationFn: async (data: { token: string; new_password: string }) =>
      api("auth/reset-password", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    onSuccess: () => {
      router.push("/dashboard");
    },
  });
}

type OAuthProvider = "google";

export async function initiateOAuth(
  provider: OAuthProvider,
  state: "login" | "register",
) {


  window.location.href = `http://localhost:8080/auth/oauth/${provider}?state=${state}`;
}

