import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { CircleNotch } from "@phosphor-icons/react";

type ButtonVariant = "primary" | "ghost" | "coral";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  loading?: boolean;
  children: ReactNode;
};

const variantClasses: Record<ButtonVariant, string> = {
  // Sky-crayon filled action — the only chromatic fill in the system.
  primary:
    "bg-sky-crayon text-charcoal-ink border-2 border-charcoal-ink hover:translate-x-[3px] hover:translate-y-[3px] hover:shadow-none active:translate-x-[6px] active:translate-y-[6px] active:shadow-none font-medium",
  // Charcoal outlined companion.
  ghost:
    "bg-frost-white text-charcoal-ink border-2 border-charcoal-ink hover:translate-x-[3px] hover:translate-y-[3px] hover:shadow-none active:translate-x-[6px] active:translate-y-[6px] active:shadow-none font-medium",
  // Canary highlight — decorative emphasis (race bursts).
  coral:
    "bg-canary-banner text-charcoal-ink border-2 border-charcoal-ink hover:translate-x-[3px] hover:translate-y-[3px] hover:shadow-none active:translate-x-[6px] active:translate-y-[6px] active:shadow-none font-medium",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", loading = false, disabled, className = "", children, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      disabled={disabled ?? loading}
      className={`inline-flex h-11 items-center justify-center gap-2 rounded-[2px] px-5 text-body-sm uppercase tracking-[0.02em] transition-all duration-100 shadow-offset disabled:cursor-not-allowed disabled:opacity-55 ${variantClasses[variant]} ${className}`}
      {...rest}
    >
      {loading && <CircleNotch size={16} className="animate-spin" aria-hidden />}
      {children}
    </button>
  );
});
