"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { AuthShell } from "@/components/auth-shell";
import { UnderlineField } from "@/components/underline-field";
import { resendVerification, verifyEmail } from "@/lib/api";
import { toastApiError, toastError, toastSuccess } from "@/lib/toast";

function VerifyForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const e = params.get("email");
    if (e) setEmail(decodeURIComponent(e));
  }, [params]);

  async function onVerify() {
    if (!email.trim() || !otp.trim()) {
      toastError("Enter the code from your email.");
      return;
    }
    setBusy(true);
    const res = await verifyEmail(email.trim(), otp.trim());
    setBusy(false);
    if (!res.ok) {
      toastApiError(res);
      return;
    }
    toastSuccess("Email verified. You can sign in now.");
    router.push("/login");
  }

  async function onResend() {
    if (!email.trim()) {
      toastError("No email on file. Go back and sign up again.");
      return;
    }
    setBusy(true);
    const res = await resendVerification(email.trim());
    setBusy(false);
    if (!res.ok) {
      toastApiError(res);
      return;
    }
    toastSuccess("Verification code sent again.");
  }

  return (
    <AuthShell
      title="Verify your email"
      subtitle="Enter the code we sent to your inbox."
      submitLabel="Verify email"
      busy={busy}
      onSubmit={onVerify}
      footer={
        <Link href="/login" className="font-medium text-primary hover:underline">
          Back to sign in
        </Link>
      }
    >
      {email ? (
        <div className="space-y-2">
          <p className="text-[10px] font-medium uppercase tracking-[0.2em] text-foreground/60">
            Email
          </p>
          <p className="border-b border-foreground/25 pb-1.5 text-sm text-foreground">
            {email}
          </p>
          <Link
            href="/register"
            className="inline-block text-xs text-primary hover:underline"
          >
            Use a different email
          </Link>
        </div>
      ) : (
        <p className="text-sm text-foreground/75">
          No email provided.{" "}
          <Link href="/register" className="text-primary hover:underline">
            Sign up again
          </Link>
        </p>
      )}
      <UnderlineField
        id="otp"
        label="Verification code"
        type="text"
        inputMode="numeric"
        autoComplete="one-time-code"
        value={otp}
        onChange={(e) => setOtp(e.target.value)}
        placeholder="6-digit code"
        required
      />
      <button
        type="button"
        disabled={busy || !email.trim()}
        onClick={() => void onResend()}
        className="w-full text-sm text-foreground/75 hover:text-foreground disabled:opacity-50"
      >
        Resend code
      </button>
    </AuthShell>
  );
}

export default function VerifyPage() {
  return (
    <Suspense fallback={<div className="p-8 text-center text-sm">Loading…</div>}>
      <VerifyForm />
    </Suspense>
  );
}
