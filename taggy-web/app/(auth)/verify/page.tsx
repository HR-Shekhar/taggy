"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { AuthShell } from "@/components/auth-shell";
import { Loading } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { UnderlineField } from "@/components/underline-field";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  resendVerification,
  verifyEmail,
} from "@/lib/api";

function VerifyForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [otp, setOtp] = useState(params.get("otp") ?? "");
  const [busy, setBusy] = useState(false);
  const [info, setInfo] = useState<string | null>(null);

  return (
    <AuthShell
      title="Verify"
      subtitle="Enter the 6-digit code we sent to your inbox."
      submitLabel="Verify email"
      busy={busy}
      footer={
        <Link href="/login" className="text-foreground underline-offset-4 hover:underline">
          Back to login
        </Link>
      }
      onSubmit={async () => {
        setBusy(true);
        setInfo(null);
        const result = await verifyEmail(email, otp);
        setBusy(false);
        if (!result.ok) {
          toastApiError(result);
          return;
        }
        router.push("/login");
      }}
    >
      <UnderlineField
        id="email"
        label="Email"
        type="email"
        placeholder="Enter your email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
        autoComplete="email"
      />
      <UnderlineField
        id="otp"
        label="One-time code"
        placeholder="6-digit code"
        value={otp}
        onChange={(e) => setOtp(e.target.value)}
        required
        maxLength={6}
        pattern="[0-9]{6}"
        inputMode="numeric"
        autoComplete="one-time-code"
      />
      {info && (
        <Alert>
          <AlertDescription>{info}</AlertDescription>
        </Alert>
      )}
      <Button
        variant="ghost"
        type="button"
        className="w-full text-sm text-muted-foreground"
        disabled={busy || !email}
        onClick={async () => {
          setBusy(true);
          const result = await resendVerification(email);
          setBusy(false);
          if (!result.ok && result.status !== 204) {
            toastApiError(result);
            return;
          }
          const otpDev = (result.data as { dev_otp?: string } | undefined)?.dev_otp;
          setInfo(
            otpDev ? `New OTP sent (dev): ${otpDev}` : "OTP resent if account exists."
          );
          if (otpDev) setOtp(otpDev);
        }}
      >
        Resend code
      </Button>
    </AuthShell>
  );
}

export default function VerifyPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center">
          <Loading />
        </div>
      }
    >
      <VerifyForm />
    </Suspense>
  );
}
