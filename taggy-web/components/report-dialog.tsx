"use client";

import { useState } from "react";
import { AlertDialog as AlertDialogPrimitive } from "@base-ui/react/alert-dialog";
import { createReport } from "@/lib/api";
import { toastApiError, toastSuccess } from "@/lib/toast";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

export type ReportTargetType =
  | "USER"
  | "PROPOSAL"
  | "POD"
  | "MESSAGE"
  | "AUDIO_ROOM"
  | "COMMUNITY_CHANNEL";

export function ReportDialog({
  open,
  onOpenChange,
  targetType,
  targetId,
  title = "Submit a report",
  description = "Tell us what’s wrong. Reports are reviewed by moderators.",
  onSubmitted,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetType: ReportTargetType;
  targetId: number | null;
  title?: string;
  description?: string;
  onSubmitted?: () => void;
}) {
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit() {
    if (!targetId || reason.trim().length < 3) return;
    setBusy(true);
    const result = await createReport({
      target_type: targetType,
      target_id: targetId,
      reason: reason.trim(),
    });
    setBusy(false);
    if (!result.ok) {
      toastApiError(result);
      return;
    }
    setReason("");
    onOpenChange(false);
    toastSuccess("Report submitted.");
    onSubmitted?.();
  }

  return (
    <AlertDialogPrimitive.Root
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          setReason("");
        }
        onOpenChange(next);
      }}
    >
      <AlertDialogPrimitive.Portal>
        <AlertDialogPrimitive.Backdrop
          className={cn(
            "fixed inset-0 z-50 bg-black/40 backdrop-blur-sm",
            "transition-opacity duration-200 data-ending-style:opacity-0 data-starting-style:opacity-0"
          )}
        />
        <AlertDialogPrimitive.Popup
          className={cn(
            "fixed top-1/2 left-1/2 z-50 w-[min(24rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2",
            "rounded-xl border border-border/70 bg-card/95 p-5 shadow-xl backdrop-blur-md outline-none",
            "transition-all duration-200 data-ending-style:scale-95 data-ending-style:opacity-0 data-starting-style:scale-95 data-starting-style:opacity-0"
          )}
        >
          <AlertDialogPrimitive.Title className="font-serif text-lg font-medium">
            {title}
          </AlertDialogPrimitive.Title>
          <AlertDialogPrimitive.Description className="mt-2 text-sm text-muted-foreground">
            {description}
          </AlertDialogPrimitive.Description>

          <div className="mt-4 space-y-2">
            <Label htmlFor="report-reason">Reason</Label>
            <Textarea
              id="report-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              required
              minLength={3}
              maxLength={2000}
              placeholder="Describe the issue (min 3 characters)"
              disabled={busy || !targetId}
            />
            {!targetId ? (
              <p className="text-sm text-destructive">
                Missing target id — cannot submit this report.
              </p>
            ) : null}
          </div>

          <div className="mt-5 flex justify-end gap-2">
            <AlertDialogPrimitive.Close
              render={<Button variant="outline" disabled={busy} />}
            >
              Cancel
            </AlertDialogPrimitive.Close>
            <Button
              variant="destructive"
              disabled={busy || !targetId || reason.trim().length < 3}
              onClick={() => void submit()}
            >
              {busy ? "Submitting…" : "Submit report"}
            </Button>
          </div>
        </AlertDialogPrimitive.Popup>
      </AlertDialogPrimitive.Portal>
    </AlertDialogPrimitive.Root>
  );
}
