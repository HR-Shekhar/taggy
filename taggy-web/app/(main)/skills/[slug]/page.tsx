"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import {
  ArrowRight,
  CheckCircle2,
  Circle,
  Clock3,
  PauseCircle,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  apiErrorMessage,
  createRoadmapEditRequest,
  getRoadmapVersion,
  getSkill,
  getSkillRoadmap,
  isFreeSkillLimitError,
  joinSkill,
  listMilestones,
  listMyRoadmapEditRequests,
  listMySkills,
  switchRoadmapVersion,
  type RoadmapMilestone,
  type RoadmapVersionSummary,
} from "@/lib/api";
import {
  Empty,
  ErrorBox,
  GenerationWaitNote,
  Loading,
  PageHeader,
} from "@/components/app-ui";
import { PremiumUpgradePrompt } from "@/components/premium-upgrade";
import { toastApiError, toastError, toastSuccess } from "@/lib/toast";
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

type ProgressMilestone = {
  slug: string;
  title: string;
  description?: string;
  order_index: number;
  status: string;
  estimated_hours?: number | null;
  difficulty?: string | null;
  chapter?: string | null;
  kind?: string;
  postponed_until?: string | null;
  completed_at?: string | null;
};

function formatWhen(value?: string | null) {
  if (!value) return null;
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function SkillDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const { username } = useAuth();
  const [skillName, setSkillName] = useState(slug);
  const [description, setDescription] = useState<string | null>(null);
  const [versions, setVersions] = useState<RoadmapVersionSummary[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<number | null>(null);
  const [previewMilestones, setPreviewMilestones] = useState<RoadmapMilestone[]>([]);
  const [progress, setProgress] = useState<ProgressMilestone[] | null>(null);
  const [enrolledVersion, setEnrolledVersion] = useState<number | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);
  const [showUpgrade, setShowUpgrade] = useState(false);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [editRationale, setEditRationale] = useState("");
  const [editBusy, setEditBusy] = useState(false);
  const [editPending, setEditPending] = useState(false);

  async function load(preferredVersion?: number | null) {
    setLoading(true);
    setPageError(null);
    try {
      const [detail, roadmap] = await Promise.all([
        getSkill(slug),
        getSkillRoadmap(slug),
      ]);
      if (detail.ok && detail.data?.skill) {
        setSkillName(detail.data.skill.name);
        setDescription(detail.data.skill.description ?? null);
      }

      let nextSelected: number | null = preferredVersion ?? null;
      if (roadmap.ok && roadmap.data) {
        setVersions(roadmap.data.versions ?? []);
        const official =
          roadmap.data.current_version?.version_number ??
          roadmap.data.versions.find((v) => v.status === "ACTIVE")?.version_number ??
          roadmap.data.versions[0]?.version_number ??
          null;
        if (nextSelected == null) nextSelected = official;
      } else if (!roadmap.ok) {
        const message = apiErrorMessage(roadmap);
        setPageError(message);
        toastError(message, "Couldn't load this skill");
      }

      if (!username) {
        setProgress(null);
        setEnrolledVersion(null);
        setSelectedVersion(nextSelected);
        return;
      }

      const [ms, mine] = await Promise.all([
        listMilestones(username, slug),
        listMySkills(username),
      ]);

      if (mine.ok) {
        const enrollment = (mine.data ?? []).find((s) => s.skill_slug === slug);
        if (enrollment) {
          setEnrolledVersion(enrollment.roadmap_version_number);
          if (preferredVersion == null) {
            nextSelected = enrollment.roadmap_version_number;
          }
        } else {
          setEnrolledVersion(null);
        }
      }

      if (ms.ok) {
        setProgress((ms.data as ProgressMilestone[]) ?? []);
      } else {
        setProgress(null);
      }

      const edits = await listMyRoadmapEditRequests(username);
      if (edits.ok) {
        setEditPending(
          (edits.data ?? []).some(
            (r) => r.skill_slug === slug && r.status === "PENDING"
          )
        );
      }

      setSelectedVersion(nextSelected);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [slug, username]);

  useEffect(() => {
    if (selectedVersion == null) return;
    (async () => {
      const detail = await getRoadmapVersion(slug, selectedVersion);
      if (detail.ok && detail.data) {
        setPreviewMilestones(detail.data.milestones ?? []);
      }
    })();
  }, [slug, selectedVersion]);

  const viewingOwnProgress =
    progress != null &&
    enrolledVersion != null &&
    selectedVersion === enrolledVersion;

  const milestonesToRender = viewingOwnProgress ? progress! : previewMilestones;

  const completion = useMemo(() => {
    if (!progress || progress.length === 0) return 0;
    const done = progress.filter((m) => m.status === "COMPLETED").length;
    return (done / progress.length) * 100;
  }, [progress]);

  if (loading) return <Loading />;

  const selectable = versions.filter((v) => v.status !== "DRAFT");

  return (
    <div className="space-y-6">
      <PageHeader title={skillName} description={description ?? undefined}>
        <div className="flex flex-wrap gap-2">
          <Link
            href={`/community/${slug}`}
            className={cn(buttonVariants({ variant: "outline" }), "gap-1")}
          >
            Community chat
          </Link>
          <Link
            href={`/pods?skill=${slug}`}
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            Pods
          </Link>
          {progress != null ? (
            <Link
              href={`/progress?skill=${slug}`}
              className={cn(buttonVariants(), "gap-1")}
            >
              Track progress
              <ArrowRight className="size-3.5" />
            </Link>
          ) : null}
          {progress == null ? (
            <button
              type="button"
              className={cn(buttonVariants(), "gap-1")}
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                setShowUpgrade(false);
                const result = await joinSkill(slug);
                setBusy(false);
                if (!result.ok) {
                  if (isFreeSkillLimitError(result)) {
                    setShowUpgrade(true);
                  } else {
                    toastApiError(result, "Couldn't join skill");
                  }
                } else void load(selectedVersion);
              }}
            >
              Join skill
              <ArrowRight className="size-4" />
            </button>
          ) : null}
        </div>
      </PageHeader>

      {pageError ? (
        <ErrorBox message={pageError} title="Couldn't load this skill" />
      ) : null}
      {showUpgrade && progress == null ? (
        <PremiumUpgradePrompt message="limit" />
      ) : null}

      {progress != null ? (
        <Card className="rounded-xl ring-1 ring-foreground/10">
          <CardHeader>
            <CardTitle className="font-serif text-lg">
              Request roadmap update
            </CardTitle>
            <CardDescription>
              AI drafts a full course-style topic → subtopic outline (names only).
              This can take up to a few minutes.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {editPending ? (
              <p className="text-sm text-muted-foreground">
                You already have a pending roadmap edit for this skill.{" "}
                <Link href="/requests" className="underline">
                  View requests
                </Link>
              </p>
            ) : (
              <>
                <textarea
                  className="min-h-16 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                  placeholder="What should change? (optional)"
                  value={editRationale}
                  onChange={(e) => {
                    setEditRationale(e.target.value);
                  }}
                  disabled={editBusy}
                />
                <GenerationWaitNote active={editBusy} />
                <Button
                  disabled={editBusy}
                  onClick={async () => {
                    setEditBusy(true);
                    const res = await createRoadmapEditRequest(
                      slug,
                      editRationale.trim() || undefined
                    );
                    setEditBusy(false);
                    if (!res.ok) {
                      toastApiError(res, "Couldn't submit update");
                      return;
                    }
                    const count = res.data?.draft_milestones?.length ?? 0;
                    setEditPending(true);
                    setEditRationale("");
                    toastSuccess(
                      count > 0
                        ? `Drafted ${count} milestones for admin review.`
                        : "Your roadmap update request is pending admin review.",
                      "Roadmap draft ready"
                    );
                  }}
                >
                  {editBusy ? "Generating…" : "Request update"}
                </Button>
              </>
            )}
          </CardContent>
        </Card>
      ) : null}

      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader>
          <CardTitle className="font-serif text-lg">Roadmap version</CardTitle>
          <CardDescription>
            Each skill has one roadmap with published versions. Pick which version
            you follow — matching milestone slugs keep your completed work.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {selectable.map((v) => {
              const active = selectedVersion === v.version_number;
              const isMine = enrolledVersion === v.version_number;
              return (
                <button
                  key={v.version_number}
                  type="button"
                  onClick={() => setSelectedVersion(v.version_number)}
                  className={cn(
                    "rounded-lg border px-3 py-2 text-left text-sm transition-colors",
                    active
                      ? "border-primary bg-secondary"
                      : "border-border hover:bg-muted"
                  )}
                >
                  <div className="flex items-center gap-2 font-medium">
                    v{v.version_number}
                    {isMine ? <Badge variant="default">yours</Badge> : null}
                    {v.is_current ? <Badge variant="outline">official</Badge> : null}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {v.status} · {v.milestone_count} milestones
                  </div>
                </button>
              );
            })}
          </div>

          {progress != null &&
            selectedVersion != null &&
            enrolledVersion !== selectedVersion && (
              <div className="space-y-2">
                <button
                  type="button"
                  className={cn(buttonVariants())}
                  disabled={busy}
                  onClick={async () => {
                    if (!username || selectedVersion == null) return;
                    setBusy(true);
                    const result = await switchRoadmapVersion(
                      username,
                      slug,
                      selectedVersion
                    );
                    setBusy(false);
                    if (!result.ok) toastApiError(result, "Couldn't switch version");
                    else void load(selectedVersion);
                  }}
                >
                  Use v{selectedVersion} as my roadmap
                </button>
              </div>
            )}

          {progress != null && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">
                  Your progress
                  {enrolledVersion != null ? ` on v${enrolledVersion}` : ""}
                </span>
                <span className="font-medium">{Math.round(completion)}%</span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-primary"
                  style={{ width: `${Math.min(100, completion)}%` }}
                />
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader>
          <CardTitle className="font-serif text-lg">
            {selectedVersion != null ? `Milestones · v${selectedVersion}` : "Milestones"}
          </CardTitle>
          <CardDescription>
            {progress == null
              ? "Preview the roadmap. Join the skill to track completion."
              : viewingOwnProgress
                ? "Your roadmap status. Mark topics complete on the Progress page."
                : "Previewing another version. Switch to it to continue tracking."}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {progress != null && viewingOwnProgress ? (
            <Link
              href={`/progress?skill=${slug}`}
              className={cn(buttonVariants({ size: "sm" }), "gap-1")}
            >
              Complete milestones
              <ArrowRight className="size-3.5" />
            </Link>
          ) : null}
          {milestonesToRender.length === 0 ? (
            <Empty>No milestones on this version.</Empty>
          ) : (
            <ol className="relative space-y-0">
              {milestonesToRender.map((m, idx) => {
                const progressRow = viewingOwnProgress
                  ? (m as ProgressMilestone)
                  : null;
                const status = progressRow?.status ?? "PREVIEW";
                const postponedLabel = formatWhen(progressRow?.postponed_until);

                return (
                  <li key={m.slug} className="relative flex gap-4 pb-6 last:pb-0">
                    {idx < milestonesToRender.length - 1 && (
                      <span className="absolute left-[15px] top-8 h-[calc(100%-1.5rem)] w-px bg-border" />
                    )}
                    <StatusIcon status={status} />
                    <div
                      className={cn(
                        "min-w-0 flex-1 space-y-2 rounded-xl border p-4",
                        ("kind" in m && m.kind === "CHAPTER")
                          ? "border-amber-500/30 bg-amber-500/5"
                          : "border-border/70 bg-card/40"
                      )}
                    >
                      <div className="flex flex-wrap items-start justify-between gap-2">
                        <div>
                          {"chapter" in m &&
                          m.chapter &&
                          !("kind" in m && m.kind === "CHAPTER") ? (
                            <p className="mb-1 text-xs uppercase tracking-wide text-muted-foreground">
                              {m.chapter}
                            </p>
                          ) : null}
                          <h3 className="font-medium">
                            {m.order_index}. {m.title}
                          </h3>
                          {m.description && (
                            <p className="mt-1 text-sm text-muted-foreground">
                              {m.description}
                            </p>
                          )}
                          {postponedLabel ? (
                            <p className="mt-1 text-xs text-muted-foreground">
                              Postponed until {postponedLabel}
                            </p>
                          ) : null}
                        </div>
                        <div className="flex flex-wrap gap-1.5">
                          {"kind" in m && m.kind === "CHAPTER" ? (
                            <Badge>Topic</Badge>
                          ) : (
                            <Badge variant="secondary">Subtopic</Badge>
                          )}
                          <Badge variant="outline">
                            {status.replaceAll("_", " ")}
                          </Badge>
                          {"difficulty" in m && m.difficulty ? (
                            <Badge variant="secondary">{m.difficulty}</Badge>
                          ) : null}
                          {"estimated_hours" in m && m.estimated_hours ? (
                            <Badge variant="outline">{m.estimated_hours}h</Badge>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  </li>
                );
              })}
            </ol>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatusIcon({ status }: { status: string }) {
  if (status === "COMPLETED") {
    return <CheckCircle2 className="relative z-10 size-8 shrink-0 bg-background text-primary" />;
  }
  if (status === "POSTPONED") {
    return (
      <PauseCircle className="relative z-10 size-8 shrink-0 bg-background text-muted-foreground" />
    );
  }
  if (status === "IN_PROGRESS") {
    return <Clock3 className="relative z-10 size-8 shrink-0 bg-background text-primary" />;
  }
  return <Circle className="relative z-10 size-8 shrink-0 bg-background text-muted-foreground" />;
}
