"use client";

import Link from "next/link";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { usePathname } from "next/navigation";
import { MessageCircle, X } from "lucide-react";
import { getTokens } from "@/lib/api";
import {
  type ChatSocketEvent,
  chatRoomKeyChannel,
  chatRoomKeyPod,
} from "@/lib/chat-socket";
import { Button } from "@/components/ui/button";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

function toWsBase(httpBase: string) {
  if (httpBase.startsWith("https://")) return `wss://${httpBase.slice(8)}`;
  if (httpBase.startsWith("http://")) return `ws://${httpBase.slice(7)}`;
  return httpBase;
}

type Listener = (event: ChatSocketEvent) => void;

type ChatConnectionContextValue = {
  activeRoom: string | null;
  connected: boolean;
  setActiveRoom: (room: string | null) => void;
  clearRoom: () => void;
  subscribe: (room: string | null, listener: Listener) => () => void;
};

const ChatConnectionContext = createContext<ChatConnectionContextValue | null>(
  null
);

export function useChatConnection() {
  const ctx = useContext(ChatConnectionContext);
  if (!ctx) {
    throw new Error(
      "useChatConnection must be used within ChatConnectionProvider"
    );
  }
  return ctx;
}

function roomLabel(room: string) {
  if (room.startsWith("pod:")) return `Pod · ${room.slice(4)}`;
  if (room.startsWith("channel:")) {
    const parts = room.split(":");
    return `Community · ${parts[1] ?? "chat"}`;
  }
  return room;
}

function roomHref(room: string) {
  if (room.startsWith("pod:")) return `/pods/${room.slice(4)}`;
  if (room.startsWith("channel:")) {
    const parts = room.split(":");
    return `/community/${parts[1] ?? ""}`;
  }
  return "/home";
}

export function ChatConnectionProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const [activeRoom, setActiveRoomState] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);
  const listenersRef = useRef(new Map<string, Set<Listener>>());
  const socketRef = useRef<WebSocket | null>(null);
  const retryRef = useRef<number | null>(null);
  const attemptRef = useRef(0);
  const roomRef = useRef<string | null>(null);

  const broadcast = useCallback((room: string, event: ChatSocketEvent) => {
    const set = listenersRef.current.get(room);
    if (!set) return;
    for (const fn of set) fn(event);
  }, []);

  const disconnectSocket = useCallback(() => {
    if (retryRef.current != null) {
      window.clearTimeout(retryRef.current);
      retryRef.current = null;
    }
    socketRef.current?.close();
    socketRef.current = null;
    setConnected(false);
  }, []);

  const connectSocket = useCallback(
    (room: string) => {
      const { access } = getTokens();
      if (!access) return;

      disconnectSocket();
      const url = `${toWsBase(API_BASE)}/ws/chat?room=${encodeURIComponent(room)}&token=${encodeURIComponent(access)}`;
      const socket = new WebSocket(url);
      socketRef.current = socket;

      socket.onopen = () => {
        attemptRef.current = 0;
        setConnected(true);
      };

      socket.onmessage = (ev) => {
        try {
          const data = JSON.parse(String(ev.data)) as ChatSocketEvent;
          if (!data?.type) return;
          broadcast(room, data);
        } catch {
          // ignore malformed frames
        }
      };

      socket.onclose = () => {
        setConnected(false);
        if (roomRef.current !== room) return;
        const delay = Math.min(10_000, 800 * 2 ** attemptRef.current);
        attemptRef.current += 1;
        retryRef.current = window.setTimeout(() => {
          if (roomRef.current === room) connectSocket(room);
        }, delay);
      };
    },
    [broadcast, disconnectSocket]
  );

  const setActiveRoom = useCallback(
    (room: string | null) => {
      if (room === roomRef.current) return;
      roomRef.current = room;
      setActiveRoomState(room);
      if (!room) {
        disconnectSocket();
        return;
      }
      connectSocket(room);
    },
    [connectSocket, disconnectSocket]
  );

  const clearRoom = useCallback(() => {
    setActiveRoom(null);
  }, [setActiveRoom]);

  const subscribe = useCallback((room: string | null, listener: Listener) => {
    if (!room) return () => {};
    let set = listenersRef.current.get(room);
    if (!set) {
      set = new Set();
      listenersRef.current.set(room, set);
    }
    set.add(listener);
    return () => {
      set?.delete(listener);
      if (set && set.size === 0) listenersRef.current.delete(room);
    };
  }, []);

  useEffect(() => {
    return () => disconnectSocket();
  }, [disconnectSocket]);

  const value = useMemo(
    () => ({
      activeRoom,
      connected,
      setActiveRoom,
      clearRoom,
      subscribe,
    }),
    [activeRoom, connected, setActiveRoom, clearRoom, subscribe]
  );

  const onChatSurface =
    activeRoom != null &&
    (pathname.startsWith("/pods/") || pathname.startsWith("/community/"));
  const showChip = Boolean(activeRoom && connected && !onChatSurface);

  return (
    <ChatConnectionContext.Provider value={value}>
      {children}
      {showChip && activeRoom ? (
        <div className="fixed bottom-4 left-4 z-50 flex max-w-xs items-center gap-2 rounded-full border border-border bg-card/95 px-3 py-2 shadow-lg backdrop-blur-md">
          <MessageCircle className="size-4 shrink-0 text-primary" />
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs font-medium">
              Listening in {roomLabel(activeRoom)}
            </p>
          </div>
          <Link
            href={roomHref(activeRoom)}
            className="text-xs font-medium text-primary hover:underline"
          >
            Open
          </Link>
          <Button
            type="button"
            size="icon"
            variant="ghost"
            className="size-7"
            aria-label="Leave chat room"
            onClick={clearRoom}
          >
            <X className="size-3.5" />
          </Button>
        </div>
      ) : null}
    </ChatConnectionContext.Provider>
  );
}

/** Page hook: keep shared socket on this room; only unsubscribe on leave. */
export function usePersistentChatSocket(
  room: string | null,
  onEvent: (event: ChatSocketEvent) => void
) {
  const { setActiveRoom, subscribe } = useChatConnection();
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  useEffect(() => {
    if (room) setActiveRoom(room);
  }, [room, setActiveRoom]);

  useEffect(() => {
    return subscribe(room, (event) => onEventRef.current(event));
  }, [room, subscribe]);
}

export { chatRoomKeyPod, chatRoomKeyChannel };
