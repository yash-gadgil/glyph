package handlers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
)

func buildSnapshot(ctx context.Context, portfolio userpb.PortfolioServiceClient, userID string) (string, error) {
	if portfolio == nil {
		return "", fmt.Errorf("portfolio service unavailable")
	}

	acct, err := portfolio.GetPortfolio(ctx, &userpb.UserSpecifier{UserId: userID})
	if err != nil {
		return "", fmt.Errorf("get portfolio: %w", err)
	}

	holdings, err := portfolio.GetHoldings(ctx, &userpb.UserSpecifier{UserId: userID})
	if err != nil {
		return "", fmt.Errorf("get holdings: %w", err)
	}

	cash := acct.CashBalanceCents
	marketValue := holdings.TotalMarketValueCents
	equity := cash + marketValue

	var b strings.Builder
	b.WriteString("PORTFOLIO SNAPSHOT\n")
	fmt.Fprintf(&b, "Currency: %s\n", acct.Currency)
	fmt.Fprintf(&b, "Total equity: %s (cash plus market value of positions)\n", dollars(equity))
	fmt.Fprintf(&b, "Cash available: %s\n", dollars(cash-acct.ReservedCashCents))
	fmt.Fprintf(&b, "Invested market value: %s\n", dollars(marketValue))
	if equity > 0 {
		fmt.Fprintf(&b, "Cash is %.1f%% of equity; invested is %.1f%%.\n", pct(cash, equity), pct(marketValue, equity))
	}
	fmt.Fprintf(&b, "Total unrealized P&L: %s (dollar amount, not a percentage)\n", dollars(holdings.TotalUnrealizedPnlCents))
	fmt.Fprintf(&b, "Total realized P&L: %s (dollar amount, not a percentage)\n", dollars(holdings.TotalRealizedPnlCents))

	rows := make([]*userpb.Holding, 0, len(holdings.Holdings))
	for _, h := range holdings.Holdings {
		if h.Qty != 0 {
			rows = append(rows, h)
		}
	}

	if len(rows) == 0 {
		b.WriteString("\nThe account holds no positions. It is entirely in cash.\n")
		return b.String(), nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].MarketValueCents > rows[j].MarketValueCents
	})

	tickers := make([]string, len(rows))
	for i, h := range rows {
		tickers[i] = h.Symbol
	}

	fmt.Fprintf(&b, "\nThese are the ONLY tickers held (%s). Do not mention any company or ticker not in this list.\n", strings.Join(tickers, ", "))
	b.WriteString("POSITIONS (largest first):\n")
	for _, h := range rows {
		weight := pct(h.MarketValueCents, equity)
		fmt.Fprintf(&b,
			"- %s: %d shares | avg cost %s | last price %s | market value %s (%.1f%% of equity) | unrealized P&L %s\n",
			h.Symbol, h.Qty, dollars(h.AvgPriceCents), dollars(h.LastPriceCents),
			dollars(h.MarketValueCents), weight, dollars(h.UnrealizedPnlCents),
		)
	}

	top := rows[0]
	fmt.Fprintf(&b, "\nLargest single position: %s at %.1f%% of equity.\n", top.Symbol, pct(top.MarketValueCents, equity))

	return b.String(), nil
}

func dollars(cents int64) string {
	neg := ""
	if cents < 0 {
		neg = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", neg, cents/100, cents%100)
}

func pct(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}
