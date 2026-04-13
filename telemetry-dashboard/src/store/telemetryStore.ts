import { defineStore } from "pinia";
import type { ConnectionState, SnapshotPayload } from "@/types";

export const useTelemetryStore = defineStore("telemetry", {
  state: () => ({
    snapshot: null as SnapshotPayload | null,
    connection: "disconnected" as ConnectionState,
    lastHeartbeatAt: "" as string,
    lastError: "" as string,
  }),
  actions: {
    setSnapshot(snapshot: SnapshotPayload) {
      this.snapshot = snapshot;
    },
    setConnection(connection: ConnectionState) {
      this.connection = connection;
    },
    setHeartbeat(ts: string) {
      this.lastHeartbeatAt = ts;
    },
    setError(message: string) {
      this.lastError = message;
    },
  },
});

