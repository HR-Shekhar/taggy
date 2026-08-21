"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  getCommunityLeaderboard,
  type CommunityLeaderboardEntry,
} from "@/lib/api";
import { toastApiError } from "@/lib/toast";

export function CommunityLeaderboardPanel({
  skillSlug,
  compact = false,
}: {
  skillSlug: string;
  compact?: boolean;
}) {
  const [rows, setRows] = useState<CommunityLeaderboardEntry[]>([]);

  useEffect(() => {
    void (async () => {
      const res = await getCommunityLeaderboard(skillSlug);
      if (!res.ok) {
        toastApiError(res);
        return;
      }
      setRows(res.data ?? []);
    })();
  }, [skillSlug]);

  return (
    <div className="space-y-2 px-1 py-2">
      {!compact ? (
        <>
          <p className="text-sm font-medium">Pod standings</p>
          <p className="text-xs text-foreground/75">
            Ranked by sum of each member&apos;s best quiz score.
          </p>
        </>
      ) : null}
      {rows.length === 0 ? (
        <p className="text-xs text-foreground/75">No pod scores yet.</p>
      ) : (
        <ol className="space-y-1 text-sm">
          {rows.map((e) => (
            <li key={e.pod_slug}>
              <Link
                href={`/pods/${e.pod_slug}`}
                className="flex items-center justify-between gap-2 rounded-md px-2 py-1.5 hover:bg-muted/60"
              >
                <span className="truncate">
                  <span className="text-foreground/50">#{e.rank}</span>{" "}
                  {e.pod_name}
                  <span className="text-foreground/50">
                    {" "}
                    · {e.member_count} members
                  </span>
                </span>
                <span className="shrink-0 font-medium">{e.total_score}</span>
              </Link>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
