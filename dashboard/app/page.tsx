"use client";

import { useEffect } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";

export default function Home() {
  const { stats, connected } = useWebSocket("ws://localhost:8081/ws");

  useEffect(() => {
    console.log("[dashboard] stats update:", stats);
  }, [stats]);

  return (
    <main className="min-h-screen p-8">
      <div className="max-w-6xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold tracking-tight">
            Flowcast Analytics
          </h1>
          <span
            className={`text-sm px-3 py-1 rounded-full ${
              connected
                ? "bg-green-900 text-green-300"
                : "bg-red-900 text-red-300"
            }`}
          >
            {connected ? "Live" : "Connecting..."}
          </span>
        </div>

        <div className="rounded-lg border border-slate-700 p-6 text-center">
          <p className="text-slate-400 text-sm mb-1">Active Users (last 5 min)</p>
          <p className="text-5xl font-bold text-white">{stats.active_users}</p>
        </div>

        <p className="mt-8 text-slate-500 text-sm text-center">
          Charts coming in Day 3 — open browser console to see live WebSocket data
        </p>
      </div>
    </main>
  );
}
