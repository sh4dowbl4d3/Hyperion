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
    <section className="mt-6 rounded-card bg-pure-white p-7 sm:p-9">
      <h2 className="flex items-center gap-2.5 text-[22px] font-medium tracking-tight text-ink-black">
        <TestTube size={22} weight="duotone" className="text-sky-pop" aria-hidden />
        Playground
      </h2>
      <p className="mt-2 text-[15px] text-stone-gray">
        Drive this lab's endpoints without leaving the page. Requests use your
        current session.
      </p>

      {raceHelper && (
        <div className="mt-5 rounded-card bg-cream-paper p-5">
          <p className="text-sm leading-relaxed text-ink-black/85">
            This lab needs <strong>concurrent</strong> requests to win the race.
            Fire a burst of parallel redemptions:
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            {[12, 24].map((count) => (
              <button
                key={count}
                onClick={() => runRaceHelper(count)}
                disabled={busy !== null}
                className="inline-flex items-center gap-2 rounded-pill bg-coral-pop px-4 py-2 text-sm font-medium text-white transition-all hover:brightness-105 active:translate-y-[1px] disabled:opacity-50"
              >
                <Lightning size={14} weight="fill" aria-hidden />
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
            <div key={key} className="rounded-card bg-cream-paper p-5">
              <p className="font-mono text-sm text-ink-black">
                <span className="mr-2 rounded-[10px] bg-pure-white px-2 py-0.5 text-xs uppercase text-stone-gray">
                  {endpoint.method}
                </span>
                /targets/{slug}/{endpoint.path}
              </p>
              <p className="mt-1.5 text-sm text-stone-gray">{endpoint.description}</p>

              {(endpoint.params?.length ?? 0) > 0 && (
                <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {endpoint.params!.map((param) => (
                    <label key={param.name} className="flex flex-col gap-1">
                      <span className="text-xs font-medium text-ink-black/80">{param.label}</span>
                      <input
                        value={paramValues[param.name] ?? ""}
                        placeholder={param.placeholder}
                        onChange={(e) =>
                          setValues((prev) => ({
                            ...prev,
                            [key]: { ...prev[key], [param.name]: e.target.value },
                          }))
                        }
                        className="h-10 rounded-pill border border-hairline-mist bg-pure-white px-4 font-mono text-sm text-ink-black placeholder:text-stone-gray focus:border-ink-black"
                      />
                    </label>
                  ))}
                </div>
              )}

              <button
                onClick={() => run(key, endpoint)}
                disabled={busy !== null}
                className="mt-3 inline-flex items-center gap-2 rounded-pill bg-fresh-grass px-4 py-2 text-sm font-medium text-ink-black transition-all hover:brightness-105 active:translate-y-[1px] disabled:opacity-50"
              >
                <Play size={13} weight="fill" aria-hidden />
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
      className={`mt-3 max-h-64 overflow-auto whitespace-pre-wrap break-all rounded-card bg-pure-white px-4 py-3 font-mono text-xs leading-relaxed ${
        result.status >= 200 && result.status < 300
          ? "border border-fresh-grass text-ink-black"
          : result.status === 0
            ? "border border-coral-pop text-coral-pop"
            : "border border-hairline-mist text-stone-gray"
      }`}
    >
      {`[${result.status}] ${result.body}`}
    </pre>
  );
}
