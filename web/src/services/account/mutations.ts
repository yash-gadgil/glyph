import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useResetAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api("account/reset", { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["account"] });
      qc.invalidateQueries({ queryKey: ["portfolio"] });
      qc.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}

export function useDeleteAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api("account", { method: "DELETE" }),
    onSuccess: async () => {
      try {
        await api("auth/signout", { method: "POST" });
      } catch {
      }
      try {
        await fetch("/api/signout", { method: "POST" });
      } catch {
      }
      qc.clear();
      window.location.assign("/");
    },
  });
}
