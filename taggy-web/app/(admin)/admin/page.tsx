"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  adminApproveRoadmapEditRequest,
  adminApproveSkillRequest,
  adminListRoadmapEditRequests,
  adminListSkillRequests,
  adminRejectRoadmapEditRequest,
  adminRejectSkillRequest,
  type MilestoneDraft,
  type RoadmapEditRequest,
  type SkillCreationRequest,
} from "@/lib/api";
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

function DraftList({ drafts }: { drafts: MilestoneDraft[] }) {
  if (!drafts?.length) {
    return <p className="text-sm text-muted-foreground">No draft milestones.</p>;
  }

  const sorted = drafts.slice().sort((a, b) => a.order_index - b.order_index);
  const chapters: { title: string; items: MilestoneDraft[] }[] = [];
  for (const m of sorted) {
    const chapterTitle = m.chapter?.trim() || "General";
    const last = chapters[chapters.length - 1];
    if (!last || last.title !== chapterTitle) {
      chapters.push({ title: chapterTitle, items: [m] });
    } else {
      last.items.push(m);
    }
  }

  return (
    <div className="space-y-4 text-sm">
          <p className="text-xs text-muted-foreground">
            {drafts.length} items across {chapters.length} topics (names only)
          </p>
      {chapters.map((ch) => (
        <div key={ch.title} className="space-y-2">
          <h4 className="font-serif text-base font-medium">{ch.title}</h4>
          <ol className="space-y-2">
            {ch.items.map((m) => (
              <li
                key={`${m.order_index}-${m.slug}`}
                className="rounded-lg bg-muted/50 px-3 py-2"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium">
                    {m.order_index}. {m.title}
                  </span>
                  {m.kind === "CHAPTER" ? (
                    <Badge>Topic</Badge>
                  ) : (
                    <Badge variant="secondary">Subtopic</Badge>
                  )}
                  {m.difficulty ? (
                    <Badge variant="outline">{m.difficulty}</Badge>
                  ) : null}
                  {m.estimated_hours ? (
                    <span className="text-xs text-muted-foreground">
                      ~{m.estimated_hours}h
                    </span>
                  ) : null}
                </div>
                {m.description ? (
                  <p className="mt-1 line-clamp-3 text-muted-foreground">
                    {m.description}
                  </p>
                ) : null}
              </li>
            ))}
          </ol>
        </div>
      ))}
    </div>
  );
}

export default function AdminPage() {
  const [skills, setSkills] = useState<SkillCreationRequest[]>([]);
  const [edits, setEdits] = useState<RoadmapEditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [notes, setNotes] = useState<Record<string, string>>({});

  async function load() {
    setLoading(true);
    const [skillRes, editRes] = await Promise.all([
      adminListSkillRequests(),
      adminListRoadmapEditRequests(),
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
  }, []);

  if (loading) return <Loading />;

  const pendingSkills = skills.length;
  const pendingEdits = edits.length;

  return (
    <div className="space-y-8">
      <PageHeader
        title="Approvals"
        description="Only platform admins can approve or reject catalog changes. Learners never see this page."
      />

      <div className="grid gap-3 sm:grid-cols-2">
        <div className="rounded-xl border border-border/80 bg-muted/30 px-4 py-3">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">
            Skill requests
          </p>
          <p className="font-serif text-2xl tabular-nums">{pendingSkills}</p>
        </div>
        <div className="rounded-xl border border-border/80 bg-muted/30 px-4 py-3">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">
            Roadmap edits
          </p>
          <p className="font-serif text-2xl tabular-nums">{pendingEdits}</p>
        </div>
      </div>

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Skill creation queue</h2>
        {skills.length === 0 ? (
          <Empty>No pending skill requests.</Empty>
        ) : (
          <div className="space-y-4">
            {skills.map((req) => (
              <Card key={req.id} className="rounded-xl ring-1 ring-foreground/10">
                <CardHeader>
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <CardTitle className="font-serif text-lg">{req.name}</CardTitle>
                      <CardDescription>
                        slug candidate: {req.slug_candidate}
                      </CardDescription>
                    </div>
                    <Badge>{req.status}</Badge>
                  </div>
                  {req.description ? (
                    <p className="text-sm text-muted-foreground">{req.description}</p>
                  ) : null}
                </CardHeader>
                <CardContent className="space-y-4">
                  <DraftList drafts={req.draft_milestones ?? []} />
                  <textarea
                    className="min-h-16 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    placeholder="Optional reject note"
                    value={notes[req.id] ?? ""}
                    onChange={(e) =>
                      setNotes((prev) => ({ ...prev, [req.id]: e.target.value }))
                    }
                  />
                  <div className="flex flex-wrap gap-2">
                    <Button
                      disabled={busyId === req.id}
                      onClick={async () => {
                        setBusyId(req.id);
                        const res = await adminApproveSkillRequest(req.id);
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Approve
                    </Button>
                    <Button
                      variant="outline"
                      disabled={busyId === req.id}
                      onClick={async () => {
                        setBusyId(req.id);
                        const res = await adminRejectSkillRequest(
                          req.id,
                          notes[req.id]
                        );
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Reject
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Roadmap edit queue</h2>
        {edits.length === 0 ? (
          <Empty>No pending roadmap edits.</Empty>
        ) : (
          <div className="space-y-4">
            {edits.map((req) => (
              <Card key={req.id} className="rounded-xl ring-1 ring-foreground/10">
                <CardHeader>
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <CardTitle className="font-serif text-lg">
                        {req.skill_name}
                      </CardTitle>
                      <CardDescription>
                        <Link href={`/skills/${req.skill_slug}`} className="underline">
                          {req.skill_slug}
                        </Link>{" "}
                        · base v{req.base_version_number}
                      </CardDescription>
                    </div>
                    <Badge>{req.status}</Badge>
                  </div>
                  {req.rationale ? (
                    <p className="text-sm text-muted-foreground">{req.rationale}</p>
                  ) : null}
                </CardHeader>
                <CardContent className="space-y-4">
                  <DraftList drafts={req.draft_milestones ?? []} />
                  <textarea
                    className="min-h-16 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    placeholder="Optional reject note"
                    value={notes[req.id] ?? ""}
                    onChange={(e) =>
                      setNotes((prev) => ({ ...prev, [req.id]: e.target.value }))
                    }
                  />
                  <div className="flex flex-wrap gap-2">
                    <Button
                      disabled={busyId === req.id}
                      onClick={async () => {
                        setBusyId(req.id);
                        const res = await adminApproveRoadmapEditRequest(req.id);
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Approve &amp; publish
                    </Button>
                    <Button
                      variant="outline"
                      disabled={busyId === req.id}
                      onClick={async () => {
                        setBusyId(req.id);
                        const res = await adminRejectRoadmapEditRequest(
                          req.id,
                          notes[req.id]
                        );
                        setBusyId(null);
                        if (!res.ok) toastApiError(res);
                        else void load();
                      }}
                    >
                      Reject
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
