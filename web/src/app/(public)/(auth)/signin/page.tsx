"use client";

import { GenericForm } from "@/components/ui/GenericForm";
import ProviderButton from "@/components/ui/ProviderButton";
import { loginSchema } from "@/lib/form/signin-schema";
import { useSignin, initiateOAuth } from "@/services/auth/mutations";
import Link from "next/link";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";


export default function Login() {

  const signin = useSignin();

  const errorMessage = signin.error?.message || '';
  const isEmailError = errorMessage.toLowerCase().includes('email');
  const isPasswordError = errorMessage.toLowerCase().includes('password') || errorMessage.toLowerCase().includes('credentials');

  const fieldErrors: Record<string, string> = {};
  if (isEmailError) fieldErrors.email = errorMessage;
  if (isPasswordError) fieldErrors.password = errorMessage;

  const generalError = (isEmailError || isPasswordError) ? undefined : errorMessage;

  return (
    <PageEnter className="w-full h-full flex items-center justify-center text-3xl">
      <RevealStagger className="w-11/12 flex flex-col justify-center items-center gap-y-8" stagger={0.07} delay={0.05}>
        <RevealItem className="flex flex-col items-center gap-2">
          <TextEffect as="h2" preset="fade-in-blur" per="char" speedReveal={1.4} className="text-2xl font-semibold tracking-tight">
            Sign in
          </TextEffect>
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <GenericForm
            schema={loginSchema}
            fields={[
              { name: "email", label: "Email", info: { type: "email", placeholder: "you@example.com" } },
              { name: "password", label: "Password", info: { type: "password", placeholder: "••••••••" } },
            ]}
            onSubmit={signin.mutate}
            submitLabel="Sign In"
            className="w-full"
            serverError={generalError}
            fieldErrors={fieldErrors}
          />
        </RevealItem>

        <RevealItem className="flex items-center gap-4 w-10/12 max-w-sm">
          <div className="flex-1 h-px bg-white/20" />
          <span className="text-sm text-white/60">or</span>
          <div className="flex-1 h-px bg-white/20" />
        </RevealItem>

        <RevealItem className="w-full flex justify-center">
          <ProviderButton name="google" icon="google" color="white" onClick={() => {
            initiateOAuth("google", "login");
          }} />
        </RevealItem>

        <RevealItem className="flex flex-col justify-center items-center gap-y-2">
          <Link href="recovery" className="text-sm hover:underline pointer-events-auto hover:cursor-pointer">
            Forgot Password?
          </Link>
          <Link className="text-sm hover:underline pointer-events-auto hover:cursor-pointer" href="signup">
            Don&apos;t have an account?
          </Link>
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}