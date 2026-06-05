"use client";

import { useEffect, useRef, useState } from "react";

export interface TopPage {
  page_url: string;
  views: number;
}

export interface ViewsPerMinute {
  minute: string;
  views: number;
}

export interface DashboardStats {
  active_users: number;
  top_pages: TopPage[];
  views_per_minute: ViewsPerMinute[];
}

const EMPTY: DashboardStats = {
  active_users: 0,
  top_pages: [],
  views_per_minute: [],
};

export function useWebSocket(url: string) {
  const [stats, setStats] = useState<DashboardStats>(EMPTY);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    function connect() {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        console.log("[ws] connected");
      };

      ws.onmessage = (e) => {
        try {
          const data: DashboardStats = JSON.parse(e.data);
          setStats(data);
        } catch {
          console.error("[ws] bad JSON", e.data);
        }
      };

      ws.onclose = () => {
        setConnected(false);
        console.log("[ws] disconnected, retrying in 3s");
        retryRef.current = setTimeout(connect, 3000);
      };

      ws.onerror = () => ws.close();
    }

    connect();

    return () => {
      wsRef.current?.close();
      if (retryRef.current) clearTimeout(retryRef.current);
    };
  }, [url]);

  return { stats, connected };
}
