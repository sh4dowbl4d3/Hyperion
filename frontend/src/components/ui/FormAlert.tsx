import { WarningCircle } from "@phosphor-icons/react";
import type { ReactNode } from "react";

export function FormAlert({ children }: { children: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-2.5 rounded-card border border-coral-pop/50 bg-coral-pop/10 px-4 py-3 text-sm text-ink-black"
    >
      <WarningCircle
        size={18}
        weight="fill"
        className="mt-0.5 shrink-0 text-coral-pop"
        aria-hidden
      />
      <span>{children}</span>
    </div>
  );
}
