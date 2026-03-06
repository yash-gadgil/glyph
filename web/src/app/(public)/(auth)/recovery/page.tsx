'use client';
import { useState } from "react";
import { GenericForm } from "@/components/ui/GenericForm";
import { recoverySchema } from "@/lib/form/recovery-schema";
import { useForgotPassword } from "@/services/auth/mutations";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";


export default function Recovery() {

  const forgotPassword = useForgotPassword();
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (data: { email: string }) => {
    forgotPassword.mutate(data, {
      onSuccess: () => setSubmitted(true),
      onError: () => setSubmitted(true),
    });
  };

  const successMessage = submitted ? "If an account exists, a reset link was sent." : undefined;

  return (
    <PageEnter className="w-full h-full flex flex-col items-center justify-center text-3xl gap-y-6">
      <RevealStagger className="w-full flex flex-col items-center gap-y-6" stagger={0.08} delay={0.05}>
        <RevealItem className="text-center mb-2 flex flex-col items-center gap-2">
          <TextEffect as="h2" preset="fade-in-blur" per="word" speedReveal={1.3} className="text-xl font-semibold">
            Forgot Password
          </TextEffect>
          <TextEffect as="p" preset="fade" per="word" delay={0.2} className="text-sm text-white/60 mt-1 px-6">
            Enter your email and we&apos;ll send you a reset link.
          </TextEffect>
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <GenericForm
            schema={recoverySchema}
            fields={[
              { name: "email", label: "Email", info: { type: "email", placeholder: "you@example.com" } },
            ]}
            onSubmit={handleSubmit}
            submitLabel={forgotPassword.isPending ? "Sending…" : "Send Reset Link"}
            className="w-full"
            serverError={forgotPassword.error?.message}
            successMessage={successMessage}
          />
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}