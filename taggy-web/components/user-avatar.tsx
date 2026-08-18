"use client";

import { UserRound } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

export function UserAvatar({
  name,
  username,
  src,
  className,
  fallbackClassName,
}: {
  name?: string | null;
  username?: string | null;
  src?: string | null;
  className?: string;
  fallbackClassName?: string;
}) {
  const initials = avatarInitials(name, username);
  return (
    <Avatar className={cn("size-10 overflow-hidden", className)}>
      {src ? <AvatarImage src={src} alt={name || username || "Profile photo"} /> : null}
      <AvatarFallback
        className={cn(
          "bg-secondary text-sm font-medium text-secondary-foreground",
          fallbackClassName
        )}
      >
        {initials || <UserRound className="size-[45%] text-muted-foreground" />}
      </AvatarFallback>
    </Avatar>
  );
}

function avatarInitials(name?: string | null, username?: string | null) {
  const fromName = (name ?? "")
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("");
  if (fromName) return fromName.toUpperCase();
  const u = (username ?? "").trim();
  return u ? u.slice(0, 2).toUpperCase() : "";
}
