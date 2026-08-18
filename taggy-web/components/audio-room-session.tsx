"use client";

import { useEffect, useRef, useState } from "react";
import { Room, RoomEvent, Track } from "livekit-client";
import { Mic, MicOff, PhoneOff } from "lucide-react";
import {
  apiErrorMessage,
  endAudioRoom,
  getAudioRoom,
  joinAudioRoom,
  leaveAudioRoom,
} from "@/lib/api";
import { ErrorBox, Loading } from "@/components/app-ui";
import { ReportDialog } from "@/components/report-dialog";
import { toastError } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

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

export function AudioRoomSession({
  roomId,
  autoJoin = false,
}: {
  roomId: string;
  autoJoin?: boolean;
}) {
  const [room, setRoom] = useState<RoomInfo | null>(null);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [connected, setConnected] = useState(false);
  const [micOn, setMicOn] = useState(true);
  const [busy, setBusy] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);
  const liveRef = useRef<Room | null>(null);
  const audioRef = useRef<HTMLDivElement | null>(null);

  async function refresh() {
    const result = await getAudioRoom(roomId);
    if (!result.ok) {
      const message = apiErrorMessage(result);
      setLoadError(message);
      toastError(message);
      setRoom(null);
      return;
    }
    setLoadError(null);
    const data = result.data as {
      room?: RoomInfo;
      participants?: Participant[];
    };
    setRoom(data.room ?? (result.data as RoomInfo));
    setParticipants(data.participants ?? []);
  }

  useEffect(() => {
    (async () => {
      setLoading(true);
      await refresh();
      setLoading(false);
    })();
    return () => {
      void liveRef.current?.disconnect();
      liveRef.current = null;
    };
  }, [roomId]);

  useEffect(() => {
    if (autoJoin && room && room.status === "ACTIVE" && !connected) {
      void connect();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoJoin, room?.id, room?.status]);

  async function connect() {
    setBusy(true);
    const result = await joinAudioRoom(roomId);
    if (!result.ok || !result.data?.token || !result.data.livekit_url) {
      setBusy(false);
      toastError(apiErrorMessage(result));
      return;
    }
    try {
      await liveRef.current?.disconnect();
      if (audioRef.current) audioRef.current.innerHTML = "";
      const live = new Room();
      liveRef.current = live;
      live.on(RoomEvent.TrackSubscribed, (track) => {
        if (track.kind === Track.Kind.Audio && audioRef.current) {
          const el = track.attach();
          audioRef.current.appendChild(el);
        }
      });
      live.on(RoomEvent.Disconnected, () => {
        setConnected(false);
      });
      await live.connect(result.data.livekit_url, result.data.token);
      await live.localParticipant.setMicrophoneEnabled(true);
      setMicOn(true);
      setConnected(true);
      await refresh();
    } catch (e) {
      toastError(e instanceof Error ? e.message : "Failed to connect to LiveKit");
      setConnected(false);
    } finally {
      setBusy(false);
    }
  }

  async function disconnect(leave = true) {
    setBusy(true);
    await liveRef.current?.disconnect();
    liveRef.current = null;
    setConnected(false);
    if (leave) {
      const result = await leaveAudioRoom(roomId);
      if (!result.ok) toastError(apiErrorMessage(result));
    }
    await refresh();
    setBusy(false);
  }

  async function toggleMic() {
    const live = liveRef.current;
    if (!live) return;
    const next = !micOn;
    await live.localParticipant.setMicrophoneEnabled(next);
    setMicOn(next);
  }

  if (loading) return <Loading />;
  if (!room) return <ErrorBox message={loadError ?? "Audio room not found"} />;

  return (
    <div className="space-y-4">
      <ReportDialog
        open={reportOpen}
        onOpenChange={setReportOpen}
        targetType="AUDIO_ROOM"
        targetId={room.entity_id ?? null}
        title="Report audio room"
        description="Report this room for spam, harassment, or other abuse."
      />
      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle className="font-serif text-xl">{room.title}</CardTitle>
              <CardDescription>
                Host @{room.host_username ?? "unknown"}
                {room.pod_slug ? ` · pod ${room.pod_slug}` : ""}
              </CardDescription>
            </div>
            <Badge variant={room.status === "ACTIVE" ? "default" : "secondary"}>
              {room.status}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {!connected ? (
              <Button
                disabled={busy || room.status !== "ACTIVE"}
                onClick={() => void connect()}
              >
                Join & talk
              </Button>
            ) : (
              <>
                <Button variant="outline" disabled={busy} onClick={() => void toggleMic()}>
                  {micOn ? <Mic className="size-4" /> : <MicOff className="size-4" />}
                  {micOn ? "Mute" : "Unmute"}
                </Button>
                <Button
                  variant="destructive"
                  disabled={busy}
                  onClick={() => void disconnect(true)}
                >
                  <PhoneOff className="size-4" />
                  Leave
                </Button>
              </>
            )}
            <Button
              variant="outline"
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                const result = await endAudioRoom(roomId);
                setBusy(false);
                if (!result.ok) toastError(apiErrorMessage(result));
                else {
                  await disconnect(false);
                  await refresh();
                }
              }}
            >
              End room
            </Button>
            <Button
              variant="ghost"
              disabled={!room.entity_id}
              onClick={() => setReportOpen(true)}
            >
              Report
            </Button>
          </div>

          <div
            ref={audioRef}
            className="min-h-8 rounded-lg border border-dashed border-border bg-muted/30 p-3 text-sm text-muted-foreground"
          >
            {connected
              ? "Live audio connected — remote speakers appear here."
              : "Join to connect microphone and hear others."}
          </div>

          <div>
            <h3 className="mb-2 text-sm font-medium">Participants</h3>
            {participants.length === 0 ? (
              <p className="text-sm text-muted-foreground">No one here yet.</p>
            ) : (
              <ul className="space-y-2">
                {participants.map((p) => (
                  <li
                    key={p.username}
                    className="flex items-center justify-between rounded-lg border border-border/70 px-3 py-2 text-sm"
                  >
                    <span>
                      @{p.username}
                      {p.name ? ` · ${p.name}` : ""}
                    </span>
                    {p.role ? <Badge variant="outline">{p.role}</Badge> : null}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
