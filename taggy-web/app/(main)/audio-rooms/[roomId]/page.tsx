"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { ArrowLeft } from "lucide-react";
import { AudioRoomSession } from "@/components/audio-room-session";
import { Loading, PageHeader } from "@/components/app-ui";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

function AudioRoomInner() {
  const { roomId } = useParams<{ roomId: string }>();
  const params = useSearchParams();
  const from = params.get("from");

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audio room"
        description="Join with LiveKit to talk with your pod or community."
      >
        {from ? (
          <Link href={from} className={cn(buttonVariants({ variant: "outline" }), "gap-1.5")}>
            <ArrowLeft className="size-4" />
            Back
          </Link>
        ) : null}
      </PageHeader>
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
