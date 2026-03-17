'use client';
import CreateWatchlistButton from "@/components/ui/CreateWatchlistButton";
import SymbolCard from "@/components/ui/SymbolCard";
import SymbolSearchInput from "@/components/ui/SymbolSearchInput";
import WatchlistButton from "@/components/ui/WatchlistButton";
import { socketForWatchlist } from "@/lib/socket";
import { getWatchlists, useWatchlist } from "@/services/watchlists/queries";
import { deleteWatchlist, modifyWatchlist } from "@/services/watchlists/mutations";
import { useQueryClient } from "@tanstack/react-query";
import { useState, useEffect, useRef, useMemo, useCallback } from "react";
import { Activity, LayoutList } from "lucide-react";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import MarketStatusBadge from "@/components/ui/MarketStatusBadge";
import { motion, AnimatePresence } from "motion/react";

export default function Watchlist() {

  const { data: watchlistsData, error, isLoading } = getWatchlists();
  const [currentWatchlistId, setCurrentWatchlistId] = useState<string | null>(null);

  useEffect(() => {
    if (watchlistsData?.w_metadata?.length && !currentWatchlistId) {
      setCurrentWatchlistId(watchlistsData.w_metadata[0].id);
    }
  }, [watchlistsData?.w_metadata, currentWatchlistId]);

  const { data: watchlistData } = useWatchlist(currentWatchlistId);

  const symbols: string[] = useMemo(
    () => watchlistData?.symbols ?? [],
    [watchlistData?.symbols]
  );

  const [closePrices, setClosePrices] = useState<Record<string, number>>({});
  const socketRef = useRef<WebSocket | null>(null);

  const qc = useQueryClient();
  const deleteWatchlistMutation = deleteWatchlist();
  const modifyWatchlistMutation = modifyWatchlist();

  const changeWatchlist = useCallback((id: string) => {
    setCurrentWatchlistId(id);
    setClosePrices({});
  }, []);

  const handleDeleteWatchlist = useCallback((id: string) => {
    const name = watchlistsData?.w_metadata?.find((w: { id: string }) => w.id === id)?.name ?? "this watchlist";
    if (!window.confirm(`Delete "${name}"? This cannot be undone.`)) return;
    deleteWatchlistMutation.mutate(id, {
      onSuccess: () => {
        qc.invalidateQueries({ queryKey: ["watchlists"] });
        if (currentWatchlistId === id) {
          const remaining = watchlistsData?.w_metadata?.filter((w: { id: string }) => w.id !== id);
          setCurrentWatchlistId(remaining?.[0]?.id ?? null);
          setClosePrices({});
        }
      },
    });
  }, [currentWatchlistId, watchlistsData, deleteWatchlistMutation, qc]);

  const handleRemoveSymbol = useCallback((symbol: string) => {
    if (!currentWatchlistId) return;
    modifyWatchlistMutation.mutate(
      { watchlistId: currentWatchlistId, action: "unsubscribe", symbols: [symbol] },
      {
        onSuccess: () => {
          qc.invalidateQueries({ queryKey: ["watchlists"] });
          qc.invalidateQueries({ queryKey: ["watchlist", currentWatchlistId] });
        },
      }
    );
  }, [currentWatchlistId, modifyWatchlistMutation, qc]);

  const symbolsKey = useMemo(() => [...symbols].sort().join(","), [symbols]);

  useEffect(() => {
    if (!currentWatchlistId || symbolsKey === "") return;

    const socket = socketForWatchlist(currentWatchlistId);
    socketRef.current = socket;

    socket.addEventListener("message", (e) => {
      const msg = JSON.parse(e.data);
      if (msg.symbol_bar) {
        const updates: Record<string, number> = {};
        for (const bar of msg.symbol_bar) {
          updates[bar.symbol] = bar.close;
        }
        setClosePrices((prev) => ({ ...prev, ...updates }));
      }
    });

    return () => socket.close();
  }, [currentWatchlistId, symbolsKey]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen w-full items-center justify-center p-8 bg-neutral-950">
        <div className="flex flex-col items-center gap-4">
          <div className="h-8 w-8 animate-spin rounded-full border-b-2 border-t-2 border-emerald-500"></div>
          <p className="text-sm font-mono text-neutral-400">Loading Watchlists...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen w-full items-center justify-center bg-neutral-950 p-8">
        <div className="rounded-xl border border-red-500/20 bg-red-500/10 p-6 text-red-400 font-mono text-sm max-w-lg shadow-xl shadow-red-900/10">
          <p className="mb-2 font-bold text-red-500 flex items-center gap-2">
            <Activity size={18} /> Error Retrieving Watchlists
          </p>
          <p className="opacity-80">{error instanceof Error ? error.message : 'Unknown error'}</p>
        </div>
      </div>
    );
  }

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-7xl space-y-8 pb-32" stagger={0.08}>
        <RevealItem>
          <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-white/10">
            <div className="space-y-1">
              <TextEffect
                as="h1"
                preset="fade-in-blur"
                per="word"
                className="text-3xl md:text-4xl font-semibold text-white tracking-tight drop-shadow-md"
              >
                Market Watchlists
              </TextEffect>
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.4, duration: 0.4 }}
                className="text-sm text-white/60"
              >
                <MarketStatusBadge />
              </motion.div>
            </div>
            <div className="flex items-center gap-4 z-50">
              <div className="w-full sm:w-80">
                <SymbolSearchInput
                  watchlistId={currentWatchlistId ?? ""}
                  currentSymbols={symbols}
                />
              </div>
            </div>
          </header>
        </RevealItem>

        <section className="flex flex-col space-y-8 mt-4">
          <RevealItem>
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6 z-40 relative">
              <div className="flex flex-wrap items-center gap-x-8 gap-y-4">
                {watchlistsData?.w_metadata?.map(({ name, id }: { name: string, id: string }) => (
                  <WatchlistButton
                    key={id}
                    name={name}
                    id={id}
                    selected={id === currentWatchlistId}
                    changeWatchlist={changeWatchlist}
                    onDelete={handleDeleteWatchlist}
                  />
                ))}
                <div className="pl-4 border-l border-white/20">
                  <CreateWatchlistButton />
                </div>

                {(!watchlistsData?.w_metadata || watchlistsData.w_metadata.length === 0) && (
                  <div className="text-xs text-white/50 italic">
                    No watchlists found.
                  </div>
                )}
              </div>

              <AnimatePresence mode="wait">
                {watchlistsData?.w_metadata?.find((w: { id: string }) => w.id === currentWatchlistId) && (
                  <motion.div
                    key={`${currentWatchlistId}-${symbols.length}`}
                    initial={{ opacity: 0, scale: 0.9, y: 4 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.9, y: -4 }}
                    transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
                    className="px-4 py-1.5 rounded-full bg-black/20 border border-white/10 backdrop-blur-md text-xs font-semibold text-white/70 self-start sm:self-center shrink-0 tracking-wide uppercase"
                  >
                    {symbols.length} Assets Tracked
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </RevealItem>

          <div className="min-h-[400px] z-10 relative">
            <AnimatePresence mode="wait">
              {symbols.length > 0 ? (
                <motion.div
                  key={`grid-${currentWatchlistId}`}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="flex flex-wrap gap-5"
                >
                  {symbols.map((symbol: string, idx: number) => (
                    <motion.div
                      key={symbol}
                      initial={{ opacity: 0, y: 18, scale: 0.96 }}
                      animate={{ opacity: 1, y: 0, scale: 1 }}
                      transition={{
                        duration: 0.45,
                        ease: [0.22, 1, 0.36, 1],
                        delay: Math.min(idx * 0.06, 0.4),
                      }}
                      className="min-w-fit pointer-events-auto"
                    >
                      <SymbolCard
                        symbol={symbol}
                        livePrice={closePrices[symbol]}
                        onRemove={handleRemoveSymbol}
                      />
                    </motion.div>
                  ))}
                </motion.div>
              ) : (
                <motion.div
                  key="empty"
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.4 }}
                >
                  <GlassSurface borderRadius={16} order="center" alignItems="center" flexDirection="col" className="h-72 mt-8 border-dashed text-center">
                    <div className="p-4 rounded-full bg-white/5 mb-5 border border-white/10 shadow-inner">
                      <LayoutList size={32} className="text-white/60" />
                    </div>
                    <p className="text-white font-medium mb-2 text-xl drop-shadow-sm">No symbols in this watchlist</p>
                    <p className="text-sm text-white/60 max-w-sm drop-shadow-sm">Use the search bar above to discover and track new ticker assets.</p>
                  </GlassSurface>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </section>
      </RevealStagger>
    </PageEnter>
  );
}