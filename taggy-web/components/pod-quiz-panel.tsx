"use client";

import { useCallback, useEffect, useState } from "react";
import {
  answerQuizQuestion,
  apiErrorMessage,
  completePodQuiz,
  getPodLeaderboard,
  startPodQuiz,
  startQuizQuestion,
  type PodLeaderboardEntry,
  type PodQuiz,
} from "@/lib/api";
import { toastError } from "@/lib/toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export function PodQuizPanel({
  podSlug,
  enabled,
  mode = "full",
}: {
  podSlug: string;
  enabled: boolean;
  /** @deprecated Errors are shown via toast */
  onError?: (message: string) => void;
  /** full = quiz + leaderboard; quiz = take quiz only; leaderboard = standings only */
  mode?: "full" | "quiz" | "leaderboard";
}) {
  const [leaderboard, setLeaderboard] = useState<PodLeaderboardEntry[]>([]);
  const [quiz, setQuiz] = useState<PodQuiz | null>(null);
  const [order, setOrder] = useState(1);
  const [selected, setSelected] = useState<number[]>([]);
  const [secondsLeft, setSecondsLeft] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  const showError = useCallback((message: string) => {
    toastError(message);
  }, []);

  const loadLeaderboard = useCallback(async () => {
    if (!enabled) return;
    const res = await getPodLeaderboard(podSlug);
    if (res.ok) setLeaderboard(res.data ?? []);
  }, [enabled, podSlug]);

  useEffect(() => {
    void loadLeaderboard();
  }, [loadLeaderboard]);

  useEffect(() => {
    if (secondsLeft == null || secondsLeft <= 0 || !quiz) return;
    const id = window.setInterval(() => {
      setSecondsLeft((s) => (s == null ? s : Math.max(0, s - 1)));
    }, 1000);
    return () => window.clearInterval(id);
  }, [secondsLeft, quiz?.id, order]);

  const current = quiz?.questions.find((q) => q.order_index === order);

  async function beginQuiz() {
    setBusy(true);
    setFeedback(null);
    const res = await startPodQuiz(podSlug);
    setBusy(false);
    if (!res.ok) {
      showError(apiErrorMessage(res));
      return;
    }
    if (!res.data) {
      showError("Couldn't start the quiz. Try again.");
      return;
    }
    setQuiz(res.data);
    setOrder(1);
    setSelected([]);
    const q1 = res.data.questions.find((q) => q.order_index === 1);
    setSecondsLeft(q1?.seconds_left ?? 60);
  }

  async function submitAnswer() {
    if (!quiz || !current || current.answered) return;
    setBusy(true);
    const res = await answerQuizQuestion(
      podSlug,
      quiz.id,
      order,
      selected
    );
    setBusy(false);
    if (!res.ok) {
      showError(apiErrorMessage(res));
      return;
    }
    if (!res.data) {
      showError("Couldn't submit that answer. Try again.");
      return;
    }
    const { is_correct, timed_out, correct_indices } = res.data;
    setFeedback(
      timed_out
        ? "Time's up — marked incorrect."
        : is_correct
          ? "Correct!"
          : "Not quite."
    );
    setQuiz((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        questions: prev.questions.map((q) =>
          q.order_index === order
            ? {
                ...q,
                answered: true,
                is_correct,
                timed_out,
                correct_indices,
              }
            : q
        ),
      };
    });
    setSecondsLeft(null);
  }

  async function goNext() {
    if (!quiz) return;
    if (order >= 10) {
      setBusy(true);
      const res = await completePodQuiz(podSlug, quiz.id);
      setBusy(false);
      if (!res.ok) {
        showError(apiErrorMessage(res));
        return;
      }
      if (!res.data) {
        showError("Couldn't finish the quiz. Try again.");
        return;
      }
      setQuiz(res.data);
      setFeedback(
        `Score ${res.data.score} (${res.data.topic_count} topics × ${res.data.correct_count} correct)`
      );
      void loadLeaderboard();
      return;
    }
    const next = order + 1;
    setOrder(next);
    setSelected([]);
    setFeedback(null);
    setBusy(true);
    const started = await startQuizQuestion(podSlug, quiz.id, next);
    setBusy(false);
    if (!started.ok) {
      showError(apiErrorMessage(started));
      setSecondsLeft(60);
      return;
    }
    if (!started.data) {
      showError("Couldn't load the next question. Try again.");
      setSecondsLeft(60);
      return;
    }
    setSecondsLeft(started.data.seconds_left);
  }

  function toggleOption(idx: number) {
    if (current?.answered || busy) return;
    setSelected((prev) =>
      prev.includes(idx) ? prev.filter((x) => x !== idx) : [...prev, idx]
    );
  }

  const showQuiz = mode === "full" || mode === "quiz";
  const showLeaderboard = mode === "full" || mode === "leaderboard";

  if (!enabled) {
    return (
      <div className="space-y-2 px-1 py-2 text-sm text-muted-foreground">
        {showQuiz
          ? "Join this pod to take evaluation quizzes and appear on the leaderboard."
          : "Join this pod to appear on the leaderboard."}
      </div>
    );
  }

  return (
    <div className="space-y-4 px-1 py-2">
      {showQuiz ? (
        <div className="space-y-2">
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm font-medium">Evaluate</p>
            {!quiz || quiz.status === "COMPLETED" ? (
              <Button size="sm" disabled={busy} onClick={() => void beginQuiz()}>
                {busy ? "Generating…" : "Start quiz"}
              </Button>
            ) : null}
          </div>
          <p className="text-xs text-muted-foreground">
            10 AI questions from your completed topics. 60s each. Score =
            topics × correct.
          </p>
        </div>
      ) : null}

      {showQuiz && quiz && quiz.status === "IN_PROGRESS" && current ? (
        <div className="space-y-3 rounded-lg border border-border/80 bg-muted/30 p-3">
          <div className="flex flex-wrap items-center gap-2 text-xs">
            <Badge variant="outline">Q{order}/10</Badge>
            <Badge variant="secondary">Difficulty {current.difficulty}</Badge>
            {secondsLeft != null && !current.answered ? (
              <Badge variant={secondsLeft <= 10 ? "destructive" : "outline"}>
                {secondsLeft}s
              </Badge>
            ) : null}
          </div>
          <p className="text-sm font-medium leading-snug">{current.prompt}</p>
          <p className="text-xs text-muted-foreground">
            Topic: {current.topic_title} · select all that apply
          </p>
          <ul className="space-y-1.5">
            {current.options.map((opt, idx) => {
              const picked = selected.includes(idx);
              const showKey = current.answered;
              const isRight = current.correct_indices?.includes(idx);
              return (
                <li key={`${order}-${idx}`}>
                  <button
                    type="button"
                    disabled={Boolean(current.answered) || busy}
                    onClick={() => toggleOption(idx)}
                    className={cn(
                      "w-full rounded-lg border px-2.5 py-2 text-left text-sm transition-colors",
                      picked && !showKey && "border-primary bg-secondary",
                      showKey && isRight && "border-emerald-600/50 bg-emerald-500/10",
                      showKey && picked && !isRight && "border-destructive/40 bg-destructive/5",
                      !picked && !showKey && "border-border hover:bg-muted/60"
                    )}
                  >
                    <span className="mr-2 text-muted-foreground">
                      {String.fromCharCode(65 + idx)}.
                    </span>
                    {opt}
                  </button>
                </li>
              );
            })}
          </ul>
          {feedback ? (
            <p className="text-sm text-muted-foreground">{feedback}</p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            {!current.answered ? (
              <Button
                size="sm"
                disabled={busy || selected.length === 0}
                onClick={() => void submitAnswer()}
              >
                Submit
              </Button>
            ) : (
              <Button size="sm" disabled={busy} onClick={() => void goNext()}>
                {order >= 10 ? "Finish & score" : "Next question"}
              </Button>
            )}
          </div>
        </div>
      ) : null}

      {showQuiz && quiz?.status === "COMPLETED" ? (
        <div className="rounded-lg border border-border/80 bg-muted/30 p-3 text-sm">
          <p className="font-medium">Last score: {quiz.score}</p>
          <p className="text-muted-foreground">
            {quiz.correct_count}/10 correct · {quiz.topic_count} completed topics
          </p>
        </div>
      ) : null}

      {showLeaderboard ? (
        <div className="space-y-2">
          {mode === "leaderboard" ? null : (
            <p className="text-sm font-medium">Pod leaderboard</p>
          )}
          {leaderboard.length === 0 ? (
            <p className="text-xs text-muted-foreground">No quiz scores yet.</p>
          ) : (
            <ol className="space-y-1 text-sm">
              {leaderboard.map((e) => (
                <li
                  key={e.public_id}
                  className="flex items-center justify-between gap-2 rounded-md px-2 py-1 hover:bg-muted/50"
                >
                  <span className="truncate">
                    <span className="text-muted-foreground">#{e.rank}</span>{" "}
                    {e.name || e.username}
                  </span>
                  <span className="shrink-0 font-medium">{e.best_score}</span>
                </li>
              ))}
            </ol>
          )}
        </div>
      ) : null}
    </div>
  );
}
