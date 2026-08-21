"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import { usePathname, useRouter } from "next/navigation";
import { ArrowRight } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { listMyPods, listMySkills } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export const ONBOARDING_DONE_KEY = "taggy_onboarding_done";
export const ONBOARDING_STEP_KEY = "taggy_onboarding_step";

export type TourStepId =
  | "home_overview"
  | "skills_join_or_request"
  | "skills_waiting"
  | "pods_join_or_create"
  | "pods_waiting"
  | "pod_chat"
  | "pod_audio"
  | "pod_quiz"
  | "community_leaderboards"
  | "done";

const STEP_ORDER: TourStepId[] = [
  "home_overview",
  "skills_join_or_request",
  "skills_waiting",
  "pods_join_or_create",
  "pods_waiting",
  "pod_chat",
  "pod_audio",
  "pod_quiz",
  "community_leaderboards",
  "done",
];

/** Map legacy step ids from earlier tours */
function normalizeStepId(raw: string): TourStepId | null {
  if (raw === "pod_features") return "pod_chat";
  if (STEP_ORDER.includes(raw as TourStepId)) return raw as TourStepId;
  return null;
}

type StepDef = {
  id: TourStepId;
  title: string;
  body: string;
  href?: string;
  targets: string[];
  ctaLabel: string;
};

const STEPS: Record<TourStepId, StepDef> = {
  home_overview: {
    id: "home_overview",
    title: "Your Home hub",
    body: "This whole area is your dashboard — continue skills, streaks, search, and shortcuts. Click Go to Skills when you’re ready.",
    href: "/home",
    targets: ["home-cta", "home-main", "nav-home"],
    ctaLabel: "Go to Skills",
  },
  skills_join_or_request: {
    id: "skills_join_or_request",
    title: "Join or request a skill",
    body: "Use the catalog cards to Join a skill, or the request form to generate a roadmap. Click a Join button or fill the request form.",
    href: "/skills",
    targets: ["skills-catalog", "skills-request", "skills-page", "nav-skills"],
    ctaLabel: "Continue",
  },
  skills_waiting: {
    id: "skills_waiting",
    title: "Enroll to keep going",
    body: "Click Join on a skill card (or wait for your request). This tip moves on once you have at least one skill.",
    href: "/skills",
    targets: ["skills-catalog", "nav-skills", "skills-page"],
    ctaLabel: "I’ve joined — continue",
  },
  pods_join_or_create: {
    id: "pods_join_or_create",
    title: "Find or create a pod",
    body: "Browse pods for your skill, Join one, or Create pod. After you’re in a pod we’ll walk through chat, audio, and quizzes.",
    href: "/pods",
    targets: ["pods-create", "pods-page", "nav-pods"],
    ctaLabel: "Continue",
  },
  pods_waiting: {
    id: "pods_waiting",
    title: "Get into a pod",
    body: "Join or create a pod on this page. Once accepted, we’ll open it and explore chat, audio rooms, and quizzes together.",
    href: "/pods",
    targets: ["pods-create", "pods-page", "nav-pods"],
    ctaLabel: "I’m in a pod — continue",
  },
  pod_chat: {
    id: "pod_chat",
    title: "Pod chat",
    body: "This is your pod chatroom — send messages, reply, and stay accountable. Try typing a hello, then continue to audio rooms.",
    targets: ["pod-chat", "pod-workspace"],
    ctaLabel: "Next: audio rooms",
  },
  pod_audio: {
    id: "pod_audio",
    title: "Live audio rooms",
    body: "Create or join an audio room from this panel to study together live. Click Create room (or open an existing one), then continue.",
    targets: ["pod-audio-panel", "pod-audio", "pod-audio-mobile", "pod-workspace"],
    ctaLabel: "Next: quizzes",
  },
  pod_quiz: {
    id: "pod_quiz",
    title: "Quizzes & leaderboard",
    body: "Pod quizzes and leaderboards live here. Open Take quiz on Progress when you want to compete — then we’ll visit Community.",
    targets: ["pod-quiz", "pod-members", "pod-workspace"],
    ctaLabel: "Go to Community",
  },
  community_leaderboards: {
    id: "community_leaderboards",
    title: "Community channels",
    body: "Open a skill community for wider chat channels. Leaderboards also appear on Home and in pods — click a community card to explore.",
    href: "/community",
    targets: ["community-page", "nav-community"],
    ctaLabel: "Finish tour",
  },
  done: {
    id: "done",
    title: "You’re set",
    body: "You covered Home → Skills → Pods (chat, audio, quiz) → Community. Click Home whenever you want to continue learning.",
    href: "/home",
    targets: ["home-cta", "nav-home"],
    ctaLabel: "Go to Home",
  },
};

type OnboardingContextValue = {
  active: boolean;
  stepId: TourStepId | null;
  skip: () => void;
  next: () => void;
  goToStep: (id: TourStepId) => void;
};

const OnboardingContext = createContext<OnboardingContextValue | null>(null);

export function useOnboarding() {
  return useContext(OnboardingContext);
}

function doneKey(username: string) {
  return `${ONBOARDING_DONE_KEY}:${username}`;
}
function stepKey(username: string) {
  return `${ONBOARDING_STEP_KEY}:${username}`;
}

function readDone(username: string) {
  if (typeof window === "undefined") return true;
  if (window.localStorage.getItem(doneKey(username)) === "1") return true;
  // Legacy global flag from earlier tour
  if (window.localStorage.getItem(ONBOARDING_DONE_KEY) === "1") return true;
  return false;
}

function readStep(username: string): TourStepId | null {
  if (typeof window === "undefined") return null;
  const raw =
    window.localStorage.getItem(stepKey(username)) ??
    window.localStorage.getItem(ONBOARDING_STEP_KEY);
  if (!raw) return null;
  return normalizeStepId(raw);
}

function writeStep(username: string, id: TourStepId) {
  window.localStorage.setItem(stepKey(username), id);
}

function markDone(username: string) {
  window.localStorage.setItem(doneKey(username), "1");
  window.localStorage.removeItem(stepKey(username));
  window.localStorage.setItem(ONBOARDING_DONE_KEY, "1");
  window.localStorage.removeItem(ONBOARDING_STEP_KEY);
  window.dispatchEvent(new Event("taggy-onboarding-done"));
}

function stepIndex(id: TourStepId) {
  return STEP_ORDER.indexOf(id);
}

type SpotlightRect = {
  top: number;
  left: number;
  width: number;
  height: number;
};

function findTargetEl(targets: string[]) {
  for (const id of targets) {
    const el = document.querySelector(`[data-tour="${id}"]`);
    if (el instanceof HTMLElement) {
      const r = el.getBoundingClientRect();
      // Prefer visible, reasonably sized targets
      if (r.width > 0 && r.height > 0) return el;
    }
  }
  return null;
}

function useTargetRect(targets: string[], active: boolean) {
  const [rect, setRect] = useState<SpotlightRect | null>(null);

  const measure = useCallback(() => {
    if (!active || targets.length === 0) {
      setRect(null);
      return;
    }
    const el = findTargetEl(targets);
    if (!el) {
      setRect(null);
      return;
    }
    const r = el.getBoundingClientRect();
    // Full feature card — only clamp to the visible viewport
    const top = Math.max(8, r.top);
    const left = Math.max(8, r.left);
    const right = Math.min(window.innerWidth - 8, r.right);
    const bottom = Math.min(window.innerHeight - 8, r.bottom);
    setRect({
      top,
      left,
      width: Math.max(40, right - left),
      height: Math.max(40, bottom - top),
    });
  }, [targets, active]);

  useLayoutEffect(() => {
    if (!active) return;
    const el = findTargetEl(targets);
    el?.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
    measure();
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    const t = window.setInterval(measure, 400);
    return () => {
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
      window.clearInterval(t);
    };
  }, [measure, active, targets]);

  return rect;
}

function clamp(n: number, min: number, max: number) {
  return Math.min(max, Math.max(min, n));
}

function tipPosition(rect: SpotlightRect | null): CSSProperties {
  const tipW = Math.min(352, window.innerWidth - 24);
  const tipH = 220;
  const margin = 12;
  const gap = 12;

  if (!rect) {
    return {
      position: "fixed",
      left: "50%",
      bottom: margin + 8,
      transform: "translateX(-50%)",
      width: tipW,
      maxHeight: "min(70vh, 22rem)",
    };
  }

  const spaceRight = window.innerWidth - (rect.left + rect.width);
  const spaceLeft = rect.left;
  const spaceBelow = window.innerHeight - (rect.top + rect.height);
  const spaceAbove = rect.top;

  let top: number;
  let left: number;

  // Prefer: right of target → below → above → left → overlay near top of target
  if (spaceRight >= tipW + gap + 8) {
    left = rect.left + rect.width + gap;
    top = rect.top;
  } else if (spaceBelow >= tipH + gap) {
    left = rect.left + Math.max(0, (Math.min(rect.width, tipW + 40) - tipW) / 2);
    top = rect.top + rect.height + gap;
  } else if (spaceAbove >= tipH + gap) {
    left = rect.left + Math.max(0, (Math.min(rect.width, tipW + 40) - tipW) / 2);
    top = rect.top - tipH - gap;
  } else if (spaceLeft >= tipW + gap + 8) {
    left = rect.left - tipW - gap;
    top = rect.top;
  } else {
    // Tall/full-page spotlight: keep tip near the target's top edge
    left = rect.left + Math.max(0, (rect.width - tipW) / 2);
    top = rect.top + gap;
  }

  left = clamp(left, margin, window.innerWidth - tipW - margin);
  top = clamp(top, margin, window.innerHeight - tipH - margin);

  return {
    position: "fixed",
    top,
    left,
    width: tipW,
    maxHeight: "min(70vh, 22rem)",
  };
}

function Coachmark({
  step,
  stepNum,
  stepTotal,
  onNext,
  onSkip,
  onGoThere,
  needsNavigate,
}: {
  step: StepDef;
  stepNum: number;
  stepTotal: number;
  onNext: () => void;
  onSkip: () => void;
  onGoThere: () => void;
  needsNavigate: boolean;
}) {
  const rect = useTargetRect(step.targets, true);
  const pad = 6;
  const style = useMemo(() => tipPosition(rect), [rect]);

  return (
    <div className="pointer-events-none fixed inset-0 z-[55] overflow-hidden">
      {rect ? (
        <div
          aria-hidden
          className="pointer-events-none absolute rounded-xl ring-2 ring-primary/95 shadow-[0_0_0_9999px_rgba(0,0,0,0.5)] ring-offset-2 ring-offset-background dark:ring-[3px] dark:ring-primary dark:shadow-[0_0_0_9999px_rgba(0,0,0,0.78),0_0_0_4px_var(--primary),0_0_36px_10px_color-mix(in_oklab,var(--primary)_45%,transparent)] dark:ring-offset-0"
          style={{
            top: rect.top - pad,
            left: rect.left - pad,
            width: rect.width + pad * 2,
            height: rect.height + pad * 2,
          }}
        />
      ) : null}

      <div
        className="pointer-events-auto z-[56] max-h-[min(70vh,22rem)] overflow-y-auto rounded-xl border border-border bg-card p-4 shadow-xl"
        style={style}
        role="dialog"
        aria-labelledby="tour-title"
      >
        <p className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
          Tour · {stepNum}/{stepTotal}
        </p>
        <h2 id="tour-title" className="mt-1 font-serif text-lg leading-snug">
          {step.title}
        </h2>
        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
          {step.body}
        </p>
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {needsNavigate ? (
            <Button
              type="button"
              size="sm"
              className="gap-1.5"
              onClick={onGoThere}
            >
              Take me there
              <ArrowRight className="size-3.5" />
            </Button>
          ) : (
            <Button
              type="button"
              size="sm"
              className="gap-1.5"
              onClick={onNext}
            >
              {step.ctaLabel}
              <ArrowRight className="size-3.5" />
            </Button>
          )}
          <Button type="button" size="sm" variant="ghost" onClick={onSkip}>
            Skip tour
          </Button>
        </div>
      </div>
    </div>
  );
}

const POD_EXPLORE_STEPS: TourStepId[] = ["pod_chat", "pod_audio", "pod_quiz"];

export function OnboardingProvider({ children }: { children: ReactNode }) {
  const { username, ready } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [active, setActive] = useState(false);
  const [stepId, setStepId] = useState<TourStepId | null>(null);
  const [acceptedPodSlug, setAcceptedPodSlug] = useState<string | null>(null);
  const [skillCount, setSkillCount] = useState(0);
  const [podCount, setPodCount] = useState(0);
  const [bootstrapped, setBootstrapped] = useState(false);

  // New users only: per-username flag; skip if they already have skills/pods
  useEffect(() => {
    if (!ready || !username) return;
    let cancelled = false;

    (async () => {
      if (readDone(username)) {
        if (!cancelled) {
          setActive(false);
          setStepId(null);
          setBootstrapped(true);
        }
        return;
      }

      const inProgress = readStep(username);
      const [sk, pd] = await Promise.all([
        listMySkills(username),
        listMyPods(username),
      ]);
      if (cancelled) return;

      const skills = sk.ok ? sk.data ?? [] : [];
      const pods = pd.ok
        ? Array.isArray(pd.data)
          ? pd.data
          : ((pd.data as { pods?: { status?: string }[] })?.pods ?? [])
        : [];
      const accepted = pods.filter(
        (p) => (p.status ?? "ACCEPTED") === "ACCEPTED"
      );

      setSkillCount(skills.length);
      setPodCount(accepted.length);
      const first = accepted[0] as
        | { pod_slug?: string; slug?: string }
        | undefined;
      setAcceptedPodSlug(first?.pod_slug ?? first?.slug ?? null);

      // Existing users (already enrolled) — never show tour unless mid-tour
      if (!inProgress && (skills.length > 0 || accepted.length > 0)) {
        markDone(username);
        setActive(false);
        setStepId(null);
        setBootstrapped(true);
        return;
      }

      const start = inProgress ?? "home_overview";
      writeStep(username, start);
      setStepId(start);
      setActive(true);
      setBootstrapped(true);
    })();

    return () => {
      cancelled = true;
    };
  }, [ready, username]);

  const goToStep = useCallback(
    (id: TourStepId) => {
      if (!username) return;
      writeStep(username, id);
      setStepId(id);
      setActive(true);
    },
    [username]
  );

  const skip = useCallback(() => {
    if (!username) return;
    markDone(username);
    setActive(false);
    setStepId(null);
  }, [username]);

  const advanceFrom = useCallback(
    (current: TourStepId) => {
      const i = stepIndex(current);
      if (i < 0 || i >= STEP_ORDER.length - 1) {
        skip();
        return;
      }
      goToStep(STEP_ORDER[i + 1]);
    },
    [goToStep, skip]
  );

  const next = useCallback(() => {
    if (!stepId) return;
    if (stepId === "done") {
      skip();
      return;
    }
    advanceFrom(stepId);
  }, [stepId, advanceFrom, skip]);

  useEffect(() => {
    if (!active || !username) return;
    let cancelled = false;
    async function refresh() {
      const [sk, pd] = await Promise.all([
        listMySkills(username!),
        listMyPods(username!),
      ]);
      if (cancelled) return;
      const skills = sk.ok ? sk.data ?? [] : [];
      setSkillCount(skills.length);
      const pods = pd.ok
        ? Array.isArray(pd.data)
          ? pd.data
          : ((pd.data as { pods?: { pod_slug?: string; slug?: string; status?: string }[] })
              ?.pods ?? [])
        : [];
      const accepted = pods.filter(
        (p) => (p.status ?? "ACCEPTED") === "ACCEPTED"
      );
      setPodCount(accepted.length);
      const first = accepted[0];
      setAcceptedPodSlug(first?.pod_slug ?? first?.slug ?? null);
    }
    void refresh();
    const t = window.setInterval(() => void refresh(), 4000);
    return () => {
      cancelled = true;
      window.clearInterval(t);
    };
  }, [active, username]);

  useEffect(() => {
    if (!active || !stepId) return;
    if (stepId === "skills_waiting" && skillCount > 0) {
      goToStep("pods_join_or_create");
    }
    if (stepId === "pods_waiting" && podCount > 0) {
      goToStep("pod_chat");
    }
  }, [active, stepId, skillCount, podCount, goToStep]);

  const step = stepId ? STEPS[stepId] : null;

  const inPodDetail = /^\/pods\/[^/]+/.test(pathname);

  const onPreferredRoute = useMemo(() => {
    if (!step) return true;
    if (POD_EXPLORE_STEPS.includes(step.id)) return inPodDetail;
    if (step.id === "skills_waiting") return pathname.startsWith("/skills");
    if (step.id === "pods_waiting") return pathname.startsWith("/pods");
    if (!step.href) return true;
    if (step.href === "/home") return pathname === "/home";
    return pathname === step.href || pathname.startsWith(`${step.href}/`);
  }, [step, pathname, inPodDetail]);

  const goThere = useCallback(() => {
    if (!step) return;
    if (POD_EXPLORE_STEPS.includes(step.id)) {
      router.push(acceptedPodSlug ? `/pods/${acceptedPodSlug}` : "/pods");
      return;
    }
    if (step.href) router.push(step.href);
  }, [step, acceptedPodSlug, router]);

  // Keep user inside a pod while exploring chat / audio / quiz
  useEffect(() => {
    if (!active || !stepId) return;
    if (!POD_EXPLORE_STEPS.includes(stepId)) return;
    if (!inPodDetail && acceptedPodSlug) {
      router.push(`/pods/${acceptedPodSlug}`);
    }
  }, [active, stepId, inPodDetail, acceptedPodSlug, router]);

  const value = useMemo(
    () => ({ active, stepId, skip, next, goToStep }),
    [active, stepId, skip, next, goToStep]
  );

  return (
    <OnboardingContext.Provider value={value}>
      {children}
      {bootstrapped && active && step ? (
        <Coachmark
          step={step}
          stepNum={stepIndex(step.id) + 1}
          stepTotal={STEP_ORDER.length}
          onSkip={skip}
          needsNavigate={!onPreferredRoute}
          onGoThere={goThere}
          onNext={() => {
            if (step.id === "done") {
              skip();
              router.push("/home");
              return;
            }
            if (step.id === "home_overview") {
              goToStep("skills_join_or_request");
              router.push("/skills");
              return;
            }
            if (step.id === "skills_join_or_request") {
              goToStep(
                skillCount > 0 ? "pods_join_or_create" : "skills_waiting"
              );
              if (skillCount > 0) router.push("/pods");
              return;
            }
            if (step.id === "pods_join_or_create") {
              if (podCount > 0) {
                goToStep("pod_chat");
                if (acceptedPodSlug) router.push(`/pods/${acceptedPodSlug}`);
              } else {
                goToStep("pods_waiting");
              }
              return;
            }
            if (step.id === "pod_chat") {
              goToStep("pod_audio");
              return;
            }
            if (step.id === "pod_audio") {
              goToStep("pod_quiz");
              return;
            }
            if (step.id === "pod_quiz") {
              goToStep("community_leaderboards");
              router.push("/community");
              return;
            }
            if (step.id === "community_leaderboards") {
              goToStep("done");
              return;
            }
            next();
          }}
        />
      ) : null}
    </OnboardingContext.Provider>
  );
}

export function OnboardingChecklist() {
  return null;
}

export function OnboardingTour() {
  return null;
}

export function TourTarget({
  id,
  className,
  children,
}: {
  id: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div data-tour={id} className={cn(className)}>
      {children}
    </div>
  );
}
