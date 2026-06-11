export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const v = bytes / Math.pow(1024, i);
  return `${v.toFixed(v >= 100 || i === 0 ? 0 : 1)} ${units[i]}`;
}

export function formatSpeed(bytesPerSec: number): string {
  if (!bytesPerSec) return "—";
  return `${formatBytes(bytesPerSec)}/s`;
}

export function formatETA(d: {
  totalSize: number;
  downloaded: number;
  speed: number;
}): string {
  if (!d.speed || d.totalSize <= 0) return "—";
  const remaining = d.totalSize - d.downloaded;
  if (remaining <= 0) return "done";
  const secs = Math.round(remaining / d.speed);
  if (secs < 60) return `${secs}s`;
  if (secs < 3600) return `${Math.floor(secs / 60)}m ${secs % 60}s`;
  return `${Math.floor(secs / 3600)}h ${Math.floor((secs % 3600) / 60)}m`;
}

export function progress(d: { totalSize: number; downloaded: number }): number {
  if (d.totalSize <= 0) return 0;
  return Math.min(100, (d.downloaded / d.totalSize) * 100);
}
