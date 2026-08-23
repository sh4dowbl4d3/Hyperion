import { useState } from "react";
import type { FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { isApiError } from "../lib/api";
import { useAuth } from "../hooks/useAuth";
import { AuthLayout } from "../layouts/AuthLayout";
import { Button } from "../components/ui/Button";
import { TextField } from "../components/ui/TextField";
import { FormAlert } from "../components/ui/FormAlert";

const friendlyError: Record<string, string> = {
  invalid_credentials: "Incorrect email or password.",
  network_error: "Cannot reach the server. Is the backend running?",
};

export default function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    setFieldErrors({});

    if (!email.trim()) {
      setFieldErrors((prev) => ({ ...prev, email: "Email is required." }));
      return;
    }
    if (!password) {
      setFieldErrors((prev) => ({ ...prev, password: "Password is required." }));
      return;
    }

    setSubmitting(true);
    try {
      await login(email, password);
      const from = (location.state as { from?: string } | null)?.from ?? "/dashboard";
      navigate(from, { replace: true });
    } catch (cause) {
      setFormError(
        isApiError(cause)
          ? friendlyError[cause.code] ?? cause.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthLayout>
      <h2 className="text-heading-sm font-medium tracking-tight text-ink-black">Sign in</h2>
      <p className="mt-2 mb-8 text-[15px] text-stone-gray">Access your training dashboard.</p>

      {formError && (
        <div className="mb-6">
          <FormAlert>{formError}</FormAlert>
        </div>
      )}

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-5">
        <TextField
          label="Email"
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={fieldErrors.email}
        />
        <TextField
          label="Password"
          type="password"
          autoComplete="current-password"
          placeholder="Your password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
        />
        <Button type="submit" loading={submitting} className="mt-1 w-full">
          Sign in
        </Button>
      </form>

      <p className="mt-8 text-sm text-stone-gray">
        New here?{" "}
        <Link
          to="/register"
          className="font-medium text-ink-black underline decoration-stone-gray underline-offset-4 hover:decoration-ink-black"
        >
          Create an account
        </Link>
      </p>
    </AuthLayout>
  );
}
