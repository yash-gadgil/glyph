import { describe, expect, it } from "vitest";
import { orderSchemaWithRules, OrderType, Side, TimeInForce } from "./order-schema";
import { signupSchema } from "./signup-schema";
import { loginSchema } from "./signin-schema";
import { recoverySchema } from "./recovery-schema";
import { passwordChangeSchema } from "./password-change-schema";

const baseOrder = {
  symbol: "AAPL",
  side: Side.BUY,
  orderType: OrderType.MARKET,
  quantity: 10,
  timeInForce: TimeInForce.GTC,
};

function issuePaths(result: { success: boolean; error?: { issues: { path: PropertyKey[] }[] } }) {
  return result.success ? [] : result.error!.issues.map((i) => i.path.join("."));
}

describe("orderSchemaWithRules", () => {
  it("accepts a market order without price", () => {
    expect(orderSchemaWithRules.safeParse(baseOrder).success).toBe(true);
  });

  it("rejects a market order with a price", () => {
    const result = orderSchemaWithRules.safeParse({ ...baseOrder, price: 100 });
    expect(result.success).toBe(false);
    expect(issuePaths(result)).toContain("price");
  });

  it("requires price for limit orders", () => {
    const missing = orderSchemaWithRules.safeParse({
      ...baseOrder,
      orderType: OrderType.LIMIT,
    });
    expect(issuePaths(missing)).toContain("price");

    const ok = orderSchemaWithRules.safeParse({
      ...baseOrder,
      orderType: OrderType.LIMIT,
      price: 189.5,
    });
    expect(ok.success).toBe(true);
  });

  it("requires stopPrice for stop orders", () => {
    const missing = orderSchemaWithRules.safeParse({
      ...baseOrder,
      orderType: OrderType.STOP,
    });
    expect(issuePaths(missing)).toContain("stopPrice");
  });

  it("requires both prices for stop-limit orders", () => {
    const missing = orderSchemaWithRules.safeParse({
      ...baseOrder,
      orderType: OrderType.STOP_LIMIT,
    });
    const paths = issuePaths(missing);
    expect(paths).toContain("price");
    expect(paths).toContain("stopPrice");

    const ok = orderSchemaWithRules.safeParse({
      ...baseOrder,
      orderType: OrderType.STOP_LIMIT,
      price: 100,
      stopPrice: 95,
    });
    expect(ok.success).toBe(true);
  });

  it("rejects empty symbol and non-positive quantities", () => {
    expect(orderSchemaWithRules.safeParse({ ...baseOrder, symbol: "" }).success).toBe(false);
    expect(orderSchemaWithRules.safeParse({ ...baseOrder, quantity: 0 }).success).toBe(false);
    expect(orderSchemaWithRules.safeParse({ ...baseOrder, quantity: -1 }).success).toBe(false);
    expect(orderSchemaWithRules.safeParse({ ...baseOrder, quantity: 1.5 }).success).toBe(false);
  });

  it("coerces string quantities from form inputs", () => {
    const result = orderSchemaWithRules.safeParse({ ...baseOrder, quantity: "12" });
    expect(result.success).toBe(true);
    if (result.success) expect(result.data.quantity).toBe(12);
  });
});

describe("signupSchema", () => {
  const valid = {
    name: "Yash",
    email: "yash@example.com",
    password: "Passw0rd!",
    confirmPassword: "Passw0rd!",
  };

  it("accepts a valid signup", () => {
    expect(signupSchema.safeParse(valid).success).toBe(true);
  });

  it("rejects mismatched passwords on the confirm field", () => {
    const result = signupSchema.safeParse({ ...valid, confirmPassword: "Different1!" });
    expect(result.success).toBe(false);
    expect(issuePaths(result)).toContain("confirmPassword");
  });

  it("enforces the password complexity rules", () => {
    for (const bad of ["short1!", "alllowercase1!", "ALLUPPERCASE1!", "NoNumbers!", "NoSymbols1", "Has Space1!"]) {
      expect(
        signupSchema.safeParse({ ...valid, password: bad, confirmPassword: bad }).success,
        `password ${JSON.stringify(bad)} should be rejected`
      ).toBe(false);
    }
  });

  it("bounds the name length", () => {
    expect(signupSchema.safeParse({ ...valid, name: "" }).success).toBe(false);
    expect(signupSchema.safeParse({ ...valid, name: "x".repeat(21) }).success).toBe(false);
  });

  it("rejects invalid emails", () => {
    expect(signupSchema.safeParse({ ...valid, email: "not-an-email" }).success).toBe(false);
  });
});

describe("loginSchema", () => {
  it("accepts valid credentials", () => {
    expect(loginSchema.safeParse({ email: "a@b.co", password: "12345678" }).success).toBe(true);
  });

  it("rejects bad email or short password", () => {
    expect(loginSchema.safeParse({ email: "nope", password: "12345678" }).success).toBe(false);
    expect(loginSchema.safeParse({ email: "a@b.co", password: "short" }).success).toBe(false);
  });
});

describe("recoverySchema", () => {
  it("validates the email", () => {
    expect(recoverySchema.safeParse({ email: "a@b.co" }).success).toBe(true);
    expect(recoverySchema.safeParse({ email: "nope" }).success).toBe(false);
  });
});

describe("passwordChangeSchema", () => {
  it("accepts matching strong passwords", () => {
    expect(
      passwordChangeSchema.safeParse({ password: "Passw0rd!", confirmPassword: "Passw0rd!" }).success
    ).toBe(true);
  });

  it("rejects mismatches and weak passwords", () => {
    expect(
      passwordChangeSchema.safeParse({ password: "Passw0rd!", confirmPassword: "Other1!aa" }).success
    ).toBe(false);
    expect(
      passwordChangeSchema.safeParse({ password: "weak", confirmPassword: "weak" }).success
    ).toBe(false);
  });
});
