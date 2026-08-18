"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { mountDevTester } from "./legacy-main";
import "@/styles/dev-tester.css";

export default function DevPage() {
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!rootRef.current) return;
    return mountDevTester(rootRef.current);
  }, []);

  return (
    <>
      <div
        style={{
          padding: "0.5rem 1rem",
          background: "#1a2332",
          borderBottom: "1px solid #2d3a4f",
          display: "flex",
          gap: "1rem",
          alignItems: "center",
        }}
      >
        <Link href="/">← Taggy app</Link>
        <strong>API Tester</strong>
        <span style={{ color: "#8b9bb4", fontSize: "0.85rem" }}>
          /dev — shares the same localStorage tokens as the app
        </span>
      </div>
      <div ref={rootRef} />
    </>
  );
}
