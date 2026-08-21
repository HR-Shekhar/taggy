"use client";

import { useEffect, useState } from "react";
import { Mic, MicOff, PhoneOff } from "lucide-react";
import { useAudioRoom } from "@/components/audio-room-provider";
import { ErrorBox, Loading } from "@/components/app-ui";
import { ReportDialog } from "@/components/report-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function AudioRoomSession({
  roomId,
  autoJoin = false,
}: {
  roomId: string;
  autoJoin?: boolean;
}) {
  const {
    room,
    participants,
    loadError,
    loading,
    connected,
    micOn,
    busy,
    bindRoom,
    connect,
    leave,
    endRoom,
    toggleMic,
    roomId: activeId,
  } = useAudioRoom();
  const [reportOpen, setReportOpen] = useState(false);

  useEffect(() => {
    void bindRoom(roomId);
  }, [roomId, bindRoom]);

  useEffect(() => {
    if (
      autoJoin &&
      room &&
      room.id === roomId &&
      room.status === "ACTIVE" &&
      !connected
    ) {
      void connect(roomId);
    }
  }, [autoJoin, room, roomId, connected, connect]);

  if (loading && (!room || activeId !== roomId)) return <Loading />;
  if (!room || room.id !== roomId) {
    return <ErrorBox message={loadError ?? "Audio room not found"} />;
  }

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
      <Card>
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
                onClick={() => void connect(roomId)}
              >
                Join & talk
              </Button>
            ) : (
              <>
                <Button
                  variant="outline"
                  disabled={busy}
                  onClick={() => void toggleMic()}
                >
                  {micOn ? <Mic className="size-4" /> : <MicOff className="size-4" />}
                  {micOn ? "Mute" : "Unmute"}
                </Button>
                <Button
                  variant="destructive"
                  disabled={busy}
                  onClick={() => void leave()}
                >
                  <PhoneOff className="size-4" />
                  Leave
                </Button>
              </>
            )}
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => void endRoom()}
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

          <div className="min-h-8 rounded-lg border border-dashed border-border bg-muted/30 p-3 text-sm text-foreground/75">
            {connected
              ? "Live audio connected — you can leave this page and stay in the call."
              : "Join to connect microphone and hear others."}
          </div>

          <div>
            <h3 className="mb-2 text-sm font-medium">Participants</h3>
            {participants.length === 0 ? (
              <p className="text-sm text-foreground/75">No one here yet.</p>
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
