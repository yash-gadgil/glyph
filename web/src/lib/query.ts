import { useQuery } from "@tanstack/react-query";
import { api } from "./api";

type QueryBuilderOptions<T> = {
  staleTime?: number;
  refetchInterval?: number;
  select?: (data: any) => T;
};

export default function queryBuilder<T = any>(
  queryKeys: string[],
  route: string,
  options?: number | QueryBuilderOptions<T>,
) {
  const opts: QueryBuilderOptions<T> =
    typeof options === "number" ? { staleTime: options } : options ?? {};
  return function query() {
    return useQuery({
      queryKey: queryKeys,
      queryFn: () => api(route),
      staleTime: opts.staleTime,
      refetchInterval: opts.refetchInterval,
      select: opts.select,
    });
  };
}
