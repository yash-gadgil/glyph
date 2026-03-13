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
