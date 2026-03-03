'use client';
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";


const q = new QueryClient();

export default function ReactQueryProvider(
  { children }: { children: ReactNode }
) {
  return (
    <QueryClientProvider client={q} >
      {children}
    </QueryClientProvider>
  );
}