import { z } from "zod";

export const passwordChangeSchema = z
  .object({
    password: z
      .string()
      .regex(
        /^(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z])(?=.*\W)(?!.* ).{8,16}$/,
        "Password must be 8-16 chars and include upper, lower, number and symbol",
      ),

    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    path: ["confirmPassword"],
    message: "Passwords do not match",
  });

export type PasswordChangeSchemaType = z.infer<typeof passwordChangeSchema>;
