import { useEffect, useRef, useState } from "react";
import type { ReactNode } from "react";
import { useAuth } from "../hooks/useAuth";

const DEMO_CREDENTIALS = {
  email: "demo@hyperion.test",
  password: "Password123!",
};

const FALLBACK_CREDENTIALS = {
  email: "operator@hyperion.local",
  password: "password12345",
};

export function AutoAuth({ children }: { children: ReactNode }) {
  const { status, login, register } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const isAttempting = useRef(false);

  useEffect(() => {
    if (status !== "unauthenticated" || isAttempting.current) return;

    isAttempting.current = true;
    setError(null);

    async function authenticate() {
      try {
        await login(DEMO_CREDENTIALS.email, DEMO_CREDENTIALS.password);
        isAttempting.current = false;
        return;
      } catch {
        // If login failed, the demo user might not exist yet in DB; try registering it
        try {
          await register(DEMO_CREDENTIALS.email, DEMO_CREDENTIALS.password);
          isAttempting.current = false;
          return;
        } catch {
          // Fallback to secondary operator credentials
          try {
            await login(FALLBACK_CREDENTIALS.email, FALLBACK_CREDENTIALS.password);
            isAttempting.current = false;
            return;
          } catch {
            try {
              await register(FALLBACK_CREDENTIALS.email, FALLBACK_CREDENTIALS.password);
              isAttempting.current = false;
              return;
            } catch (err) {
              console.error("AutoAuth failed to initialize operator session:", err);
              isAttempting.current = false;
              setError("Could not connect to Hyperion backend service. Please check that the server is running.");
            }
          }
        }
      }
    }

    authenticate();
  }, [status, login, register]);

  if (status === "authenticated") {
    return <>{children}</>;
  }

  if (error) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center font-mono text-[#8a8a6f]">
        <div className="w-full max-w-md border border-[#ff5555] bg-[#161614] p-8 text-center">
          <div className="mb-3 text-[11px] uppercase tracking-[0.1em] text-[#ff5555]">
            // BACKEND CONNECTION ERROR
          </div>
          <p className="mb-6 text-body-sm text-[#eeeeee] leading-relaxed">{error}</p>
          <button
            type="button"
            onClick={() => {
              isAttempting.current = false;
              setError(null);
              window.location.reload();
            }}
            className="inline-flex h-10 items-center justify-center border border-[#ffa133] bg-[#ffa133] px-6 text-caption font-medium uppercase tracking-[0.06em] text-black transition-colors hover:bg-[#e47b1a]"
          >
            RETRY CONNECTION
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center font-mono text-[#8a8a6f]">
      <div className="w-full max-w-md border border-[#2a2a26] bg-[#0d0d0b] p-8 sm:p-10 text-center">
        <div className="mb-3 text-[11px] uppercase tracking-[0.1em] text-[#ffa133]">
          // INITIALIZING HYPERION
        </div>
        <div className="flex items-center justify-center gap-3 text-body-sm text-[#eeeeee]">
          <span className="size-2 animate-pulse bg-[#ffa133]" aria-hidden />
          ESTABLISHING OPERATOR SESSION...
        </div>
        <div className="mt-4 text-xs tracking-[0.02em] text-[#777766]">
          Zero-radius security testing sandbox
        </div>
      </div>
    </div>
  );
}
