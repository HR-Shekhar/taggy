"use client";

import { useParams } from "next/navigation";
import { Suspense } from "react";
import { AudioRoomSession } from "@/components/audio-room-session";
import { Loading, PageHeader } from "@/components/app-ui";

function AudioRoomInner() {
  const { roomId } = useParams<{ roomId: string }>();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audio room"
        description="Join with LiveKit to talk with your pod or community."
        backHref="/pods"
      />
      <AudioRoomSession roomId={roomId} autoJoin />
    </div>
  );
}

export default function AudioRoomPage() {
  return (
    <Suspense fallback={<Loading />}>
      <AudioRoomInner />
    </Suspense>
  );
}
