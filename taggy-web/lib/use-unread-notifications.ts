"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { listNotifications } from "@/lib/api";

export const NOTIFICATIONS_CHANGED_EVENT = "taggy-notifications-changed";

export function notifyNotificationsChanged() {
  if (typeof window === "undefined") return;
  window.dispatchEvent(new Event(NOTIFICATIONS_CHANGED_EVENT));
}

export function useUnreadNotifications() {
  const { username } = useAuth();
  const pathname = usePathname();
  const [count, setCount] = useState(0);

  const refresh = useCallback(async () => {
    if (!username) {
      setCount(0);
      return;
    }
    const result = await listNotifications(username, true);
    if (!result.ok) return;
    const data = result.data as {
      notifications?: unknown[];
      unread_count?: number;
    };
    if (typeof data?.unread_count === "number") {
      setCount(data.unread_count);
      return;
    }
    const list = Array.isArray(result.data)
      ? result.data
      : data?.notifications ?? [];
    setCount(Array.isArray(list) ? list.length : 0);
  }, [username]);

  useEffect(() => {
    void refresh();
    const onFocus = () => void refresh();
    const onChanged = () => void refresh();
    window.addEventListener("focus", onFocus);
    window.addEventListener(NOTIFICATIONS_CHANGED_EVENT, onChanged);
    const t = window.setInterval(() => void refresh(), 30_000);
    return () => {
      window.removeEventListener("focus", onFocus);
      window.removeEventListener(NOTIFICATIONS_CHANGED_EVENT, onChanged);
      window.clearInterval(t);
    };
  }, [refresh, pathname]);

  return { count, refresh };
}
