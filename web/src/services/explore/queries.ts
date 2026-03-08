import queryBuilder from '@/lib/query';
import { dollarsFromCents, timeAgo } from '@/lib/utils';

export interface NewsArticle {
  id: string;
  headline: string;
  summary: string;
  source: string;
  url: string;
  symbols: string[];
  imageUrl: string;
  timestamp: string;
}

export interface MarketMover {
  symbol: string;
  name: string;
  price: number;
  changePercent: number;
  volume: number;
  type: 'gainer' | 'loser';
}

type NewsApiArticle = {
  id: string;
  headline: string;
  summary: string;
  source: string;
  url: string;
  symbols: string[] | null;
  image_url: string;
  created_at: string;
};

type MoverApi = {
  symbol: string;
  company_name: string;
  price_cents: number;
  change_percent: number;
  volume: number;
};

export const useLatestNews = queryBuilder<NewsArticle[]>(['news', 'latest'], 'explore/news?limit=12', {
  staleTime: 5 * 60 * 1000,
  select: (data) => {
    const articles: NewsApiArticle[] = data?.articles ?? [];
    return articles.map((a) => ({
      id: a.id,
      headline: a.headline,
      summary: a.summary,
      source: a.source || 'Newswire',
      url: a.url || '#',
      symbols: a.symbols ?? [],
      imageUrl: a.image_url || '',
      timestamp: timeAgo(a.created_at),
    }));
  },
});

export const useTopMovers = queryBuilder<MarketMover[]>(['movers', 'top'], 'explore/movers', {
  staleTime: 60 * 1000,
  select: (data) => {
    const toMover = (m: MoverApi, type: 'gainer' | 'loser'): MarketMover => ({
      symbol: m.symbol,
      name: m.company_name || m.symbol,
      price: dollarsFromCents(m.price_cents),
      changePercent: m.change_percent,
      volume: m.volume,
      type,
    });
    const gainers: MarketMover[] = (data?.gainers ?? []).map((m: MoverApi) => toMover(m, 'gainer'));
    const losers: MarketMover[] = (data?.losers ?? []).map((m: MoverApi) => toMover(m, 'loser'));
    return [...gainers, ...losers];
  },
});
