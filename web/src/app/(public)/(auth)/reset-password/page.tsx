'use client';
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { GenericForm } from "@/components/ui/GenericForm";
import { passwordChangeSchema } from "@/lib/form/password-change-schema";
import { useResetPassword } from "@/services/auth/mutations";
import Link from "next/link";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";

function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const resetPassword = useResetPassword();

  if (!token) {
    return (
      <PageEnter className="w-full h-full flex flex-col items-center justify-center gap-y-6 px-6 text-center">
        <TextEffect as="h2" preset="fade-in-blur" per="word" className="text-xl font-semibold">
          Invalid Reset Link
        </TextEffect>
        <p className="text-sm text-white/70 max-w-xs">
          This password reset link is invalid or has expired. Please request a new one.
        </p>
        <Link
          href="recovery"
          className="text-sm hover:underline pointer-events-auto hover:cursor-pointer mt-4 text-white/80"
        >
          Request New Reset Link
        </Link>
      </PageEnter>
    );
  }

  const handleSubmit = (data: { password: string; confirmPassword: string }) => {
    resetPassword.mutate({
      token,
      new_password: data.password,
    });
  };

  const errorMessage = resetPassword.error?.message || '';
  const isPasswordError = errorMessage.toLowerCase().includes('password');

  const fieldErrors: Record<string, string> = {};
  if (isPasswordError) fieldErrors.password = errorMessage;

  const generalError = isPasswordError ? undefined : errorMessage;

  return (
    <PageEnter className="w-full h-full flex flex-col items-center justify-center text-3xl gap-y-6">
      <RevealStagger className="w-full flex flex-col items-center gap-y-6" stagger={0.08} delay={0.05}>
        <RevealItem className="text-center mb-2 flex flex-col items-center gap-2">
          <TextEffect as="h2" preset="fade-in-blur" per="word" speedReveal={1.3} className="text-xl font-semibold">
            Reset Password
          </TextEffect>
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <GenericForm
            schema={passwordChangeSchema}
            fields={[
              { name: "password", label: "New Password", info: { type: "password", placeholder: "••••••••" } },
              { name: "confirmPassword", label: "Confirm Password", info: { type: "password", placeholder: "••••••••" } },
            ]}
            onSubmit={handleSubmit}
            submitLabel={resetPassword.isPending ? "Resetting…" : "Reset Password"}
            className="w-full"
            serverError={generalError}
            fieldErrors={fieldErrors}
          />
        </RevealItem>

        <RevealItem>
          <Link
            href="signin"
            className="text-sm hover:underline pointer-events-auto hover:cursor-pointer text-white/80"
          >
            Back to Sign In
          </Link>
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}

export default function ResetPassword() {
  return (
    <Suspense fallback={
      <div className="w-full h-full flex items-center justify-center">
        <p className="text-sm text-white/60">Loading...</p>
      </div>
    }>
      <ResetPasswordForm />
    </Suspense>
  );
}
