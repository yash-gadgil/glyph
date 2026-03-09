import { api } from "@/lib/api";
import { useMutation } from "@tanstack/react-query";

export function modifyWatchlist() {
  return useMutation({
    mutationFn: ({
      watchlistId,
      action,
      symbols,
    }: {
      watchlistId: string;
      action: "subscribe" | "unsubscribe";
      symbols: string[];
    }) =>
      api(`watchlists/${watchlistId}?action=${action}`, {
        method: "PATCH",
        body: JSON.stringify({
          symbols: symbols,
        }),
      }),
    mutationKey: ["watchlists"],
  });
}

export function createWatchlist() {
  return useMutation({
    mutationFn: (name: string) =>
      api("watchlists", {
        method: "POST",
        body: JSON.stringify({
          name: name,
        }),
      }),
    mutationKey: ["watchlists"],
  });
}

export function deleteWatchlist() {
  return useMutation({
    mutationFn: (watchlistId: string) =>
      api(`watchlists/${watchlistId}`, {
        method: "DELETE",
      }),
    mutationKey: ["watchlists"],
  });
}

export function deleteSymbolFromWatchlist() {
  return useMutation({
    mutationFn: ({
      watchlistId,
      symbol,
    }: {
      watchlistId: string;
      symbol: string;
    }) =>
      api(`watchlists/${watchlistId}?symbol=${symbol}`, {
        method: "DELETE",
      }),
    mutationKey: ["watchlists"],
  });
}
