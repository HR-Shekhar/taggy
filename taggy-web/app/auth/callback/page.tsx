"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import { toastError } from "@/lib/toast";

function parseHash() {
  if (typeof window === "undefined") return new URLSearchParams();
  const raw = window.location.hash.replace(/^#/, "");
  return new URLSearchParams(raw);
}

export default function AuthCallbackPage() {
  const { acceptTokens } = useAuth();
  const router = useRouter();
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const query = new URLSearchParams(window.location.search);
    const err = query.get("error");
    if (err) {
      toastError(err);
      setFailed(true);
      return;
    }

    const hash = parseHash();
    const access = hash.get("access_token");
    const refresh = hash.get("refresh_token");
    const username = hash.get("username");

    if (access && refresh && username) {
      acceptTokens({
        access_token: access,
        refresh_token: refresh,
        username,
      });
      window.history.replaceState(null, "", "/auth/callback");
      router.replace("/home");
      return;
    }

    toastError("Missing tokens in callback. Try logging in again.");
    setFailed(true);
  }, [acceptTokens, router]);

  return (
    <main className="auth-box stack">
      <h1>Signing you in…</h1>
      {failed ? (
        <Link href="/login">Back to login</Link>
      ) : (
        <p className="muted">Finishing Google sign-in.</p>
      )}
    </main>
  );
}
