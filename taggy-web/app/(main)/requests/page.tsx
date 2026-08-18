"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  cancelRoadmapEditRequest,
  cancelSkillRequest,
  listMyRoadmapEditRequests,
  listMySkillRequests,
  type RoadmapEditRequest,
  type SkillCreationRequest,
} from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function MyRequestsPage() {
  const { username } = useAuth();
  const [skills, setSkills] = useState<SkillCreationRequest[]>([]);
  const [edits, setEdits] = useState<RoadmapEditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  async function load() {
    if (!username) return;
    setLoading(true);
    const [skillRes, editRes] = await Promise.all([
      listMySkillRequests(username),
      listMyRoadmapEditRequests(username),
    ]);
    setLoading(false);
    if (!skillRes.ok) {
      toastApiError(skillRes);
      return;
    }
    if (!editRes.ok) {
      toastApiError(editRes);
      return;
    }
    setSkills(skillRes.data ?? []);
    setEdits(editRes.data ?? []);
  }

  useEffect(() => {
    void load();
  }, [username]);

  if (loading) return <Loading />;

  return (
    <div className="space-y-8">
      <PageHeader
        title="My requests"
        description="Skill creation and roadmap update requests awaiting admin review."
      />

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Skill requests</h2>
        {skills.length === 0 ? (
          <Empty>
            No skill requests yet.{" "}
            <Link href="/skills" className="underline">
              Request one from Skills
            </Link>
            .
          </Empty>
        ) : (
          <div className="space-y-3">
            {skills.map((req) => (
              <Card key={req.id} className="rounded-xl ring-1 ring-foreground/10">
                <CardHeader className="pb-2">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <CardTitle className="font-serif text-lg">{req.name}</CardTitle>
                    <Badge variant="secondary">{req.status}</Badge>
                  </div>
                  <CardDescription>
                    {req.draft_milestones?.length ?? 0} draft milestones ·{" "}
                    {new Date(req.created_at).toLocaleString()}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  {req.admin_note ? (
                    <p className="text-sm text-muted-foreground">
                      Note: {req.admin_note}
                    </p>
                  ) : null}
                  {req.status === "PENDING" ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={busyId === req.id}
                      onClick={async () => {
                        if (!username) return;
                        setBusyId(req.id);
                        const res = await cancelSkillRequest(username, req.id);
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Cancel
                    </Button>
                  ) : null}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Roadmap edit requests</h2>
        {edits.length === 0 ? (
          <Empty>No roadmap edit requests yet.</Empty>
        ) : (
          <div className="space-y-3">
            {edits.map((req) => (
              <Card key={req.id} className="rounded-xl ring-1 ring-foreground/10">
                <CardHeader className="pb-2">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <CardTitle className="font-serif text-lg">
                      {req.skill_name}
                    </CardTitle>
                    <Badge variant="secondary">{req.status}</Badge>
                  </div>
                  <CardDescription>
                    <Link href={`/skills/${req.skill_slug}`} className="underline">
                      {req.skill_slug}
                    </Link>{" "}
                    · base v{req.base_version_number}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  {req.admin_note ? (
                    <p className="text-sm text-muted-foreground">
                      Note: {req.admin_note}
                    </p>
                  ) : null}
                  {req.status === "PENDING" ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={busyId === req.id}
                      onClick={async () => {
                        if (!username) return;
                        setBusyId(req.id);
                        const res = await cancelRoadmapEditRequest(
                          username,
                          req.id
                        );
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Cancel
                    </Button>
                  ) : null}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
