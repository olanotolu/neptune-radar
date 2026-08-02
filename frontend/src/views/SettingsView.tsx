import {
  useDisableUser,
  useDLQ,
  useEnableUser,
  useIngestStatus,
  useRotateAPIKey,
  useUsers,
} from "../api/hooks";
import { useToast } from "../components/Toast";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";
import { isOnboarded, resetOnboarding } from "../components/OnboardingTour";

function SystemInfo() {
  const { data: status } = useIngestStatus();
  const { data: dlq } = useDLQ("pending", 1);
  const dlqCount = dlq?.length ?? 0;

  return (
    <section className="trust-panel">
      <h3 className="trust-panel__title">System</h3>
      <div className="funnel-kpis">
        <div className="funnel-kpi">
          <strong>{status?.provider ?? "—"}</strong>
          <span>Provider</span>
        </div>
        <div className="funnel-kpi">
          <strong>{status?.provider_available ? "Available" : "Down"}</strong>
          <span>Provider status</span>
        </div>
        <div className="funnel-kpi">
          <strong>{status?.poll_interval ?? "—"}</strong>
          <span>Poll interval</span>
        </div>
        <div className="funnel-kpi">
          <strong>{status?.daily_budget ?? "—"}</strong>
          <span>Daily budget</span>
        </div>
        <div className={`funnel-kpi ${dlqCount > 0 ? "kpi--warn" : ""}`}>
          <strong>{dlqCount}</strong>
          <span>DLQ pending</span>
        </div>
      </div>
      {dlqCount > 0 && (
        <p className="ops-view__warn">
          ⚠ {dlqCount} item{dlqCount === 1 ? "" : "s"} in the dead letter queue — review in the DLQ tab.
        </p>
      )}
    </section>
  );
}

export function SettingsView() {
  const { data: users, isLoading, error } = useUsers();
  const rotate = useRotateAPIKey();
  const disable = useDisableUser();
  const enable = useEnableUser();
  const toast = useToast();

  const list = users ?? [];

  return (
    <div className="view view--settings">
      <header className="ops-view__header">
        <div>
          <h2 className="ops-view__title">Admin</h2>
          <p className="ops-view__sub">Team access, API keys, and system configuration.</p>
        </div>
      </header>

      <SystemInfo />

      <section className="trust-panel">
        <h3 className="trust-panel__title">Onboarding</h3>
        <p className="ops-view__sub" style={{ marginBottom: 12 }}>
          {isOnboarded()
            ? "You've completed the welcome tour."
            : "The welcome tour is available."}
        </p>
        <button
          type="button"
          className="btn btn--ghost btn--sm"
          onClick={() => {
            resetOnboarding();
            window.location.hash = "#/today";
            window.location.reload();
          }}
        >
          {isOnboarded() ? "Restart tour" : "Take the tour"}
        </button>
      </section>

      <section className="trust-panel">
        <h3 className="trust-panel__title">Users</h3>
        {isLoading && <LoadingState variant="skeleton" message="Loading users…" />}
        {error && <EmptyState variant="warning" icon="⚠" title="Users unavailable" message={(error as Error).message} />}
        {!isLoading && !error && list.length === 0 && (
          <EmptyState variant="empty" icon="👥" title="No users found" message="Operator accounts will appear here once created." />
        )}
        {!isLoading && !error && list.length > 0 && (
          <table className="dossier-ledger">
            <thead>
              <tr>
                <th>Email</th>
                <th>Role</th>
                <th>Last seen</th>
                <th>API key</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {list.map((u) => {
                const busy = rotate.isPending || disable.isPending || enable.isPending;
                return (
                  <tr key={u.id}>
                    <td>{u.email}</td>
                    <td>{u.role}</td>
                    <td>{u.last_seen_at?.slice(0, 16)?.replace("T", " ") ?? "—"}</td>
                    <td><code>{u.api_key_masked ?? "—"}</code></td>
                    <td className="ops-view__actions">
                      <button
                        type="button"
                        className="btn"
                        disabled={busy}
                        onClick={() =>
                          rotate.mutate(u.id, {
                            onSuccess: () => toast.push("API key rotated", "ok"),
                            onError: (e) => toast.push((e as Error).message, "err"),
                          })
                        }
                      >
                        Rotate key
                      </button>
                      {u.disabled ? (
                        <button
                          type="button"
                          className="btn btn--primary"
                          disabled={busy}
                          onClick={() =>
                            enable.mutate(u.id, {
                              onSuccess: () => toast.push("User enabled", "ok"),
                              onError: (e) => toast.push((e as Error).message, "err"),
                            })
                          }
                        >
                          Enable
                        </button>
                      ) : (
                        <button
                          type="button"
                          className="btn"
                          disabled={busy}
                          onClick={() =>
                            disable.mutate(u.id, {
                              onSuccess: () => toast.push("User disabled", "ok"),
                              onError: (e) => toast.push((e as Error).message, "err"),
                            })
                          }
                        >
                          Disable
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
