import { ShieldWarning } from "@phosphor-icons/react";
import type { ReactNode } from "react";

export function FormAlert({ children }: { children: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex items-start gap-2.5 rounded-lg border border-alert-500/40 bg-alert-500/10 px-3.5 py-3 text-sm text-fog-100"
    >
      <ShieldWarning size={18} weight="fill" className="mt-0.5 shrink-0 text-alert-500" aria-hidden />
      <span>{children}</span>
    </div>
  );
}
