import { useState } from "react";
import { Play, Lightning, TestTube } from "@phosphor-icons/react";
import { apiFetch } from "../lib/api";
import {
  PLAYGROUNDS,
  hasRaceHelper,
  type PlaygroundEndpoint,
} from "../services/playground";

type Result = {
  label: string;
  status: number;
  body: string;
};

function buildPath(endpoint: PlaygroundEndpoint, values: Record<string, string>): string {
  let path = `targets/${endpoint.path}`;
  for (const param of endpoint.params ?? []) {
    const value = encodeURIComponent(values[param.name] ?? "");
    if (path.includes(`:${param.name}`)) {
      path = path.replace(`:${param.name}`, value);
    } else {
      path += (path.includes("?") ? "&" : "?") + `${param.name}=${value}`;
    }
  }
  return path;
}

function fillBody(
  template: Record<string, string> | undefined,
  values: Record<string, string>,
): string | undefined {
  if (!template) return undefined;
  const filled: Record<string, string> = {};
  for (const [key, raw] of Object.entries(template)) {
    filled[key] = raw.replace(/%PARAM:(\w+)%/g, (_, name) => values[name] ?? "");
  }
  return JSON.stringify(filled);
}

export function LabPlayground({ slug }: { slug: string }) {
  const endpoints = PLAYGROUNDS[slug];
  const [values, setValues] = useState<Record<string, Record<string, string>>>({});
  const [results, setResults] = useState<Record<string, Result | null>>({});
  const [busy, setBusy] = useState<string | null>(null);

  const raceHelper = hasRaceHelper(slug);

  async function run(key: string, endpoint: PlaygroundEndpoint) {
    const paramValues = values[key] ?? {};
    setBusy(key);
    setResults((prev) => ({ ...prev, [key]: null }));
    try {
      const path = buildPath(endpoint, paramValues);
      const headers: Record<string, string> = {};
      const badge = paramValues["badge"];
      if (badge) headers["X-Lab-Badge"] = `Bearer ${badge}`;
      // Badge goes via header, not query — keep it out of URLs.
      const cleanPath = badge
        ? path.replace(/([?&])badge=[^&]*/, "$1").replace(/[?&]$/, "")
        : path;

      let payload: unknown;
      try {
        payload = await apiFetch<unknown>(`/${cleanPath}`, {
          method: endpoint.method,
          body: fillBody(endpoint.bodyTemplate, paramValues),
          headers,
        });
      } catch (cause) {
        const status = (cause as { status?: number }).status ?? 0;
        const message = (cause as Error).message || "Request failed.";
        setResults((prev) => ({
          ...prev,
          [key]: { label: endpoint.label, status, body: message },
        }));
        return;
      }
      setResults((prev) => ({
        ...prev,
        [key]: {
          label: endpoint.label,
          status: 200,
          body: JSON.stringify(payload, null, 2),
        },
      }));
    } finally {
      setBusy(null);
    }
  }

  /** Fire N parallel vulnerable redemptions and summarise the outcome. */
  async function runRaceHelper(count: number) {
    setBusy("race");
    setResults((prev) => ({ ...prev, race: null }));
    try {
      const attempts = await Promise.allSettled(
        Array.from({ length: count }, () =>
          apiFetch<{ redeemed: boolean }>("/targets/race-condition/redeem", {
            method: "POST",
            body: { code: "LAUNCH-2026" },
          }),
        ),
      );
      const ok = attempts.filter((a) => a.status === "fulfilled").length;
      let coupon: { max_redemptions: number; redeemed: number; mine: boolean };
      try {
        coupon = await apiFetch<{ coupon: typeof coupon }>(
          "/targets/race-condition/coupon",
        ).then((r) => r.coupon);
      } catch {
        setResults((prev) => ({
          ...prev,
          race: {
            label: "Race helper",
            status: 0,
            body: `${ok}/${count} requests succeeded, but coupon state could not be read.`,
          },
        }));
        return;
      }
      const over = coupon.redeemed > coupon.max_redemptions;
      setResults((prev) => ({
        ...prev,
        race: {
          label: "Race helper",
          status: over ? 200 : 409,
          body: [
            `${ok}/${count} concurrent redemptions succeeded.`,
            `Coupon state: ${coupon.redeemed}/${coupon.max_redemptions} redeemed.`,
            over
              ? "Counter pushed past the cap — race condition exploited!"
              : "Cap not exceeded yet — try firing again with more requests.",
            coupon.mine ? "You hold at least one redemption." : "You hold no redemption.",
          ].join("\n"),
        },
      }));
    } finally {
      setBusy(null);
    }
  }

  if (!endpoints && !raceHelper) return null;

  return (
    <section className="mt-6 rounded-xl border border-ink-800 bg-ink-900 p-6 sm:p-8">
      <h2 className="flex items-center gap-2.5 font-display text-lg font-semibold tracking-tight text-fog-100">
        <TestTube size={20} className="text-signal-500" aria-hidden />
        Playground
      </h2>
      <p className="mt-2 text-sm text-fog-300">
        Drive this lab's endpoints without leaving the page. Requests use your
        current session.
      </p>

      {raceHelper && (
        <div className="mt-5 rounded-lg border border-caution-500/30 bg-caution-500/5 p-4">
          <p className="text-sm leading-relaxed text-fog-200">
            This lab needs <strong>concurrent</strong> requests to win the race.
            Fire a burst of parallel redemptions:
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            {[12, 24].map((count) => (
              <button
                key={count}
                onClick={() => runRaceHelper(count)}
                disabled={busy !== null}
                className="inline-flex items-center gap-2 rounded-lg border border-signal-500/50 bg-signal-500/10 px-4 py-2 text-sm font-medium text-signal-500 transition-colors hover:bg-signal-500/20 disabled:opacity-50"
              >
                <Lightning size={15} weight="fill" aria-hidden />
                {busy === "race" ? "Firing…" : `Fire ${count} redemptions`}
              </button>
            ))}
          </div>
          <ResultView result={results["race"] ?? null} />
        </div>
      )}

      <div className="mt-5 flex flex-col gap-4">
        {(endpoints ?? []).map((endpoint) => {
          const key = endpoint.label;
          const paramValues = values[key] ?? {};
          return (
            <div key={key} className="rounded-lg border border-ink-700 p-4">
              <div className="flex items-baseline justify-between gap-3">
                <p className="font-mono text-sm font-medium text-fog-100">
                  <span className="mr-2 rounded bg-ink-800 px-1.5 py-0.5 text-xs uppercase text-fog-300">
                    {endpoint.method}
                  </span>
                  /targets/{slug}/{endpoint.path}
                </p>
              </div>
              <p className="mt-1 text-sm text-fog-400">{endpoint.description}</p>

              {(endpoint.params?.length ?? 0) > 0 && (
                <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {endpoint.params!.map((param) => (
                    <label key={param.name} className="flex flex-col gap-1">
                      <span className="text-xs font-medium text-fog-300">{param.label}</span>
                      <input
                        value={paramValues[param.name] ?? ""}
                        placeholder={param.placeholder}
                        onChange={(e) =>
                          setValues((prev) => ({
                            ...prev,
                            [key]: { ...prev[key], [param.name]: e.target.value },
                          }))
                        }
                        className="h-9 rounded-lg border border-ink-600 bg-ink-850 px-3 font-mono text-sm text-fog-100 placeholder:text-fog-400/60 focus:border-signal-500"
                      />
                    </label>
                  ))}
                </div>
              )}

              <button
                onClick={() => run(key, endpoint)}
                disabled={busy !== null}
                className="mt-3 inline-flex items-center gap-2 rounded-lg border border-ink-600 px-4 py-2 text-sm font-medium text-fog-200 transition-colors hover:border-fog-400 hover:text-fog-100 disabled:opacity-50"
              >
                <Play size={14} weight="fill" aria-hidden />
                {busy === key ? "Sending…" : "Send"}
              </button>

              <ResultView result={results[key] ?? null} />
            </div>
          );
        })}
      </div>
    </section>
  );
}

function ResultView({ result }: { result: Result | null }) {
  if (!result) return null;
  return (
    <pre
      aria-live="polite"
      className={`mt-3 max-h-64 overflow-auto whitespace-pre-wrap break-all rounded-lg border px-3 py-2 font-mono text-xs leading-relaxed ${
        result.status >= 200 && result.status < 300
          ? "border-signal-500/40 bg-signal-500/10 text-signal-500"
          : result.status === 0
            ? "border-alert-500/40 bg-alert-500/10 text-alert-500"
            : "border-caution-500/40 bg-caution-500/10 text-caution-500"
      }`}
    >
      {`[${result.status}] ${result.body}`}
    </pre>
  );
}
