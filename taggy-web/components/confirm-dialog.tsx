"use client";

import { AlertDialog as AlertDialogPrimitive } from "@base-ui/react/alert-dialog";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  destructive = false,
  busy = false,
  onConfirm,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
  busy?: boolean;
  onConfirm: () => void | Promise<void>;
}) {
  return (
    <AlertDialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
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
          <div className="mt-5 flex justify-end gap-2">
            <AlertDialogPrimitive.Close
              render={<Button variant="outline" disabled={busy} />}
            >
              {cancelLabel}
            </AlertDialogPrimitive.Close>
            <Button
              variant={destructive ? "destructive" : "default"}
              disabled={busy}
              onClick={async () => {
                await onConfirm();
              }}
            >
              {busy ? "Please wait…" : confirmLabel}
            </Button>
          </div>
        </AlertDialogPrimitive.Popup>
      </AlertDialogPrimitive.Portal>
    </AlertDialogPrimitive.Root>
  );
}
