import { motion } from "framer-motion";
import type { ReactNode } from "react";
import type { Download } from "../types";
import { api } from "../api/client";
import { formatBytes, formatSpeed, formatETA, progress } from "../lib/format";

const statusColor: Record<string, string> = {
  running: "text-pounce-glow",
  paused: "text-yellow-400",
  completed: "text-green-400",
  error: "text-red-400",
  queued: "text-white/50",
  canceled: "text-white/30",
};

const cardInitial = { opacity: 0, y: 8 };
const cardAnimate = { opacity: 1, y: 0 };

export function DownloadCard({ d }: { d: Download }) {
  const pct = progress(d);
  const barStyle = { width: `${pct}%` };

  return (
    <motion.div
      layout
      initial={cardInitial}
      animate={cardAnimate}
      className="glass rounded-2xl p-4"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="truncate font-medium">{d.filename}</p>
          <p className="truncate text-xs text-white/40">{d.url}</p>
        </div>
        <span
          className={`shrink-0 text-xs font-semibold uppercase ${statusColor[d.status] ?? ""}`}
        >
          {d.status}
        </span>
      </div>

      <div className="mt-3 h-2 overflow-hidden rounded-full bg-black/40">
        <div
          className="h-full rounded-full bg-gradient-to-r from-pounce-accent to-pounce-glow transition-all"
          style={barStyle}
        />
      </div>

      <div className="mt-2 flex flex-wrap items-center justify-between gap-2 text-xs text-white/50">
        <span>
          {formatBytes(d.downloaded)}
          {d.totalSize > 0
            ? ` / ${formatBytes(d.totalSize)} (${pct.toFixed(0)}%)`
            : ""}
        </span>
        <span className="flex gap-3">
          <span>{formatSpeed(d.speed)}</span>
          <span>ETA {formatETA(d)}</span>
          <span>
            {d.connections}× {d.resumable ? "resumable" : "single"}
          </span>
        </span>
      </div>

      {d.error && <p className="mt-2 text-xs text-red-400">{d.error}</p>}

      <div className="mt-3 flex gap-2">
        {d.status === "running" ? (
          <Action onClick={() => api.pause(d.id)}>Pause</Action>
        ) : d.status === "paused" || d.status === "error" ? (
          <Action onClick={() => api.resume(d.id)}>Resume</Action>
        ) : null}
        {d.status !== "completed" && (
          <Action onClick={() => api.cancel(d.id)}>Cancel</Action>
        )}
        {d.status === "completed" && (
          <Action onClick={() => api.remove(d.id)}>Clear</Action>
        )}
      </div>
    </motion.div>
  );
}

function Action({
  children,
  onClick,
}: {
  children: ReactNode;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="rounded-lg bg-white/5 px-3 py-1.5 text-xs font-medium text-white/70 transition hover:bg-white/15 hover:text-white"
    >
      {children}
    </button>
  );
}
