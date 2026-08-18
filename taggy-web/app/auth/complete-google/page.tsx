"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { AuthShell } from "@/components/auth-shell";
import { Loading } from "@/components/app-ui";
import { UnderlineField } from "@/components/underline-field";
import { useAuth } from "@/lib/auth";
import { completeGoogleRegistration } from "@/lib/api";
import { toastApiError } from "@/lib/toast";

function CompleteGoogleForm() {
  const { acceptTokens } = useAuth();
  const router = useRouter();
  const [registrationToken, setRegistrationToken] = useState("");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const hash = new URLSearchParams(window.location.hash.replace(/^#/, ""));
    setRegistrationToken(hash.get("registration_token") ?? "");
    setEmail(hash.get("email") ?? "");
    setName(hash.get("name") ?? "");
  }, []);

  return (
    <AuthShell
      mode="complete"
      title="Almost there"
      subtitle={
        email ? `Pick a username to finish signup for ${email}.` : "Pick a username to finish signup."
      }
      submitLabel="Create account"
      busy={busy}
      footer={
        <Link href="/login" className="text-foreground underline-offset-4 hover:underline">
          Cancel
        </Link>
      }
      onSubmit={async () => {
        setBusy(true);
        const result = await completeGoogleRegistration({
          registration_token: registrationToken,
          username,
          name: name || undefined,
        });
        setBusy(false);
        if (!result.ok || !result.data?.access_token) {
          toastApiError(result);
          return;
        }
        acceptTokens(result.data);
        router.replace("/home");
      }}
    >
      <UnderlineField
        id="username"
        label="Username"
        placeholder="Choose a username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        required
        minLength={3}
        maxLength={30}
        autoComplete="username"
      />
      <UnderlineField
        id="name"
        label="Name"
        placeholder="Your display name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        autoComplete="name"
      />
    </AuthShell>
  );
}

export default function CompleteGooglePage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center">
          <Loading />
        </div>
      }
    >
      <CompleteGoogleForm />
    </Suspense>
  );
}
