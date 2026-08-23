import { useId } from "react";
import type { InputHTMLAttributes } from "react";

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  error?: string;
  hint?: string;
};

export function TextField({ label, error, hint, className = "", ...rest }: TextFieldProps) {
  const id = useId();
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;

  return (
    <div className="flex flex-col gap-2">
      <label htmlFor={id} className="text-[15px] font-medium text-ink-black">
        {label}
      </label>
      <input
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : hint ? hintId : undefined}
        className={`h-11 rounded-pill border bg-pure-white px-4 text-[15px] text-ink-black placeholder:text-stone-gray transition-colors duration-150 focus:border-ink-black ${
          error ? "border-coral-pop" : "border-hairline-mist"
        } ${className}`}
        {...rest}
      />
      {hint && !error && (
        <p id={hintId} className="text-sm text-stone-gray">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-sm text-coral-pop">
          {error}
        </p>
      )}
    </div>
  );
}
