"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ChevronDown, ChevronRight, Trash2 } from "lucide-react";
import {
  cancelRoadmapEditRequest,
  cancelSkillRequest,
  listMyRoadmapEditRequests,
  listMySkillRequests,
  type MilestoneDraft,
  type RoadmapEditRequest,
  type SkillCreationRequest,
} from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Empty, PageHeader, PageSkeleton } from "@/components/app-ui";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { toastApiError, toastSuccess } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

function DraftRoadmap({ drafts }: { drafts: MilestoneDraft[] }) {
  if (!drafts?.length) {
    return (
      <p className="text-sm text-foreground/75">No draft milestones yet.</p>
    );
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
    <div className="space-y-4 rounded-xl border border-border bg-muted/30 p-4">
      <p className="text-sm text-foreground/75">
        {drafts.length} milestones across {chapters.length} topic
        {chapters.length === 1 ? "" : "s"}
      </p>
      {chapters.map((ch) => (
        <div key={ch.title} className="space-y-2">
          <h4 className="font-serif text-base font-medium">{ch.title}</h4>
          <ol className="space-y-2">
            {ch.items.map((m) => (
              <li
                key={`${m.order_index}-${m.slug || m.title}`}
                className="rounded-lg border border-border bg-card px-3 py-2"
              >
                <div className="flex flex-wrap items-baseline justify-between gap-2">
                  <span className="font-medium">
                    {m.order_index}. {m.title}
                  </span>
                  {m.kind ? (
                    <Badge variant="outline" className="text-xs">
                      {m.kind}
                    </Badge>
                  ) : null}
                </div>
                {m.description ? (
                  <p className="mt-1 text-sm text-foreground/75">
                    {m.description}
                  </p>
                ) : null}
                {m.estimated_hours > 0 ? (
                  <p className="mt-1 text-sm text-foreground/75">
                    ~{m.estimated_hours}h · {m.difficulty || "unspecified"}
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

type DiscardTarget =
  | { kind: "skill"; id: string; name: string }
  | { kind: "edit"; id: string; name: string }
  | null;

export default function MyRequestsPage() {
  const { username } = useAuth();
  const [skills, setSkills] = useState<SkillCreationRequest[]>([]);
  const [edits, setEdits] = useState<RoadmapEditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [discardTarget, setDiscardTarget] = useState<DiscardTarget>(null);

  async function load(opts?: { silent?: boolean }) {
    if (!username) return;
    if (!opts?.silent) setLoading(true);
    const [skillRes, editRes] = await Promise.all([
      listMySkillRequests(username),
      listMyRoadmapEditRequests(username),
    ]);
    if (!opts?.silent) setLoading(false);
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

  useEffect(() => {
    if (!username) return;
    const hasGenerating =
      skills.some((r) => r.status === "GENERATING") ||
      edits.some((r) => r.status === "GENERATING");
    if (!hasGenerating) return;
    const id = window.setInterval(() => {
      void load({ silent: true });
    }, 5000);
    return () => window.clearInterval(id);
  }, [username, skills, edits]);

  function toggleExpand(id: string) {
    setExpanded((prev) => ({ ...prev, [id]: !prev[id] }));
  }

  async function confirmDiscard() {
    if (!username || !discardTarget) return;
    setBusyId(discardTarget.id);
    const res =
      discardTarget.kind === "skill"
        ? await cancelSkillRequest(username, discardTarget.id)
        : await cancelRoadmapEditRequest(username, discardTarget.id);
    setBusyId(null);
    if (!res.ok) {
      toastApiError(res);
      return;
    }
    setDiscardTarget(null);
    toastSuccess(
      "Draft discarded. Request again with a clearer goal, audience, and topics you want covered.",
      "Be more specific next time"
    );
    void load();
  }

  if (loading) return <PageSkeleton variant="list" />;

  return (
    <div className="space-y-8">
      <PageHeader
        title="My requests"
        description="Review AI-drafted roadmaps, then wait for admin approval — or discard and try again with more detail."
        backHref="/skills"
      />

      <ConfirmDialog
        open={Boolean(discardTarget)}
        onOpenChange={(open) => {
          if (!open) setDiscardTarget(null);
        }}
        title="Discard this draft roadmap?"
        description={
          discardTarget
            ? `This removes “${discardTarget.name}” from the review queue. If the outline felt off, request again and be more specific — skill focus, level (beginner/advanced), and topics you want covered.`
            : ""
        }
        confirmLabel="Discard draft"
        cancelLabel="Keep it"
        destructive
        busy={busyId === discardTarget?.id}
        onConfirm={confirmDiscard}
      />

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Skill requests</h2>
        {skills.length === 0 ? (
          <Empty
            title="No skill requests yet"
            description="Suggest a skill from the Skills page to generate a draft roadmap."
            action={
              <Link href="/skills" className={cn(buttonVariants())}>
                Request a skill
              </Link>
            }
          />
        ) : (
          <div className="space-y-3">
            {skills.map((req) => {
              const draftCount = req.draft_milestones?.length ?? 0;
              const isOpen = Boolean(expanded[req.id]);
              const canDiscard =
                req.status === "PENDING" || req.status === "GENERATING";
              return (
                <Card key={req.id}>
                  <CardHeader className="pb-2">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <CardTitle className="font-serif text-lg">
                        {req.name}
                      </CardTitle>
                      <Badge variant="secondary">{req.status}</Badge>
                    </div>
                    <CardDescription>
                      {req.status === "GENERATING"
                        ? "AI is reviewing and drafting your skill…"
                        : req.status === "APPROVED"
                          ? "Approved — skill is live in the catalog"
                          : req.status === "REJECTED"
                            ? "Not approved"
                            : `${draftCount} draft milestone${draftCount === 1 ? "" : "s"}`}{" "}
                      · {new Date(req.created_at).toLocaleString()}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {req.description ? (
                      <p className="text-sm text-foreground/75">
                        {req.description}
                      </p>
                    ) : null}
                    {req.admin_note ? (
                      <p className="text-sm text-foreground/75">
                        Note: {req.admin_note}
                      </p>
                    ) : null}
                    <div className="flex flex-wrap gap-2">
                      {draftCount > 0 ? (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => toggleExpand(req.id)}
                          className="gap-1"
                        >
                          {isOpen ? (
                            <ChevronDown className="size-4" />
                          ) : (
                            <ChevronRight className="size-4" />
                          )}
                          {isOpen ? "Hide draft" : "View draft roadmap"}
                        </Button>
                      ) : null}
                      {canDiscard ? (
                        <Button
                          size="sm"
                          variant="outline"
                          className="gap-1 text-destructive hover:text-destructive"
                          disabled={busyId === req.id}
                          onClick={() =>
                            setDiscardTarget({
                              kind: "skill",
                              id: req.id,
                              name: req.name,
                            })
                          }
                        >
                          <Trash2 className="size-3.5" />
                          Discard
                        </Button>
                      ) : null}
                    </div>
                    {isOpen ? (
                      <DraftRoadmap drafts={req.draft_milestones ?? []} />
                    ) : null}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="font-serif text-lg">Roadmap edit requests</h2>
        {edits.length === 0 ? (
          <Empty>No roadmap edit requests yet.</Empty>
        ) : (
          <div className="space-y-3">
            {edits.map((req) => {
              const draftCount = req.draft_milestones?.length ?? 0;
              const isOpen = Boolean(expanded[req.id]);
              const canDiscard =
                req.status === "PENDING" || req.status === "GENERATING";
              return (
                <Card key={req.id}>
                  <CardHeader className="pb-2">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <CardTitle className="font-serif text-lg">
                        {req.skill_name}
                      </CardTitle>
                      <Badge variant="secondary">{req.status}</Badge>
                    </div>
                    <CardDescription>
                      <Link
                        href={`/skills/${req.skill_slug}`}
                        className="underline"
                      >
                        {req.skill_slug}
                      </Link>{" "}
                      · base v{req.base_version_number}
                      {draftCount > 0
                        ? ` · ${draftCount} draft milestones`
                        : ""}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {req.rationale ? (
                      <p className="text-sm text-foreground/75">
                        {req.rationale}
                      </p>
                    ) : null}
                    {req.admin_note ? (
                      <p className="text-sm text-foreground/75">
                        Note: {req.admin_note}
                      </p>
                    ) : null}
                    <div className="flex flex-wrap gap-2">
                      {draftCount > 0 ? (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => toggleExpand(req.id)}
                          className="gap-1"
                        >
                          {isOpen ? (
                            <ChevronDown className="size-4" />
                          ) : (
                            <ChevronRight className="size-4" />
                          )}
                          {isOpen ? "Hide draft" : "View draft roadmap"}
                        </Button>
                      ) : null}
                      {canDiscard ? (
                        <Button
                          size="sm"
                          variant="outline"
                          className="gap-1 text-destructive hover:text-destructive"
                          disabled={busyId === req.id}
                          onClick={() =>
                            setDiscardTarget({
                              kind: "edit",
                              id: req.id,
                              name: req.skill_name,
                            })
                          }
                        >
                          <Trash2 className="size-3.5" />
                          Discard
                        </Button>
                      ) : null}
                    </div>
                    {isOpen ? (
                      <DraftRoadmap drafts={req.draft_milestones ?? []} />
                    ) : null}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
