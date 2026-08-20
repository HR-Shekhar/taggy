"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { FormEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { GoogleIconButton } from "@/components/google-button";
import { toastApiError, toastError } from "@/lib/toast";
import { UnderlineField } from "@/components/underline-field";
import { useAuth } from "@/lib/auth";
import { register } from "@/lib/api";
import { BrandLogo } from "@/components/brand-logo";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

export function AuthPair() {
  const pathname = usePathname();
  const isRegister = pathname === "/register";

  return (
    <AuthChrome>
      <div className="relative w-full max-w-3xl overflow-hidden bg-card shadow-2xl ring-1 ring-border lg:h-[min(27rem,calc(100dvh-5.5rem))]">
        {/* Login stays on the left */}
        <div
          className={cn(
            "bg-card lg:absolute lg:inset-y-0 lg:left-0 lg:w-1/2",
            isRegister ? "hidden lg:block" : "block"
          )}
          aria-hidden={isRegister}
        >
          <div className={cn(isRegister && "lg:pointer-events-none")}>
            <LoginPanel />
          </div>
        </div>

        {/* Signup stays on the right */}
        <div
          className={cn(
            "bg-card lg:absolute lg:inset-y-0 lg:left-1/2 lg:w-1/2",
            isRegister ? "block" : "hidden lg:block"
          )}
          aria-hidden={!isRegister}
        >
          <div className={cn(!isRegister && "lg:pointer-events-none")}>
            <RegisterPanel />
          </div>
        </div>

        {/* Photo slides as a cover — forms do not move */}
        <div
          className={cn(
            "absolute inset-y-0 left-0 z-10 hidden w-1/2 transform-gpu transition-transform duration-700 ease-in-out lg:block",
            isRegister ? "translate-x-0" : "translate-x-full"
          )}
        >
          <Image
            src="/images/authpage.jpg"
            alt=""
            fill
            priority
            sizes="50vw"
            className="object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-black/10" />

          {isRegister ? (
            <Link
              href="/login"
              className="absolute top-8 right-0 z-20 flex translate-x-1/2 items-center gap-1.5 bg-primary px-5 py-2.5 text-[11px] font-medium uppercase tracking-[0.2em] text-primary-foreground shadow-lg hover:opacity-90"
            >
              <span aria-hidden className="text-sm leading-none">
                ←
              </span>
              Login
            </Link>
          ) : (
            <Link
              href="/register"
              className="absolute top-8 left-0 z-20 flex -translate-x-1/2 items-center gap-1.5 bg-primary px-5 py-2.5 text-[11px] font-medium uppercase tracking-[0.2em] text-primary-foreground shadow-lg hover:opacity-90"
            >
              Sign up
              <span aria-hidden className="text-sm leading-none">
                →
              </span>
            </Link>
          )}
        </div>
      </div>
    </AuthChrome>
  );
}

function AuthChrome({ children }: { children: ReactNode }) {
  return (
    <div className="relative flex min-h-dvh flex-col overflow-x-clip bg-transparent lg:h-dvh lg:overflow-hidden">
      <header className="relative z-20 flex shrink-0 items-center justify-between px-5 py-3 sm:px-8">
        <Link href="/" className="text-lg text-foreground">
          <BrandLogo size={32} wordmarkClassName="text-lg" />
        </Link>
        <ThemeToggle />
      </header>

      <main className="relative z-10 flex min-h-0 flex-1 items-center justify-center px-4 pb-12 pt-4 sm:px-6 lg:overflow-hidden lg:pb-16">
        {children}
      </main>
    </div>
  );
}

function LoginPanel() {
  const { login, isAuthenticated } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      router.replace("/home");
    }
  }, [isAuthenticated, router]);

  return (
    <AuthForm
      title="Login"
      submitLabel="Log in"
      busy={busy}
      extra={
        <div className="flex justify-center">
          <GoogleIconButton disabled={busy} />
        </div>
      }
      onSubmit={async () => {
        setBusy(true);
        const err = await login(email, password);
        setBusy(false);
        if (err) {
          toastError(err);
          return;
        }
        router.replace("/home");
      }}
    >
      <UnderlineField
        id="login-email"
        label="Email"
        type="email"
        placeholder="Enter your email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
        autoComplete="email"
      />
      <UnderlineField
        id="login-password"
        label="Password"
        type="password"
        placeholder="Enter your password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
        autoComplete="current-password"
      />
      <p className="pt-1 text-center text-sm text-muted-foreground lg:hidden">
        New here?{" "}
        <Link href="/register" className="text-foreground underline-offset-4 hover:underline">
          Sign up
        </Link>
      </p>
    </AuthForm>
  );
}

function RegisterPanel() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const checks = useMemo(() => passwordChecks(password), [password]);

  return (
    <AuthForm
      title="Sign up"
      titleAlign="end"
      submitLabel="Create account"
      busy={busy}
      extra={
        <div className="flex justify-center">
          <GoogleIconButton disabled={busy} />
        </div>
      }
      onSubmit={async () => {
        const hint = passwordHint(password);
        if (hint) {
          toastError(hint);
          return;
        }
        setBusy(true);
        const result = await register({
          email,
          username,
          password,
          name: name || undefined,
        });
        setBusy(false);
        if (!result.ok) {
          toastApiError(result);
          return;
        }
        const q = new URLSearchParams({ email });
        if (result.data?.dev_otp) q.set("otp", result.data.dev_otp);
        router.push(`/verify?${q.toString()}`);
      }}
    >
      <UnderlineField
        id="register-email"
        label="Email"
        type="email"
        placeholder="Enter your email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        required
        autoComplete="email"
      />
      <div className="grid grid-cols-2 gap-4">
        <UnderlineField
          id="register-username"
          label="Username"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          required
          minLength={3}
          maxLength={30}
          autoComplete="username"
        />
        <UnderlineField
          id="register-name"
          label="Name"
          placeholder="Optional"
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoComplete="name"
        />
      </div>
      <div className="space-y-2">
        <UnderlineField
          id="register-password"
          label="Password"
          type="password"
          placeholder="Create a password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          autoComplete="new-password"
        />
        <ul className="flex flex-wrap gap-x-3 gap-y-0.5 text-[10px] text-muted-foreground">
          <Rule ok={checks.length} label="8+" />
          <Rule ok={checks.upper} label="A-Z" />
          <Rule ok={checks.lower} label="a-z" />
          <Rule ok={checks.number} label="0-9" />
          <Rule ok={checks.special} label="symbol" />
        </ul>
      </div>
      <p className="pt-1 text-center text-sm text-muted-foreground lg:hidden">
        Already have an account?{" "}
        <Link href="/login" className="text-foreground underline-offset-4 hover:underline">
          Login
        </Link>
      </p>
    </AuthForm>
  );
}

function passwordChecks(password: string) {
  return {
    length: password.length >= 8,
    upper: /[A-Z]/.test(password),
    lower: /[a-z]/.test(password),
    number: /[0-9]/.test(password),
    special: /[^A-Za-z0-9]/.test(password),
  };
}

function passwordHint(password: string) {
  const checks = passwordChecks(password);
  if (!checks.length) return "Use at least 8 characters.";
  if (!checks.upper) return "Add at least one capital letter.";
  if (!checks.lower) return "Add at least one lowercase letter.";
  if (!checks.number) return "Add at least one number.";
  if (!checks.special) return "Add at least one special character.";
  return null;
}

function Rule({ ok, label }: { ok: boolean; label: string }) {
  return (
    <li className={cn(ok ? "text-primary" : "text-muted-foreground")}>
      {ok ? "✓" : "○"} {label}
    </li>
  );
}

export function AuthForm({
  title,
  titleAlign = "start",
  subtitle,
  children,
  submitLabel,
  busy,
  onSubmit,
  footer,
  extra,
}: {
  title: string;
  titleAlign?: "start" | "end";
  subtitle?: ReactNode;
  children: ReactNode;
  submitLabel: string;
  busy?: boolean;
  onSubmit: (e: FormEvent) => void | Promise<void>;
  footer?: ReactNode;
  extra?: ReactNode;
}) {
  return (
    <form
      className="flex h-full flex-col justify-center overflow-y-auto px-6 py-5 sm:px-8 lg:py-5"
      onSubmit={async (e) => {
        e.preventDefault();
        await onSubmit(e);
      }}
    >
      <div
        className={cn(
          "mb-3.5",
          titleAlign === "end" && "lg:pr-2 lg:text-right"
        )}
      >
        <h1 className="font-serif text-2xl tracking-tight text-foreground sm:text-3xl">
          {title}
        </h1>
        {subtitle && <p className="mt-1 text-sm text-muted-foreground">{subtitle}</p>}
      </div>

      <div className="space-y-3">{children}</div>

      <div className="mt-4 flex shrink-0 flex-col items-center gap-3 pb-1">
        <Button
          type="submit"
          disabled={busy}
          className="h-10 min-w-40"
        >
          {busy && <Loader2 className="size-4 animate-spin" />}
          {busy ? "Please wait…" : submitLabel}
        </Button>

        {extra && (
          <div className="w-full max-w-xs space-y-2.5">
            <div className="flex items-center gap-3">
              <div className="h-px flex-1 bg-border" />
              <span className="text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
                or
              </span>
              <div className="h-px flex-1 bg-border" />
            </div>
            {extra}
          </div>
        )}
      </div>

      {footer && (
        <p className="mt-3 text-center text-sm text-muted-foreground">{footer}</p>
      )}
    </form>
  );
}

export function AuthStage({ children }: { children: ReactNode }) {
  return <AuthChrome>{children}</AuthChrome>;
}

/** Standalone auth screens (verify, complete google). */
export function AuthShell({
  title,
  subtitle,
  children,
  submitLabel,
  busy,
  onSubmit,
  footer,
  extra,
}: {
  mode?: string;
  title: string;
  subtitle?: ReactNode;
  children: ReactNode;
  submitLabel: string;
  busy?: boolean;
  onSubmit: (e: FormEvent) => void | Promise<void>;
  footer?: ReactNode;
  extra?: ReactNode;
}) {
  return (
    <AuthChrome>
      <div className="relative w-full max-w-3xl overflow-hidden shadow-2xl lg:h-[min(27rem,calc(100dvh-5.5rem))]">
        <div className="bg-card lg:absolute lg:inset-y-0 lg:left-0 lg:w-1/2">
          <AuthForm
            title={title}
            subtitle={subtitle}
            submitLabel={submitLabel}
            busy={busy}
            onSubmit={onSubmit}
            footer={footer}
            extra={extra}
          >
            {children}
          </AuthForm>
        </div>
        <div className="relative hidden min-h-[18rem] lg:absolute lg:inset-y-0 lg:left-1/2 lg:block lg:w-1/2 lg:min-h-0">
          <Image
            src="/images/authpage.jpg"
            alt=""
            fill
            priority
            sizes="50vw"
            className="object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-black/10" />
        </div>
      </div>
    </AuthChrome>
  );
}
