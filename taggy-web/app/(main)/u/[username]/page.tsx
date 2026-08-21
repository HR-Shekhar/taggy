"use client";

import Image from "next/image";
import { useParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { Camera, Loader2, Shield } from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  apiErrorMessage,
  createReport,
  getProfile,
  updateProfile,
  uploadProfilePhoto,
} from "@/lib/api";
import {
  Empty,
  ErrorBox,
  PageSkeleton,
} from "@/components/app-ui";
import { toastError, toastSuccess } from "@/lib/toast";
import { UserAvatar } from "@/components/user-avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

const MAX_AVATAR_BYTES = 500 * 1024;

type Profile = {
  username?: string;
  name?: string;
  bio?: string | null;
  profile_picture_url?: string | null;
  email?: string;
  email_verified?: boolean;
  subscription?: string;
  is_admin?: boolean;
  id?: number;
};

export default function ProfilePage() {
  const { username: routeUser } = useParams<{ username: string }>();
  const { username: me, refreshLocal, acceptTokens, setAvatarUrlState } =
    useAuth();
  const isSelf = me === routeUser;
  const fileRef = useRef<HTMLInputElement>(null);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [name, setName] = useState("");
  const [bio, setBio] = useState("");
  const [newUsername, setNewUsername] = useState("");
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [reportReason, setReportReason] = useState("");

  async function load() {
    setLoading(true);
    const result = await getProfile(routeUser);
    setLoading(false);
    if (!result.ok) {
      setLoadError(apiErrorMessage(result));
      setProfile(null);
      return;
    }
    const data = (result.data ?? {}) as Profile;
    setProfile(data);
    setName(String(data.name ?? ""));
    setBio(String(data.bio ?? ""));
    setNewUsername(String(data.username ?? routeUser));
    if (isSelf) {
      setAvatarUrlState(
        typeof data.profile_picture_url === "string"
          ? data.profile_picture_url
          : null
      );
    }
  }

  useEffect(() => {
    void load();
  }, [routeUser]);

  if (loading) return <PageSkeleton variant="detail" />;
  if (!profile) return <ErrorBox message={loadError ?? "Profile not found"} />;

  const picture =
    typeof profile.profile_picture_url === "string"
      ? profile.profile_picture_url
      : null;
  const displayName = profile.name || profile.username || routeUser;

  async function onPickPhoto(file: File | undefined) {
    if (!file || !isSelf) return;
    if (file.size > MAX_AVATAR_BYTES) {
      toastError("Choose a JPG, PNG, WEBP, or GIF under 500 KB.");
      return;
    }
    setUploading(true);
    const result = await uploadProfilePhoto(routeUser, file);
    setUploading(false);
    if (!result.ok) {
      toastError(apiErrorMessage(result));
      return;
    }
    const data = (result.data ?? {}) as Profile;
    setProfile(data);
    const url =
      typeof data.profile_picture_url === "string"
        ? data.profile_picture_url
        : null;
    setAvatarUrlState(url);
    toastSuccess("Profile photo updated.");
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Card className="gap-0 overflow-hidden rounded-2xl p-0">
        <div className="relative h-32 w-full sm:h-40">
          <Image
            src="/images/banner.jpg"
            alt=""
            fill
            priority
            sizes="(max-width: 768px) 100vw, 48rem"
            className="object-cover object-center"
          />
          <div className="absolute bottom-0 left-5 z-10 translate-y-1/2 sm:left-7">
            <div className="relative">
              <UserAvatar
                name={profile.name}
                username={profile.username}
                src={picture}
                className="size-24 ring-4 ring-card sm:size-28"
                fallbackClassName="text-2xl"
              />
              {isSelf ? (
                <>
                  <input
                    ref={fileRef}
                    type="file"
                    accept="image/jpeg,image/png,image/webp,image/gif"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      e.target.value = "";
                      void onPickPhoto(file);
                    }}
                  />
                  <Button
                    type="button"
                    size="icon"
                    variant="secondary"
                    className="absolute right-0 bottom-0 size-8 rounded-full shadow-sm"
                    disabled={uploading}
                    aria-label="Upload profile photo"
                    onClick={() => fileRef.current?.click()}
                  >
                    {uploading ? (
                      <Loader2 className="size-3.5 animate-spin" />
                    ) : (
                      <Camera className="size-3.5" />
                    )}
                  </Button>
                </>
              ) : null}
            </div>
          </div>
        </div>
        <CardContent className="px-5 pb-6 pt-14 sm:px-7 sm:pt-16">
          <div className="space-y-5 sm:pl-32">
            <div className="space-y-2">
              <div>
                <h1 className="font-serif text-2xl font-medium tracking-tight">
                  {displayName}
                </h1>
                <p className="text-sm text-foreground/75">
                  @{profile.username ?? routeUser}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                {profile.subscription ? (
                  <Badge variant="secondary">{profile.subscription}</Badge>
                ) : null}
                {profile.is_admin ? (
                  <Badge className="gap-1">
                    <Shield className="size-3" />
                    Admin
                  </Badge>
                ) : null}
                {profile.email_verified ? (
                  <Badge variant="outline">Verified</Badge>
                ) : null}
              </div>
              {uploading ? (
                <p className="text-xs text-foreground/75">Uploading photo…</p>
              ) : isSelf ? (
                <p className="text-xs text-foreground/75">
                  Click the camera to change your photo. JPG, PNG, WEBP, or GIF ·
                  up to 500 KB.
                </p>
              ) : null}
            </div>
            <p className="max-w-2xl text-sm leading-relaxed text-foreground/75">
              {profile.bio?.trim() || "No bio yet."}
            </p>
            {isSelf && profile.email ? (
              <p className="text-xs text-foreground/75">{profile.email}</p>
            ) : null}
          </div>
        </CardContent>
      </Card>

      {isSelf ? (
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="font-serif text-lg">Edit profile</CardTitle>
            <CardDescription>
              Name and bio are public. Username changes your URL.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              className="space-y-4"
              onSubmit={async (e) => {
                e.preventDefault();
                setBusy(true);
                const body: Record<string, unknown> = { name, bio };
                if (newUsername && newUsername !== me)
                  body.username = newUsername;
                const result = await updateProfile(routeUser, body);
                setBusy(false);
                if (!result.ok) {
                  toastError(apiErrorMessage(result));
                  return;
                }
                if (newUsername && newUsername !== me) {
                  const tokens = {
                    access_token: localStorage.getItem("access_token")!,
                    refresh_token: localStorage.getItem("refresh_token")!,
                    username: newUsername,
                  };
                  acceptTokens(tokens);
                  window.location.href = `/u/${newUsername}`;
                  return;
                }
                refreshLocal();
                toastSuccess("Profile updated.");
                void load();
              }}
            >
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    maxLength={100}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={newUsername}
                    onChange={(e) => setNewUsername(e.target.value)}
                    minLength={3}
                    maxLength={30}
                    required
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="bio">Bio</Label>
                <Textarea
                  id="bio"
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  maxLength={500}
                  placeholder="A short note about what you’re learning."
                />
              </div>
              <Button type="submit" disabled={busy}>
                {busy ? "Saving…" : "Save changes"}
              </Button>
            </form>
          </CardContent>
        </Card>
      ) : (
        <Card className="rounded-2xl">
          <CardHeader>
            <CardTitle className="font-serif text-lg">Report user</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              className="space-y-4"
              onSubmit={async (e) => {
                e.preventDefault();
                setBusy(true);
                const id = Number(profile.id);
                if (!id) {
                  setBusy(false);
                  toastError("Cannot report: user id not available on profile.");
                  return;
                }
                const result = await createReport({
                  target_type: "USER",
                  target_id: id,
                  reason: reportReason,
                });
                setBusy(false);
                if (!result.ok) toastError(apiErrorMessage(result));
                else {
                  setReportReason("");
                  toastSuccess("Report submitted.");
                }
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="reason">Reason</Label>
                <Textarea
                  id="reason"
                  value={reportReason}
                  onChange={(e) => setReportReason(e.target.value)}
                  required
                  minLength={3}
                />
              </div>
              <Button type="submit" disabled={busy}>
                Submit report
              </Button>
            </form>
            {!profile.id ? (
              <div className="mt-3">
                <Empty>
                  Public profile loaded. Reporting needs a numeric user id.
                </Empty>
              </div>
            ) : null}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
