'use client';

import { useState } from "react";
import { useOrders } from "@/services/orders/queries";
import { createOrder, deleteOrder, OrderPayload, OrderResponse } from "@/services/orders/mutations";
import { useQueryClient } from "@tanstack/react-query";
import { Side, OrderType, TimeInForce } from "@/lib/form/order-schema";
import {
    Activity, ArrowDownRight, ArrowUpRight, Box, ChevronDown,
    Clock, ShieldAlert, Trash2, Zap
} from "lucide-react";
import PixelHover from "@/components/ui/PixelHover";
import SymbolCombobox from "@/components/ui/SymbolCombobox";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import MarketStatusBadge, { useMarketOpen } from "@/components/ui/MarketStatusBadge";
import { motion, AnimatePresence } from "motion/react";

type StatusTab = "all" | "open" | "filled" | "cancelled";

function formatOrderTime(value: string): string {
    if (!value) return "-";
    const date = /^\d+$/.test(value) ? new Date(Number(value) * 1000) : new Date(value);
    if (Number.isNaN(date.getTime())) return "-";
    return date.toLocaleString(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });
}

export default function Orders() {
    const [activeTab, setActiveTab] = useState<StatusTab>("all");
    const { data: ordersData, isLoading, error } = useOrders(activeTab);
    const orderMutation = createOrder();
    const deleteMutation = deleteOrder();
    const qc = useQueryClient();
    const marketOpen = useMarketOpen();

    const orders: OrderResponse[] = ordersData?.orders || [];

    const [symbol, setSymbol] = useState("");
    const [side, setSide] = useState<Side>(Side.BUY);
    const [orderType, setOrderType] = useState<OrderType>(OrderType.MARKET);
    const [quantity, setQuantity] = useState("");
    const [price, setPrice] = useState("");
    const [stopPrice, setStopPrice] = useState("");
    const [timeInForce, setTimeInForce] = useState<TimeInForce>(TimeInForce.GTC);
    const [formError, setFormError] = useState("");

    const showPrice = orderType === OrderType.LIMIT || orderType === OrderType.STOP_LIMIT;
    const showStopPrice = orderType === OrderType.STOP || orderType === OrderType.STOP_LIMIT;

    const handlePlaceOrder = () => {
        setFormError("");

        if (!symbol.trim()) { setFormError("Symbol is required"); return; }
        const qty = parseInt(quantity);
        if (!qty || qty <= 0) { setFormError("Quantity must be a positive integer"); return; }

        const payload: OrderPayload = {
            symbol: symbol.toUpperCase().trim(),
            side,
            orderType,
            timeInForce,
            quantity: qty,
        };

        if (showPrice) {
            const p = parseInt(price);
            if (!p || p <= 0) { setFormError("Price is required for this order type"); return; }
            payload.price = p;
        }
        if (showStopPrice) {
            const sp = parseInt(stopPrice);
            if (!sp || sp <= 0) { setFormError("Stop price is required for this order type"); return; }
            payload.stopPrice = sp;
        }

        orderMutation.mutate(payload, {
            onSuccess: () => {
                qc.invalidateQueries({ queryKey: ["orders"] });
                setSymbol(""); setQuantity(""); setPrice(""); setStopPrice("");
                setFormError("");
            },
            onError: (err: Error) => {
                setFormError(err.message || "Failed to place order");
            }
        });
    };

    const handleCancelOrder = (id: string) => {
        if (!window.confirm("Cancel this order?")) return;
        deleteMutation.mutate(id, {
            onSuccess: () => qc.invalidateQueries({ queryKey: ["orders"] }),
        });
    };

    const tabs: { label: string; value: StatusTab }[] = [
        { label: "All", value: "all" },
        { label: "Open", value: "open" },
        { label: "Filled", value: "filled" },
        { label: "Cancelled", value: "cancelled" },
    ];

    const isBuy = side === Side.BUY;

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
                                className="text-3xl md:text-4xl font-semibold text-white tracking-tight drop-shadow-md flex items-center gap-3"
                            >
                                Order Management
                            </TextEffect>
                            <motion.div
                                initial={{ opacity: 0, y: 6 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.4, duration: 0.4 }}
                                className="text-sm text-white/60"
                            >
                                <MarketStatusBadge />
                            </motion.div>
                        </div>
                    </header>
                </RevealItem>

                <section className="grid grid-cols-1 lg:grid-cols-3 gap-8">

                    <RevealItem className="order-2 lg:order-1 col-span-1 lg:col-span-2 flex flex-col space-y-4">

                        <div className="flex flex-wrap items-center gap-x-1.5 gap-y-2 relative">
                            {tabs.map(tab => {
                                const active = activeTab === tab.value;
                                return (
                                    <button
                                        key={tab.value}
                                        onClick={() => setActiveTab(tab.value)}
                                        className={`relative px-3 sm:px-4 py-2 text-xs font-bold uppercase tracking-wider rounded-lg transition-colors ${active
                                            ? "text-white"
                                            : "text-white/40 hover:text-white/70"
                                            }`}
                                    >
                                        {active && (
                                            <motion.div
                                                layoutId="orders-tab-pill"
                                                className="absolute inset-0 bg-white/10 border border-white/20 rounded-lg shadow-sm"
                                                transition={{ type: "spring", stiffness: 400, damping: 30 }}
                                            />
                                        )}
                                        <span className="relative z-10">{tab.label}</span>
                                    </button>
                                );
                            })}
                            <AnimatePresence mode="wait">
                                <motion.span
                                    key={`${activeTab}-${orders.length}`}
                                    initial={{ opacity: 0, scale: 0.9 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    exit={{ opacity: 0, scale: 0.9 }}
                                    transition={{ duration: 0.2 }}
                                    className="ml-auto px-3 py-1 bg-white/5 rounded-full border border-white/10 text-xs text-white/50 font-medium"
                                >
                                    {orders.length} {activeTab === "all" ? "Total" : activeTab}
                                </motion.span>
                            </AnimatePresence>
                        </div>

                        <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" className="w-full">
                            {error && (
                                <div className="p-4 bg-red-500/10 border-b border-red-500/20 text-red-400 text-sm flex items-center gap-2">
                                    <ShieldAlert size={16} /> Error retrieving orders
                                </div>
                            )}

                            <div className="overflow-x-auto">
                                <table className="w-full text-left text-sm whitespace-nowrap">
                                    <thead className="bg-white/5 border-b border-white/10 text-white/50 text-xs uppercase tracking-wider">
                                        <tr>
                                            <th className="px-5 py-4 font-medium">Symbol</th>
                                            <th className="px-5 py-4 font-medium">Side</th>
                                            <th className="px-5 py-4 font-medium">Type</th>
                                            <th className="px-5 py-4 font-medium text-right">Qty</th>
                                            <th className="px-5 py-4 font-medium text-right">Filled</th>
                                            <th className="px-5 py-4 font-medium text-right">Price</th>
                                            <th className="px-5 py-4 font-medium">Status</th>
                                            <th className="px-5 py-4 font-medium">Time</th>
                                            <th className="px-5 py-4 font-medium text-right">Action</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-white/5">
                                        {isLoading ? (
                                            <tr>
                                                <td colSpan={9} className="px-6 py-12 text-center">
                                                    <div className="flex items-center justify-center gap-2 text-white/40">
                                                        <Activity size={16} className="animate-spin" /> Loading orders...
                                                    </div>
                                                </td>
                                            </tr>
                                        ) : orders.length > 0 ? (
                                            orders.map((order, idx) => {
                                                const orderIsBuy = order.side === "buy";
                                                const canCancel = ["pending", "open", "partial_fill"].includes(order.status);
                                                return (
                                                    <motion.tr
                                                        key={order.id || idx}
                                                        initial={{ opacity: 0, x: -16 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.25 + idx * 0.04, duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
                                                        className="hover:bg-white/5 transition-colors group"
                                                    >
                                                        <td className="px-5 py-4 font-bold tracking-wider text-white">{order.symbol}</td>
                                                        <td className="px-5 py-4">
                                                            <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-bold border ${orderIsBuy
                                                                ? "text-emerald-400 border-emerald-500/20 bg-emerald-500/10"
                                                                : "text-red-400 border-red-500/20 bg-red-500/10"
                                                                }`}>
                                                                {orderIsBuy ? <ArrowDownRight size={12} /> : <ArrowUpRight size={12} />}
                                                                {order.side.toUpperCase()}
                                                            </span>
                                                        </td>
                                                        <td className="px-5 py-4 text-white/60 uppercase text-xs font-medium">{order.orderType.replace("_", " ")}</td>
                                                        <td className="px-5 py-4 text-right text-white/80 font-medium">{order.qty}</td>
                                                        <td className="px-5 py-4 text-right text-white/60">{order.filledQty}/{order.qty}</td>
                                                        <td className="px-5 py-4 text-right text-white/70">
                                                            {order.price ? `$${(order.price / 100).toFixed(2)}` : "-"}
                                                        </td>
                                                        <td className="px-5 py-4">
                                                            <OrderStatusBadge status={order.status} />
                                                        </td>
                                                        <td className="px-5 py-4 text-white/50 text-xs tabular-nums">
                                                            {formatOrderTime(order.createdAt)}
                                                        </td>
                                                        <td className="px-5 py-4 text-right">
                                                            {canCancel ? (
                                                                <button
                                                                    onClick={() => handleCancelOrder(order.id)}
                                                                    className="p-2 text-white/30 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all active:scale-95 opacity-0 group-hover:opacity-100"
                                                                    title="Cancel Order"
                                                                >
                                                                    <Trash2 size={14} />
                                                                </button>
                                                            ) : (
                                                                <span className="text-white/20">-</span>
                                                            )}
                                                        </td>
                                                    </motion.tr>
                                                );
                                            })
                                        ) : (
                                            <tr>
                                                <td colSpan={9} className="px-6 py-16 text-center">
                                                    <div className="flex flex-col items-center justify-center gap-3">
                                                        <div className="p-3 bg-white/5 rounded-full border border-white/5 shadow-inner">
                                                            <Box size={24} className="text-white/30" />
                                                        </div>
                                                        <p className="text-white/50 text-sm font-medium">No orders found</p>
                                                        <p className="text-white/30 text-xs">Place an order using the form to get started.</p>
                                                    </div>
                                                </td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            </div>
                        </GlassSurface>
                    </RevealItem>

                    <RevealItem className="order-1 lg:order-2 col-span-1 flex flex-col space-y-4">
                        <div className="flex items-center gap-2">
                            <h2 className="text-sm font-bold uppercase tracking-wider text-white/70">Place Order</h2>
                        </div>

                        <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" innerClassName="p-6 space-y-5">

                            <div className="space-y-1.5">
                                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Symbol</label>
                                <SymbolCombobox
                                    value={symbol}
                                    onChange={(v) => setSymbol(v.toUpperCase())}
                                    placeholder="AAPL"
                                    inputClassName="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                                />
                            </div>

                            <div className="space-y-1.5">
                                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Side</label>
                                <div className="grid grid-cols-2 gap-2">
                                    <PixelHover
                                        variant="emerald"
                                        active={isBuy}
                                        className={`rounded-xl border transition-all ${isBuy
                                            ? "bg-emerald-500/20 border-emerald-500/30  "
                                            : "bg-white/5 border-white/10 hover:bg-white/10 hover:border-emerald-500/40"
                                            }`}
                                    >
                                        <button
                                            onClick={() => setSide(Side.BUY)}
                                            className={`w-full py-3 rounded-xl text-sm font-bold uppercase tracking-wider bg-transparent ${isBuy ? "text-emerald-400" : "text-white/40"
                                                }`}
                                        >
                                            Buy
                                        </button>
                                    </PixelHover>
                                    <PixelHover
                                        variant="red"
                                        active={!isBuy}
                                        className={`rounded-xl border transition-all ${!isBuy
                                            ? "bg-red-500/20 border-red-500/30  "
                                            : "bg-white/5 border-white/10 hover:bg-white/10 hover:border-red-500/40"
                                            }`}
                                    >
                                        <button
                                            onClick={() => setSide(Side.SELL)}
                                            className={`w-full py-3 rounded-xl text-sm font-bold uppercase tracking-wider bg-transparent ${!isBuy ? "text-red-400" : "text-white/40"
                                                }`}
                                        >
                                            Sell
                                        </button>
                                    </PixelHover>
                                </div>
                            </div>

                            <div className="space-y-1.5">
                                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Order Type</label>
                                <div className="relative">
                                    <select
                                        value={orderType}
                                        onChange={e => setOrderType(e.target.value as OrderType)}
                                        className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white appearance-none focus:outline-none focus:border-emerald-500/50 transition-all cursor-pointer"
                                    >
                                        <option value={OrderType.MARKET}>Market</option>
                                        <option value={OrderType.LIMIT}>Limit</option>
                                        <option value={OrderType.STOP}>Stop</option>
                                        <option value={OrderType.STOP_LIMIT}>Stop Limit</option>
                                    </select>
                                    <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none" />
                                </div>
                            </div>

                            <div className="space-y-1.5">
                                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Quantity</label>
                                <input
                                    type="number"
                                    value={quantity}
                                    onChange={e => setQuantity(e.target.value)}
                                    placeholder="0"
                                    min="1"
                                    className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                                />
                            </div>

                            <AnimatePresence>
                                {showPrice && (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0, y: -8 }}
                                        animate={{ opacity: 1, height: "auto", y: 0 }}
                                        exit={{ opacity: 0, height: 0, y: -8 }}
                                        transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
                                        className="space-y-1.5 overflow-hidden"
                                    >
                                        <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Limit Price (¢)</label>
                                        <input
                                            type="number"
                                            value={price}
                                            onChange={e => setPrice(e.target.value)}
                                            placeholder="e.g. 15025 = $150.25"
                                            min="1"
                                            className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                                        />
                                    </motion.div>
                                )}
                            </AnimatePresence>

                            <AnimatePresence>
                                {showStopPrice && (
                                    <motion.div
                                        initial={{ opacity: 0, height: 0, y: -8 }}
                                        animate={{ opacity: 1, height: "auto", y: 0 }}
                                        exit={{ opacity: 0, height: 0, y: -8 }}
                                        transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
                                        className="space-y-1.5 overflow-hidden"
                                    >
                                        <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Stop Price (¢)</label>
                                        <input
                                            type="number"
                                            value={stopPrice}
                                            onChange={e => setStopPrice(e.target.value)}
                                            placeholder="e.g. 14900 = $149.00"
                                            min="1"
                                            className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                                        />
                                    </motion.div>
                                )}
                            </AnimatePresence>

                            <div className="space-y-1.5">
                                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Time in Force</label>
                                <div className="relative">
                                    <select
                                        value={timeInForce}
                                        onChange={e => setTimeInForce(e.target.value as TimeInForce)}
                                        className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white appearance-none focus:outline-none focus:border-emerald-500/50 transition-all cursor-pointer"
                                    >
                                        <option value={TimeInForce.DAY}>DAY</option>
                                        <option value={TimeInForce.GTC}>GTC (Good Till Cancel)</option>
                                        <option value={TimeInForce.IOC}>IOC (Immediate or Cancel)</option>
                                        <option value={TimeInForce.FOK}>FOK (Fill or Kill)</option>
                                    </select>
                                    <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none" />
                                </div>
                            </div>

                            <AnimatePresence>
                                {formError && (
                                    <motion.div
                                        initial={{ opacity: 0, y: -6 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: -6 }}
                                        className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium flex items-center gap-2"
                                    >
                                        <ShieldAlert size={14} /> {formError}
                                    </motion.div>
                                )}
                            </AnimatePresence>

                            <AnimatePresence>
                                {marketOpen === false && (
                                    <motion.div
                                        initial={{ opacity: 0, y: -6 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: -6 }}
                                        className="p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-300/90 text-xs font-medium flex items-center gap-2"
                                    >
                                        <Clock size={14} className="shrink-0" />
                                        Market closed - order will queue until the next session
                                        (9:30 AM ET).
                                    </motion.div>
                                )}
                            </AnimatePresence>

                            <PixelHover
                                variant={isBuy ? "emerald" : "red"}
                                className={`group w-full rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 transition-all active:scale-[0.98] ${isBuy ? "hover:border-emerald-500/40" : "hover:border-red-500/40"
                                    }`}
                            >
                                <button
                                    onClick={handlePlaceOrder}
                                    disabled={orderMutation.isPending}
                                    className={`w-full py-3.5 rounded-xl text-sm font-bold uppercase tracking-wider text-white bg-transparent transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${isBuy ? "group-hover:text-emerald-300" : "group-hover:text-red-300"
                                        }`}
                                >
                                    {orderMutation.isPending ? "Submitting..." : `Place ${isBuy ? "Buy" : "Sell"} Order`}
                                </button>
                            </PixelHover>

                            <AnimatePresence>
                                {symbol && quantity && (
                                    <motion.div
                                        initial={{ opacity: 0 }}
                                        animate={{ opacity: 1 }}
                                        exit={{ opacity: 0 }}
                                        className="text-xs text-white/40 text-center pt-1"
                                    >
                                        {isBuy ? "Buy" : "Sell"} {quantity} × {symbol.toUpperCase()} • {orderType.toUpperCase().replace("_", " ")} • {timeInForce.toUpperCase()}
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </GlassSurface>
                    </RevealItem>

                </section>
            </RevealStagger>
        </PageEnter>
    );
}

function OrderStatusBadge({ status }: { status: string }) {
    const config: Record<string, { color: string; icon: React.ReactNode }> = {
        pending: { color: "text-amber-400 bg-amber-500/10 border-amber-500/20", icon: <Clock size={10} /> },
        open: { color: "text-blue-400 bg-blue-500/10 border-blue-500/20", icon: <Activity size={10} /> },
        partial_fill: { color: "text-purple-400 bg-purple-500/10 border-purple-500/20", icon: <Activity size={10} /> },
        filled: { color: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20", icon: <Zap size={10} /> },
        cancelled: { color: "text-white/40 bg-white/5 border-white/10", icon: <Trash2 size={10} /> },
        rejected: { color: "text-red-400 bg-red-500/10 border-red-500/20", icon: <ShieldAlert size={10} /> },
    };

    const c = config[status] || config.pending;

    return (
        <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border ${c.color}`}>
            {c.icon}
            {status.replace("_", " ")}
        </span>
    );
}
