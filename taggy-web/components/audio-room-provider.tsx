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
  type RefObject,
} from "react";
import { usePathname } from "next/navigation";
import { Room, RoomEvent, Track } from "livekit-client";
import { Mic, MicOff, PhoneOff, Radio } from "lucide-react";
import {
  apiErrorMessage,
  endAudioRoom,
  getAudioRoom,
  joinAudioRoom,
  leaveAudioRoom,
} from "@/lib/api";
import { toastError } from "@/lib/toast";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type RoomInfo = {
  id: string;
  entity_id?: number;
  title: string;
  status: string;
  host_username?: string;
  pod_slug?: string | null;
  livekit_room_name?: string;
  max_participants?: number;
};

type Participant = {
  username: string;
  name?: string;
  role?: string;
};

type AudioRoomContextValue = {
  roomId: string | null;
  room: RoomInfo | null;
  participants: Participant[];
  connected: boolean;
  micOn: boolean;
  busy: boolean;
  loadError: string | null;
  loading: boolean;
  audioElRef: RefObject<HTMLDivElement | null>;
  bindRoom: (roomId: string) => Promise<void>;
  connect: (roomId?: string) => Promise<void>;
  leave: () => Promise<void>;
  endRoom: () => Promise<void>;
  toggleMic: () => Promise<void>;
  refresh: () => Promise<void>;
};

const AudioRoomContext = createContext<AudioRoomContextValue | null>(null);

export function useAudioRoom() {
  const ctx = useContext(AudioRoomContext);
  if (!ctx) {
    throw new Error("useAudioRoom must be used within AudioRoomProvider");
  }
  return ctx;
}

export function AudioRoomProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const [roomId, setRoomId] = useState<string | null>(null);
  const [room, setRoom] = useState<RoomInfo | null>(null);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [connected, setConnected] = useState(false);
  const [micOn, setMicOn] = useState(true);
  const [busy, setBusy] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const liveRef = useRef<Room | null>(null);
  const audioElRef = useRef<HTMLDivElement | null>(null);
  const roomIdRef = useRef<string | null>(null);
  roomIdRef.current = roomId;

  const refresh = useCallback(async () => {
    const id = roomIdRef.current;
    if (!id) {
      setRoom(null);
      setParticipants([]);
      return;
    }
    const result = await getAudioRoom(id);
    if (!result.ok) {
      const message = apiErrorMessage(result);
      setLoadError(message);
      setRoom(null);
      setParticipants([]);
      return;
    }
    setLoadError(null);
    if (!result.data?.room) {
      setRoom(null);
      setParticipants([]);
      return;
    }
    setRoom(result.data.room);
    setParticipants(result.data.participants ?? []);
  }, []);

  const bindRoom = useCallback(
    async (id: string) => {
      setLoading(true);
      setRoomId(id);
      roomIdRef.current = id;
      const result = await getAudioRoom(id);
      if (!result.ok) {
        setLoadError(apiErrorMessage(result));
        setRoom(null);
        setParticipants([]);
        setLoading(false);
        return;
      }
      setLoadError(null);
      setRoom(result.data?.room ?? null);
      setParticipants(result.data?.participants ?? []);
      setLoading(false);
    },
    []
  );

  const connect = useCallback(
    async (explicitId?: string) => {
      const id = explicitId ?? roomIdRef.current;
      if (!id) return;
      if (explicitId && explicitId !== roomIdRef.current) {
        await bindRoom(explicitId);
      }
      setBusy(true);
      const result = await joinAudioRoom(id);
      if (!result.ok || !result.data?.token || !result.data.livekit_url) {
        setBusy(false);
        toastError(apiErrorMessage(result));
        return;
      }
      try {
        if (
          liveRef.current &&
          roomIdRef.current === id &&
          liveRef.current.state === "connected"
        ) {
          setBusy(false);
          return;
        }
        await liveRef.current?.disconnect();
        if (audioElRef.current) audioElRef.current.innerHTML = "";
        const live = new Room();
        liveRef.current = live;
        live.on(RoomEvent.TrackSubscribed, (track) => {
          if (track.kind === Track.Kind.Audio && audioElRef.current) {
            const el = track.attach();
            audioElRef.current.appendChild(el);
          }
        });
        live.on(RoomEvent.Disconnected, () => {
          setConnected(false);
        });
        await live.connect(result.data.livekit_url, result.data.token);
        await live.localParticipant.setMicrophoneEnabled(true);
        setMicOn(true);
        setConnected(true);
        setRoomId(id);
        roomIdRef.current = id;
        await refresh();
      } catch (e) {
        toastError(
          e instanceof Error ? e.message : "Failed to connect to LiveKit"
        );
        setConnected(false);
      } finally {
        setBusy(false);
      }
    },
    [bindRoom, refresh]
  );

  const leave = useCallback(async () => {
    const id = roomIdRef.current;
    setBusy(true);
    await liveRef.current?.disconnect();
    liveRef.current = null;
    setConnected(false);
    if (id) {
      const result = await leaveAudioRoom(id);
      if (!result.ok) toastError(apiErrorMessage(result));
    }
    setRoomId(null);
    roomIdRef.current = null;
    setRoom(null);
    setParticipants([]);
    setBusy(false);
  }, []);

  const endRoom = useCallback(async () => {
    const id = roomIdRef.current;
    if (!id) return;
    setBusy(true);
    const result = await endAudioRoom(id);
    if (!result.ok) {
      setBusy(false);
      toastError(apiErrorMessage(result));
      return;
    }
    await liveRef.current?.disconnect();
    liveRef.current = null;
    setConnected(false);
    setRoomId(null);
    roomIdRef.current = null;
    setRoom(null);
    setParticipants([]);
    setBusy(false);
  }, []);

  const toggleMic = useCallback(async () => {
    const live = liveRef.current;
    if (!live) return;
    const next = !micOn;
    await live.localParticipant.setMicrophoneEnabled(next);
    setMicOn(next);
  }, [micOn]);

  // Do not disconnect on route change — only explicit leave/end.
  useEffect(() => {
    return () => {
      void liveRef.current?.disconnect();
      liveRef.current = null;
    };
  }, []);

  const value = useMemo<AudioRoomContextValue>(
    () => ({
      roomId,
      room,
      participants,
      connected,
      micOn,
      busy,
      loadError,
      loading,
      audioElRef,
      bindRoom,
      connect,
      leave,
      endRoom,
      toggleMic,
      refresh,
    }),
    [
      roomId,
      room,
      participants,
      connected,
      micOn,
      busy,
      loadError,
      loading,
      bindRoom,
      connect,
      leave,
      endRoom,
      toggleMic,
      refresh,
    ]
  );

  const onAudioPage =
    roomId != null && pathname.startsWith(`/audio-rooms/${roomId}`);
  const showMini = connected && room && !onAudioPage;

  return (
    <AudioRoomContext.Provider value={value}>
      {children}
      {/* Persistent remote-audio sink (hidden when mini/session UIs remount). */}
      <div ref={audioElRef} className="sr-only" aria-hidden />
      {showMini ? (
        <div
          className={cn(
            "fixed bottom-4 right-4 z-50 flex max-w-sm items-center gap-2 rounded-full border border-border bg-card/95 px-3 py-2 shadow-lg backdrop-blur-md"
          )}
        >
          <Radio className="size-4 shrink-0 text-primary" />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">{room.title}</p>
            <p className="truncate text-xs text-muted-foreground">
              Still in audio room
            </p>
          </div>
          <Button
            type="button"
            size="icon"
            variant="ghost"
            aria-label={micOn ? "Mute" : "Unmute"}
            disabled={busy}
            onClick={() => void toggleMic()}
          >
            {micOn ? <Mic className="size-4" /> : <MicOff className="size-4" />}
          </Button>
          <Link
            href={`/audio-rooms/${roomId}`}
            className="rounded-md px-2 py-1 text-xs font-medium text-primary hover:underline"
          >
            Return
          </Link>
          <Button
            type="button"
            size="icon"
            variant="ghost"
            aria-label="Leave room"
            disabled={busy}
            onClick={() => void leave()}
          >
            <PhoneOff className="size-4 text-destructive" />
          </Button>
        </div>
      ) : null}
    </AudioRoomContext.Provider>
  );
}
