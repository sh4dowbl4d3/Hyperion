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
      <label htmlFor={id} className="text-body font-medium text-charcoal-ink">
        {label}
      </label>
      <input
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : hint ? hintId : undefined}
        className={`h-11 rounded-[2px] border-2 bg-frost-white px-3 text-body tracking-[0.02em] text-charcoal-ink placeholder:text-pencil-gray transition-colors focus:border-sky-crayon ${
          error ? "border-coral-sketch" : "border-charcoal-ink"
        } ${className}`}
        {...rest}
      />
      {hint && !error && (
        <p id={hintId} className="text-caption text-pencil-gray">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-caption font-medium text-charcoal-ink">
          <span className="mr-1 inline-block size-1.5 bg-coral-sketch align-middle" aria-hidden />
          {error}
        </p>
      )}
    </div>
  );
}
