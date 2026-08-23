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
  // Structural green pill — the brand accent, used sparingly.
  primary:
    "bg-fresh-grass text-ink-black hover:brightness-105 active:translate-y-[1px] font-medium",
  // Light ghost pill with a chromatic dot affordance.
  ghost:
    "bg-pure-white text-ink-black border border-hairline-mist hover:border-stone-gray active:translate-y-[1px] font-medium",
  // Coral action pill — reserved for lab-level actions.
  coral:
    "bg-coral-pop text-white hover:brightness-105 active:translate-y-[1px] font-medium",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", loading = false, disabled, className = "", children, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      disabled={disabled ?? loading}
      className={`inline-flex h-11 items-center justify-center gap-2 rounded-pill px-5 text-[15px] transition-all duration-150 disabled:cursor-not-allowed disabled:opacity-55 ${variantClasses[variant]} ${className}`}
      {...rest}
    >
      {loading && <CircleNotch size={18} className="animate-spin" aria-hidden />}
      {children}
    </button>
  );
});
