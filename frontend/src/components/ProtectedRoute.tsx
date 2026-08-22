import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { CircleNotch } from "@phosphor-icons/react";

export function ProtectedRoute() {
  const { status } = useAuth();
  const location = useLocation();

  if (status === "loading") {
    return (
      <div className="flex min-h-[100dvh] items-center justify-center bg-ink-950">
        <CircleNotch size={28} className="animate-spin text-signal-500" aria-label="Loading" />
      </div>
    );
  }

  if (status === "unauthenticated") {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <Outlet />;
}
