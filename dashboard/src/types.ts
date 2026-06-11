// Mirrors the JSON emitted by the Go engine (internal/model + internal/settings).
export type Status =
  | "queued"
  | "running"
  | "paused"
  | "completed"
  | "error"
  | "canceled";

export interface Segment {
  index: number;
  start: number;
  end: number;
  downloaded: number;
}

export interface Download {
  id: string;
  url: string;
  filename: string;
  dir: string;
  totalSize: number;
  downloaded: number;
  status: Status;
  connections: number;
  resumable: boolean;
  segments: Segment[];
  speedLimit: number;
  error?: string;
  createdAt: string;
  updatedAt: string;
  speed: number;
}

// User preferences, persisted by the engine (GET/PUT /api/settings).
export interface Settings {
  theme: string;
  defaultDir: string;
  defaultConnections: number;
  defaultSpeedLimit: number;
  maxConcurrent: number;
  notifyOnComplete: boolean;
}
