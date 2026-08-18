"use client";

import { FormEvent, ReactNode, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CheckCircle2, Loader2, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { toastError } from "@/lib/toast";

export function ErrorBox({
  message,
  title,
}: {
  message: string | null;
  title?: string;
}) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!message) {
      setVisible(false);
      return;
    }
    const id = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(id);
  }, [message]);

  if (!message) return null;

  return (
    <div
      role="alert"
      className={cn(
        "rounded-lg border border-border/80 bg-muted/50 px-3 py-2.5 text-sm transition-all duration-300 ease-out",
        visible ? "translate-y-0 opacity-100" : "-translate-y-1 opacity-0"
      )}
    >
      {title ? (
        <>
          <p className="font-medium text-foreground">{title}</p>
          <p className="mt-0.5 text-muted-foreground">{message}</p>
        </>
      ) : (
        <p className="font-medium text-foreground">{message}</p>
      )}
    </div>
  );
}

export function SuccessDialog({
  open,
  title = "Done",
  description,
  confirmLabel = "Got it",
  onClose,
  onConfirm,
}: {
  open: boolean;
  title?: string;
  description: string;
  confirmLabel?: string;
  onClose: () => void;
  onConfirm?: () => void;
}) {
  if (!open) return null;

  const confirm = onConfirm ?? onClose;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        aria-label="Dismiss"
        className="absolute inset-0 bg-background/70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />
      <Card
        role="dialog"
        aria-modal="true"
        className="relative z-10 w-full max-w-md rounded-xl border border-border/70 bg-card/95 shadow-lg duration-200 animate-in fade-in zoom-in-95"
      >
        <CardHeader className="space-y-3">
          <div className="flex items-start justify-between gap-3">
            <span className="inline-flex size-10 items-center justify-center rounded-full bg-primary/15 text-primary">
              <CheckCircle2 className="size-5" />
            </span>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 shrink-0"
              onClick={onClose}
            >
              <X className="size-4" />
            </Button>
          </div>
          <div className="space-y-1">
            <CardTitle className="font-serif text-xl">{title}</CardTitle>
            <CardDescription className="text-sm leading-relaxed">
              {description}
            </CardDescription>
          </div>
        </CardHeader>
        <CardFooter className="justify-end border-0 bg-transparent">
          <Button type="button" onClick={confirm}>
            {confirmLabel}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

export function GenerationWaitNote({
  active,
  label = "AI is drafting the roadmap. This can take up to a few minutes depending on size — keep this tab open.",
}: {
  active: boolean;
  label?: string;
}) {
  if (!active) return null;
  return (
    <div className="flex items-start gap-2 rounded-lg border border-border/70 bg-muted/40 px-3 py-2.5 text-sm text-muted-foreground transition-all duration-300">
      <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin text-primary" />
      <p>{label}</p>
    </div>
  );
}

export function Loading({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="flex items-center gap-2 text-muted-foreground">
      <Loader2 className="size-4 animate-spin" />
      <span>{label}</span>
    </div>
  );
}

export function Empty({ children }: { children: ReactNode }) {
  return (
    <p className="rounded-lg border border-dashed border-border bg-muted/35 px-4 py-8 text-center text-sm text-muted-foreground backdrop-blur-sm">
      {children}
    </p>
  );
}

export function useAsyncAction() {
  const [busy, setBusy] = useState(false);

  async function run(fn: () => Promise<void>) {
    setBusy(true);
    try {
      await fn();
    } catch (e) {
      toastError(e instanceof Error ? e.message : "Something went wrong");
    } finally {
      setBusy(false);
    }
  }

  return { busy, run };
}

export function FormCard({
  title,
  description,
  children,
  onSubmit,
  submitLabel,
  busy,
  extra,
}: {
  title?: string;
  description?: string;
  children: ReactNode;
  onSubmit: (e: FormEvent) => void | Promise<void>;
  submitLabel: string;
  busy?: boolean;
  extra?: ReactNode;
}) {
  return (
    <Card className="rounded-xl border border-border/70 bg-card/60 shadow-sm ring-0 backdrop-blur-md">
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await onSubmit(e);
        }}
      >
        {(title || description) && (
          <CardHeader>
            {title && <CardTitle className="font-serif">{title}</CardTitle>}
            {description && <CardDescription>{description}</CardDescription>}
          </CardHeader>
        )}
        <CardContent className="mx-auto w-full max-w-[14rem] space-y-4 pt-3">
          {children}
        </CardContent>
        <CardFooter className="mx-auto w-full max-w-[14rem] flex-col gap-4 border-0 bg-transparent">
          <Button
            type="submit"
            disabled={busy}
            className="w-full rounded-full"
            size="lg"
          >
            {busy && <Loader2 className="size-4 animate-spin" />}
            {busy ? "Please wait…" : submitLabel}
          </Button>
        </CardFooter>
      </form>
      {extra ? (
        <div className="mx-auto w-full max-w-[14rem] px-0 pb-(--card-spacing)">
          {extra}
        </div>
      ) : null}
    </Card>
  );
}

export function PageHeader({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children?: ReactNode;
}) {
  return (
    <div className="mb-1 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="space-y-1">
        <h1 className="font-serif text-2xl font-medium tracking-tight sm:text-3xl">
          {title}
        </h1>
        {description && (
          <p className="text-muted-foreground">{description}</p>
        )}
      </div>
      {children}
    </div>
  );
}
