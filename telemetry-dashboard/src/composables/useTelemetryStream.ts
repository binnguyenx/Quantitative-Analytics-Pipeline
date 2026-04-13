import { onMounted, onUnmounted } from "vue";
import { useTelemetryStore } from "@/store/telemetryStore";
import { parseTelemetryMessage } from "@/composables/parseTelemetryMessage";
import type { SnapshotPayload } from "@/types";

const STREAM_URL = "/telemetry/stream";
const SNAPSHOT_URL = "/telemetry/snapshot";

export function useTelemetryStream() {
  const store = useTelemetryStore();
  let es: EventSource | null = null;
  let reconnectTimer: number | null = null;
  let pollingTimer: number | null = null;

  const applySnapshot = (snapshot: SnapshotPayload) => {
    store.setSnapshot(snapshot);
  };

  const fetchSnapshot = async () => {
    try {
      const res = await fetch(SNAPSHOT_URL);
      if (!res.ok) {
        throw new Error(`snapshot request failed: ${res.status}`);
      }
      applySnapshot((await res.json()) as SnapshotPayload);
    } catch (err) {
      store.setError((err as Error).message);
    }
  };

  const startPollingFallback = () => {
    if (pollingTimer !== null) {
      return;
    }
    pollingTimer = window.setInterval(fetchSnapshot, 5000);
  };

  const stopPollingFallback = () => {
    if (pollingTimer !== null) {
      clearInterval(pollingTimer);
      pollingTimer = null;
    }
  };

  const connect = () => {
    stopPollingFallback();
    store.setConnection("reconnecting");
    es = new EventSource(STREAM_URL);

    es.addEventListener("snapshot", (ev: MessageEvent) => {
      const parsed = parseTelemetryMessage(ev.data);
      if (parsed.snapshot) {
        applySnapshot(parsed.snapshot);
      }
      store.setConnection("connected");
    });

    es.addEventListener("delta_update", (ev: MessageEvent) => {
      const parsed = parseTelemetryMessage(ev.data);
      if (parsed.snapshot) {
        applySnapshot(parsed.snapshot);
      }
      store.setConnection("connected");
    });

    es.addEventListener("heartbeat", (ev: MessageEvent) => {
      const parsed = parseTelemetryMessage(ev.data);
      store.setConnection("connected");
      if (parsed.heartbeatAt) {
        store.setHeartbeat(parsed.heartbeatAt);
      }
    });

    es.onerror = () => {
      store.setConnection("disconnected");
      es?.close();
      startPollingFallback();
      reconnectTimer = window.setTimeout(connect, 3000);
    };
  };

  onMounted(async () => {
    await fetchSnapshot();
    connect();
  });

  onUnmounted(() => {
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer);
    }
    stopPollingFallback();
    es?.close();
  });
}

