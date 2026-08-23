import type { ReactNode } from "react";

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <main className="grid min-h-[100dvh] grid-cols-1 lg:grid-cols-2">
      {/* Editorial side panel on cream canvas */}
      <section className="relative hidden flex-col justify-between overflow-hidden border-r border-hairline-mist bg-cream-paper p-12 lg:flex xl:p-16">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-50"
          style={{
            backgroundImage:
              "linear-gradient(rgba(44,46,42,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(44,46,42,0.04) 1px, transparent 1px)",
            backgroundSize: "44px 44px",
          }}
        />
        <div className="relative flex items-center gap-2.5">
          <span
            aria-hidden
            className="flex size-9 items-center justify-center rounded-[10px] bg-fresh-grass"
          >
            <span className="size-2.5 rounded-full bg-ink-black" />
          </span>
          <span className="text-[17px] font-medium tracking-tight text-ink-black">
            ModernDVWA
          </span>
        </div>

        {/* Oversized editorial headline */}
        <div className="relative max-w-xl">
          <h1 className="text-heading-sm font-medium leading-tight tracking-tight text-ink-black sm:text-[53px] sm:leading-[1.05]">
            Break modern web apps.
            <br />
            <span className="italic">Learn</span> how to defend them.
          </h1>
          <p className="mt-6 max-w-[46ch] text-body-lg font-normal leading-relaxed text-stone-gray">
            Hands-on labs for the vulnerabilities that matter in today's APIs,
            authentication flows and business logic — running entirely on your
            machine against synthetic data.
          </p>
        </div>

        {/* Paper-cut decorative shapes */}
        <div aria-hidden className="relative flex items-center gap-3">
          <span className="size-8 rounded-full bg-sky-pop" />
          <span className="h-8 w-14 rounded-pill bg-coral-pop" />
          <span className="size-8 rounded-[10px] bg-fresh-grass" />
        </div>
      </section>

      {/* Form column on white */}
      <section className="flex items-center justify-center bg-pure-white px-5 py-14 sm:px-10 lg:bg-cream-paper">
        <div className="w-full max-w-md rounded-card bg-pure-white p-8 sm:p-10">{children}</div>
      </section>
    </main>
  );
}
