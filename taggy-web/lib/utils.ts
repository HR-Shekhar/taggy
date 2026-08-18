import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** First word of a full name; falls back to username or empty string. */
export function displayFirstName(
  name?: string | null,
  username?: string | null
): string {
  const first = (name ?? "").trim().split(/\s+/).filter(Boolean)[0];
  if (first) return first;
  return (username ?? "").trim();
}
