import { useQuery, keepPreviousData } from "@tanstack/react-query";
import queryBuilder from "@/lib/query";
import { api } from "@/lib/api";

export const getWatchlists = queryBuilder(["watchlists"], "watchlists");

export function useWatchlist(watchlistId: string | null) {
  return useQuery({
    queryKey: ["watchlist", watchlistId],
    queryFn: () => api(`watchlists/${watchlistId}/info`),
    enabled: !!watchlistId,
    staleTime: 1000 * 60 * 5,
    placeholderData: keepPreviousData,
  });
}

export function searchSymbols(query: string) {
  return useQuery({
    queryKey: ["symbols", query],
    queryFn: () =>
      api(`watchlists/symbols?q=${encodeURIComponent(query)}&limit=20`),
    enabled: query.length >= 1,
    staleTime: 1000 * 60 * 5,
    placeholderData: keepPreviousData,
  });
}

