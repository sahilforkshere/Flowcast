"use client";

import { useEffect, useState } from "react";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useWebSocket } from "@/hooks/useWebSocket";

function StatCard({
  label,
  value,
  sub,
}: {
  label: string;
  value: string | number;
  sub?: string;
}) {
  return (
    <div className="rounded-xl border border-slate-700/60 bg-slate-800/50 p-6">
      <p className="text-xs font-medium uppercase tracking-widest text-slate-400">
        {label}
      </p>
      <p className="mt-2 text-4xl font-bold text-white">{value}</p>
      {sub && <p className="mt-1 text-xs text-slate-500">{sub}</p>}
    </div>
  );
}

function SectionHeader({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="mb-4">
      <h2 className="text-base font-semibold text-slate-100">{title}</h2>
      <p className="text-xs text-slate-500">{desc}</p>
    </div>
  );
}

export default function Home() {
  const { stats, connected } = useWebSocket("ws://localhost:8081/ws");
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (connected) {
      const id = setInterval(() => setTick((t) => t + 1), 2000);
      return () => clearInterval(id);
    }
  }, [connected]);

  const lastUpdated = connected
    ? new Date().toLocaleTimeString()
    : "—";

  return (
    <main className="min-h-screen bg-[#0f1117] px-6 py-8 text-slate-200">
      <div className="mx-auto max-w-6xl space-y-8">

        {/* Header */}
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-xl font-semibold tracking-tight text-white">
              Flowcast Analytics
            </h1>
            <p className="mt-0.5 text-xs text-slate-500">
              Real-time page-view pipeline — updates every 2 seconds
            </p>
          </div>
          <div className="flex flex-col items-end gap-1">
            <span
              className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ${
                connected
                  ? "bg-emerald-900/60 text-emerald-400"
                  : "bg-red-900/60 text-red-400"
              }`}
            >
              <span
                className={`h-1.5 w-1.5 rounded-full ${
                  connected ? "bg-emerald-400 animate-pulse" : "bg-red-400"
                }`}
              />
              {connected ? "Live" : "Connecting..."}
            </span>
            <span className="text-[11px] text-slate-600">
              Last update: {lastUpdated}
            </span>
          </div>
        </div>

        {/* Stat cards */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <StatCard
            label="Active Users"
            value={stats.active_users}
            sub="Unique sessions — last 5 minutes"
          />
          <StatCard
            label="Top Page Views"
            value={
              stats.top_pages && stats.top_pages.length > 0
                ? stats.top_pages[0].views.toLocaleString()
                : "—"
            }
            sub={
              stats.top_pages && stats.top_pages.length > 0
                ? stats.top_pages[0].page_url
                : "No data yet"
            }
          />
          <StatCard
            label="Data Points"
            value={
              stats.views_per_minute
                ? stats.views_per_minute
                    .reduce((s, v) => s + v.views, 0)
                    .toLocaleString()
                : "—"
            }
            sub="Views recorded — last 30 minutes"
          />
        </div>

        {/* Views per minute chart */}
        <div className="rounded-xl border border-slate-700/60 bg-slate-800/50 p-6">
          <SectionHeader
            title="Views per Minute"
            desc="Page-view volume over the last 30 minutes"
          />
          <ResponsiveContainer width="100%" height={220}>
            <LineChart
              data={stats.views_per_minute ?? []}
              margin={{ top: 4, right: 8, bottom: 0, left: -10 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
              <XAxis
                dataKey="minute"
                tick={{ fill: "#64748b", fontSize: 11 }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                tick={{ fill: "#64748b", fontSize: 11 }}
                axisLine={false}
                tickLine={false}
              />
              <Tooltip
                contentStyle={{
                  background: "#1e293b",
                  border: "1px solid #334155",
                  borderRadius: 8,
                  color: "#e2e8f0",
                  fontSize: 12,
                }}
              />
              <Line
                type="monotone"
                dataKey="views"
                stroke="#6366f1"
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4, fill: "#6366f1" }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Top pages chart */}
        <div className="rounded-xl border border-slate-700/60 bg-slate-800/50 p-6">
          <SectionHeader
            title="Top Pages"
            desc="Most visited pages in the last hour"
          />
          <ResponsiveContainer width="100%" height={220}>
            <BarChart
              data={stats.top_pages ?? []}
              layout="vertical"
              margin={{ top: 4, right: 24, bottom: 0, left: 16 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" horizontal={false} />
              <XAxis
                type="number"
                tick={{ fill: "#64748b", fontSize: 11 }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                type="category"
                dataKey="page_url"
                tick={{ fill: "#94a3b8", fontSize: 12 }}
                axisLine={false}
                tickLine={false}
                width={80}
              />
              <Tooltip
                contentStyle={{
                  background: "#1e293b",
                  border: "1px solid #334155",
                  borderRadius: 8,
                  color: "#e2e8f0",
                  fontSize: 12,
                }}
              />
              <Bar dataKey="views" fill="#6366f1" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-slate-700">
          Flowcast — Go + Kafka + ClickHouse + Next.js
        </p>
      </div>
    </main>
  );
}
