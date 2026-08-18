"use client";

import { startGoogleOAuth } from "@/lib/api";

export function GoogleIconButton({ disabled }: { disabled?: boolean }) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={() => startGoogleOAuth()}
      className="flex size-11 items-center justify-center rounded-full border border-primary/40 text-primary transition-colors hover:bg-primary/10 disabled:opacity-50"
      aria-label="Continue with Google"
    >
      <GoogleMark />
    </button>
  );
}

function GoogleMark() {
  return (
    <svg viewBox="0 0 24 24" className="size-5" aria-hidden>
      <path
        fill="currentColor"
        d="M21.6 12.23c0-.74-.07-1.45-.19-2.13H12v4.03h5.38a4.6 4.6 0 0 1-2 3.02v2.5h3.23c1.89-1.74 2.99-4.31 2.99-7.42Z"
      />
      <path
        fill="currentColor"
        d="M12 22c2.7 0 4.97-.9 6.63-2.35l-3.23-2.5c-.9.6-2.05.96-3.4.96-2.61 0-4.82-1.76-5.61-4.13H3.06v2.58A10 10 0 0 0 12 22Z"
      />
      <path
        fill="currentColor"
        d="M6.39 13.98A6.01 6.01 0 0 1 6.07 12c0-.69.12-1.35.32-1.98V7.44H3.06A10 10 0 0 0 2 12c0 1.61.39 3.14 1.06 4.56l3.33-2.58Z"
      />
      <path
        fill="currentColor"
        d="M12 5.88c1.47 0 2.79.51 3.83 1.5l2.87-2.87C16.96 2.87 14.7 2 12 2A10 10 0 0 0 3.06 7.44l3.33 2.58C7.18 7.64 9.39 5.88 12 5.88Z"
      />
    </svg>
  );
}
