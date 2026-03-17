import { z } from "zod";

export enum Side {
  BUY = "buy",
  SELL = "sell",
}

export enum OrderType {
  MARKET = "market",
  LIMIT = "limit",
  STOP = "stop",
  STOP_LIMIT = "stop_limit",
}

export enum TimeInForce {
  DAY = "day",
  GTC = "gtc",
  IOC = "ioc",
  FOK = "fok",
}

export const orderSchema = z.object({
  symbol: z.string().min(1, "Symbol is required"),

  side: z.nativeEnum(Side).default(Side.BUY),

  orderType: z.nativeEnum(OrderType).default(OrderType.MARKET),

  quantity: z.coerce.number().int().positive("Quantity must be greater than zero"),

  price: z.coerce.number().positive().optional(),

  stopPrice: z.coerce.number().positive().optional(),

  timeInForce: z.nativeEnum(TimeInForce).default(TimeInForce.GTC),
});

export const orderSchemaWithRules = orderSchema.superRefine((data, ctx) => {
  if (data.orderType === OrderType.LIMIT && !data.price) {
    ctx.addIssue({
      path: ["price"],
      code: z.ZodIssueCode.custom,
      message: "Price is required for limit orders",
    });
  }

  if (data.orderType === OrderType.STOP && !data.stopPrice) {
    ctx.addIssue({
      path: ["stopPrice"],
      code: z.ZodIssueCode.custom,
      message: "Stop price is required for stop orders",
    });
  }

  if (data.orderType === OrderType.STOP_LIMIT) {
    if (!data.price) {
      ctx.addIssue({
        path: ["price"],
        code: z.ZodIssueCode.custom,
        message: "Price required for stop-limit",
      });
    }
    if (!data.stopPrice) {
      ctx.addIssue({
        path: ["stopPrice"],
        code: z.ZodIssueCode.custom,
        message: "Stop price required for stop-limit",
      });
    }
  }

  if (data.orderType === OrderType.MARKET && data.price) {
    ctx.addIssue({
      path: ["price"],
      code: z.ZodIssueCode.custom,
      message: "Market orders should not include price",
    });
  }
});

export type OrderSchemaType = z.infer<typeof orderSchemaWithRules>;
