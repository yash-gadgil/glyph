import queryBuilder from "@/lib/query";

export const getUserId = queryBuilder(["account", "me"], "account/me");

export const getAccount = queryBuilder(["account"], "account");

export const getProfile = queryBuilder(["account", "profile"], "account/profile");

export const getTrades = queryBuilder(["account", "trades"], "account/trades");
