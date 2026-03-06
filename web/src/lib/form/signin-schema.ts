import { z } from "zod";

export const loginSchema = z.object({
  email: z.email("Invalid email address"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(100, "Password too long"),
});

export type LoginSchemaType = z.infer<typeof loginSchema>;
