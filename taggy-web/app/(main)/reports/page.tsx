"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import { listReports } from "@/lib/api";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type Report = {
  id: number;
  target_type: string;
  target_id: number;
  reason: string;
  status: string;
  created_at: string;
  resolved_at?: string | null;
};

function formatWhen(value: string) {
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export default function ReportsPage() {
  const { username } = useAuth();
  const [items, setItems] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!username) return;
    (async () => {
      setLoading(true);
      const result = await listReports(username);
      setLoading(false);
      if (!result.ok) {
        toastApiError(result);
        return;
      }
      const data = result.data;
      setItems(
        Array.isArray(data)
          ? (data as Report[])
          : ((data as { reports?: Report[] })?.reports ?? [])
      );
    })();
  }, [username]);

  if (loading) return <Loading />;

  return (
    <div className="space-y-6">
      <PageHeader
        title="My reports"
        description="Reports you’ve filed. Admins resolve them offline."
      />
      {items.length === 0 ? (
        <Empty>No reports yet.</Empty>
      ) : (
        <div className="space-y-3">
          {items.map((r) => (
            <Card
              key={r.id}
             
            >
              <CardHeader className="pb-2">
                <div className="flex flex-wrap items-center gap-2">
                  <CardTitle className="text-base">
                    {r.target_type.replaceAll("_", " ")} #{r.target_id}
                  </CardTitle>
                  <Badge
                    variant={r.status === "OPEN" ? "outline" : "secondary"}
                  >
                    {r.status}
                  </Badge>
                </div>
                <CardDescription className="mt-1 text-sm text-foreground/80">
                  {r.reason}
                </CardDescription>
              </CardHeader>
              <CardContent className="pt-0 text-xs text-muted-foreground">
                Filed {formatWhen(r.created_at)}
                {r.resolved_at
                  ? ` · Resolved ${formatWhen(r.resolved_at)}`
                  : ""}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
