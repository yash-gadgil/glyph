import { api } from "@/lib/api";
import { useMutation } from "@tanstack/react-query";

export type OrderPayload = {
  symbol: string;
  side: string;
  orderType: string;
  timeInForce: string;
  quantity: number;
  price?: number;
  stopPrice?: number;
};

export type OrderResponse = {
  id: string;
  userId: string;
  symbol: string;
  side: string;
  orderType: string;
  timeInForce: string;
  qty: number;
  filledQty: number;
  price: number;
  stopPrice: number;
  status: string;
  createdAt: string;
  updatedAt: string;
};

export function createOrder() {
  return useMutation({
    mutationFn: (order: OrderPayload) =>
      api("orders", {
        method: "POST",
        body: JSON.stringify(order),
      }),
    mutationKey: ["orders"],
  });
}

export function deleteOrder() {
  return useMutation({
    mutationFn: (orderId: string) =>
      api(`orders/${orderId}`, {
        method: "DELETE",
      }),
    mutationKey: ["orders"],
  });
}
