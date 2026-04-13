import type { SnapshotPayload, StreamEnvelope } from "@/types";

export interface ParsedMessage {
  eventType: StreamEnvelope["event_type"];
  snapshot?: SnapshotPayload;
  heartbeatAt?: string;
  error?: string;
}

export function parseTelemetryMessage(raw: string): ParsedMessage {
  const env = JSON.parse(raw) as StreamEnvelope;
  if (env.schema_version !== "v1") {
    return {
      eventType: "error",
      error: `Unsupported schema version: ${env.schema_version}`,
    };
  }

  if (env.event_type === "snapshot" || env.event_type === "delta_update") {
    return {
      eventType: env.event_type,
      snapshot: env.payload as SnapshotPayload,
    };
  }

  if (env.event_type === "heartbeat") {
    return {
      eventType: env.event_type,
      heartbeatAt: env.generated_at,
    };
  }

  return {
    eventType: "error",
    error: typeof env.payload === "string" ? env.payload : "Telemetry stream error",
  };
}

