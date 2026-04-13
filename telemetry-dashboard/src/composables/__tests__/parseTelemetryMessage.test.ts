import { describe, expect, it } from "vitest";
import { parseTelemetryMessage } from "@/composables/parseTelemetryMessage";

describe("parseTelemetryMessage", () => {
  it("parses snapshot payload", () => {
    const parsed = parseTelemetryMessage(
      JSON.stringify({
        schema_version: "v1",
        event_type: "snapshot",
        generated_at: "2026-01-01T00:00:00Z",
        payload: {
          schema_version: "v1",
          generated_at: "2026-01-01T00:00:00Z",
          windows: {},
          consumer_lag: 2,
          recent_points: [],
        },
      }),
    );
    expect(parsed.eventType).toBe("snapshot");
    expect(parsed.snapshot?.consumer_lag).toBe(2);
  });

  it("returns error on schema mismatch", () => {
    const parsed = parseTelemetryMessage(
      JSON.stringify({
        schema_version: "v9",
        event_type: "snapshot",
        generated_at: "2026-01-01T00:00:00Z",
        payload: {},
      }),
    );
    expect(parsed.eventType).toBe("error");
  });
});

