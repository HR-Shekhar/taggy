"use client";

import Link from "next/link";
import { Suspense, useEffect, useState, type ReactNode } from "react";
import { useSearchParams } from "next/navigation";
import { Search as SearchIcon } from "lucide-react";
import { search } from "@/lib/api";
import { Empty, Loading, PageHeader } from "@/components/app-ui";
import { toastApiError } from "@/lib/toast";
import { UserAvatar } from "@/components/user-avatar";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type SearchResult = {
  skills?: Array<{ slug: string; name: string; description?: string }>;
  users?: Array<{
    username: string;
    name: string;
    profile_picture_url?: string | null;
  }>;
  communities?: Array<{ skill_slug: string; name: string }>;
};

function SearchInner() {
  const params = useSearchParams();
  const initial = params.get("q") ?? "";
  const [q, setQ] = useState(initial);
  const [result, setResult] = useState<SearchResult | null>(null);
  const [busy, setBusy] = useState(false);

  async function runSearch(query: string) {
    const trimmed = query.trim();
    if (!trimmed) return;
    setBusy(true);
    const r = await search(trimmed, "skills,users,communities");
    setBusy(false);
    if (!r.ok) {
      toastApiError(r);
      return;
    }
    setResult(r.data as SearchResult);
  }

  useEffect(() => {
    const fromUrl = params.get("q") ?? "";
    setQ(fromUrl);
    if (fromUrl.trim()) void runSearch(fromUrl);
  }, [params]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Search"
        description="Find skills, people, and communities."
        backHref="/home"
      />

      <form
        className="relative"
        onSubmit={async (e) => {
          e.preventDefault();
          await runSearch(q);
        }}
      >
        <SearchIcon className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-foreground/50" />
        <Input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Global search in Taggy"
          aria-label="Global search in Taggy"
          className="h-11 rounded-xl border-foreground/15 bg-background/80 pl-10 pr-24 shadow-sm"
          required
        />
        <Button
          type="submit"
          size="sm"
          disabled={busy}
          className="absolute right-2 top-1/2 -translate-y-1/2"
        >
          {busy ? "…" : "Search"}
        </Button>
      </form>

      {result && (
        <div className="grid gap-4 lg:grid-cols-3">
          <ResultCard title="Skills">
            {!result.skills?.length ? (
              <Empty>None</Empty>
            ) : (
              <ul className="space-y-2">
                {result.skills.map((s) => (
                  <li key={s.slug}>
                    <Link
                      href={`/skills/${s.slug}`}
                      className={cn(
                        buttonVariants({ variant: "ghost" }),
                        "h-auto justify-start px-2 py-1"
                      )}
                    >
                      {s.name}
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </ResultCard>
          <ResultCard title="Users">
            {!result.users?.length ? (
              <Empty>None</Empty>
            ) : (
              <ul className="space-y-2">
                {result.users.map((u) => (
                  <li key={u.username}>
                    <Link
                      href={`/u/${u.username}`}
                      className={cn(
                        buttonVariants({ variant: "ghost" }),
                        "h-auto w-full justify-start gap-2 px-2 py-1"
                      )}
                    >
                      <UserAvatar
                        name={u.name}
                        username={u.username}
                        src={u.profile_picture_url}
                        className="size-7"
                        fallbackClassName="text-[10px]"
                      />
                      <span className="truncate">
                        @{u.username} · {u.name}
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </ResultCard>
          <ResultCard title="Communities">
            {!result.communities?.length ? (
              <Empty>None</Empty>
            ) : (
              <ul className="space-y-2">
                {result.communities.map((c) => (
                  <li key={c.skill_slug}>
                    <Link
                      href={`/community/${c.skill_slug}`}
                      className={cn(
                        buttonVariants({ variant: "ghost" }),
                        "h-auto justify-start px-2 py-1"
                      )}
                    >
                      {c.name}
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </ResultCard>
        </div>
      )}
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<Loading />}>
      <SearchInner />
    </Suspense>
  );
}

function ResultCard({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="font-serif text-lg">{title}</CardTitle>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
