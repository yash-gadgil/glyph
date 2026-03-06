'use client';
import { useState } from "react";
import { GenericForm } from "@/components/ui/GenericForm";
import ProviderButton from "@/components/ui/ProviderButton";
import { signupSchema } from "@/lib/form/signup-schema";
import { initiateOAuth, useSignup } from "@/services/auth/mutations";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";


export default function Signup() {

  const signup = useSignup();
  const [submitted, setSubmitted] = useState(false);

  const error = signup.error as { message?: string; fieldErrors?: Record<string, string> } | null;
  const generalError = error?.message || (signup.isError ? 'An error occurred' : undefined);
  const fieldErrors = error?.fieldErrors || {};

  const successMessage = submitted ? "Verification link sent! Check your email." : undefined;

  return (
    <PageEnter className="w-full h-full flex items-center justify-center text-3xl">
      <RevealStagger className="w-11/12 flex flex-col justify-center items-center gap-y-7" stagger={0.07} delay={0.05}>
        <RevealItem className="flex flex-col items-center gap-2">
          <TextEffect as="h2" preset="fade-in-blur" per="char" speedReveal={1.4} className="text-2xl font-semibold tracking-tight">
            Create your account
          </TextEffect>
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <GenericForm
            schema={signupSchema}
            fields={[
              { name: "name", label: "Name *", info: { type: "text", placeholder: "Name" } },
              { name: "email", label: "Email *", info: { type: "email", placeholder: "you@example.com" } },
              { name: "password", label: "Password *", info: { type: "password", placeholder: "••••••••" } },
              { name: "confirmPassword", label: "Confirm Password *", info: { type: "password", placeholder: "••••••••" } },
            ]}
            onSubmit={(data) => signup.mutate(data, { onSuccess: () => setSubmitted(true) })}
            submitLabel={signup.isPending ? "Signing Up..." : "Sign Up"}
            serverError={generalError}
            successMessage={successMessage}
            fieldErrors={fieldErrors}
          />
        </RevealItem>

        <RevealItem className="flex items-center gap-4 w-10/12 max-w-sm">
          <div className="flex-1 h-px bg-white/20" />
          <span className="text-sm text-white/60">or</span>
          <div className="flex-1 h-px bg-white/20" />
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <ProviderButton icon="google" name="google" color="white" onClick={() => {
            initiateOAuth("google", "register");
          }} />
        </RevealItem>

      </RevealStagger>
    </PageEnter>
  );
}