export type ConnectionState = "connected" | "reconnecting" | "disconnected";

export interface WindowMetrics {
  event_count: number;
  error_count: number;
  throughput: number;
  error_rate: number;
  latency_ms: {
    p50: number;
    p95: number;
    p99: number;
  };
  window_label: string;
}

export interface TimeseriesPoint {
  timestamp: string;
  throughput: number;
  error_rate: number;
  latency_p95: number;
  consumer_lag: number;
}

export interface SnapshotPayload {
  schema_version: string;
  generated_at: string;
  windows: Record<string, WindowMetrics>;
  consumer_lag: number;
  recent_points: TimeseriesPoint[];
}

export interface StreamEnvelope {
  schema_version: string;
  event_type: "snapshot" | "delta_update" | "heartbeat" | "error";
  generated_at: string;
  payload: unknown;
}

