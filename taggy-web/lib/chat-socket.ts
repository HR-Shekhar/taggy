"use client";

import { useEffect, useRef } from "react";
import { getTokens } from "@/lib/api";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

export type ChatSocketMessage = {
  id: number;
  author_username: string;
  author_name?: string;
  content: string;
  edited_at?: string | null;
  created_at: string;
  reply_to_message_id?: number | null;
  reply_to_content?: string | null;
  reply_to_author_username?: string | null;
};

export type ChatSocketEvent =
  | { type: "message.created"; message: ChatSocketMessage }
  | { type: "message.updated"; message: ChatSocketMessage }
  | { type: "message.deleted"; message_id: number };

function toWsBase(httpBase: string) {
  if (httpBase.startsWith("https://")) return `wss://${httpBase.slice(8)}`;
  if (httpBase.startsWith("http://")) return `ws://${httpBase.slice(7)}`;
  return httpBase;
}

export function chatRoomKeyPod(podSlug: string) {
  return `pod:${podSlug}`;
}

export function chatRoomKeyChannel(skillSlug: string, channelSlug: string) {
  return `channel:${skillSlug}:${channelSlug}`;
}

export function useChatSocket(
  room: string | null,
  onEvent: (event: ChatSocketEvent) => void
) {
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  useEffect(() => {
    if (!room) return;

    const { access } = getTokens();
    if (!access) return;

    let closed = false;
    let socket: WebSocket | null = null;
    let retryTimer: number | null = null;
    let attempt = 0;

    const connect = () => {
      if (closed) return;
      const url = `${toWsBase(API_BASE)}/ws/chat?room=${encodeURIComponent(room)}&token=${encodeURIComponent(access)}`;
      socket = new WebSocket(url);

      socket.onopen = () => {
        attempt = 0;
      };

      socket.onmessage = (ev) => {
        try {
          const data = JSON.parse(String(ev.data)) as ChatSocketEvent;
          if (!data?.type) return;
          onEventRef.current(data);
        } catch {
          // ignore malformed frames
        }
      };

      socket.onclose = () => {
        if (closed) return;
        const delay = Math.min(10_000, 800 * 2 ** attempt);
        attempt += 1;
        retryTimer = window.setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      closed = true;
      if (retryTimer != null) window.clearTimeout(retryTimer);
      socket?.close();
    };
  }, [room]);
}
