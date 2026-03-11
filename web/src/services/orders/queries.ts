import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useOrders(status: string = "all") {
  return useQuery({
    queryKey: ["orders", status],
    queryFn: () => api(`orders?status=${status}`),
    staleTime: 5000,
    refetchInterval: 5000,
  });
}

export function useOrder(orderId: string) {
  return useQuery({
    queryKey: ["order", orderId],
    queryFn: () => api(`orders/${orderId}`),
    enabled: !!orderId,
  });
}
