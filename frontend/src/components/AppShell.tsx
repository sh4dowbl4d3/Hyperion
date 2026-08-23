import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

const navItems = [
  { to: "/dashboard", label: "DASHBOARD" },
  { to: "/labs", label: "LABS" },
];

export function AppShell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="flex min-h-[100dvh] flex-col bg-cream-paper">
      {/* Flat top navigation bar */}
      <header className="border-b border-charcoal-ink bg-frost-white">
        <div className="mx-auto flex h-14 w-full max-w-[1200px] items-center justify-between px-4 sm:px-6">
          {/* Logo */}
          <NavLink to="/dashboard" className="flex items-center gap-2.5">
            <span
              aria-hidden
              className="flex size-7 items-center justify-center border-2 border-charcoal-ink bg-duck-bill-orange text-xs font-semibold"
            >
              M
            </span>
            <span className="text-body font-semibold tracking-[0.02em] text-charcoal-ink">
              ModernDVWA
            </span>
          </NavLink>

          {/* Center links (desktop) */}
          <nav
            aria-label="Main navigation"
            className="absolute left-1/2 hidden -translate-x-1/2 sm:block"
          >
            <ul className="flex items-center gap-1">
              {navItems.map((item) => (
                <li key={item.to}>
                  <NavLink
                    to={item.to}
                    className={({ isActive }) =>
                      `px-3 py-1.5 text-body-sm transition-colors ${
                        isActive
                          ? "bg-ice-wash font-medium text-charcoal-ink"
                          : "text-charcoal-ink hover:bg-chalk-gray"
                      }`
                    }
                  >
                    {item.label}
                  </NavLink>
                </li>
              ))}
            </ul>
          </nav>

          <div className="flex items-center gap-3">
            <span
              className="hidden max-w-[180px] truncate text-caption text-pencil-gray md:inline"
              title={user?.email}
            >
              {user?.email}
            </span>
            <button
              onClick={handleLogout}
              className="border-2 border-charcoal-ink bg-sky-crayon px-3 py-1.5 text-caption font-medium uppercase tracking-[0.02em] text-charcoal-ink transition-transform hover:translate-x-[2px] hover:translate-y-[2px]"
            >
              LOG OUT
            </button>
          </div>
        </div>
        {/* Mobile nav row */}
        <nav aria-label="Mobile navigation" className="border-t border-graphite sm:hidden">
          <ul className="mx-auto flex w-full max-w-[1200px]">
            {navItems.map((item) => (
              <li key={item.to} className="flex-1">
                <NavLink
                  to={item.to}
                  className={({ isActive }) =>
                    `block px-4 py-2.5 text-center text-body-sm ${
                      isActive ? "bg-ice-wash font-medium" : "text-charcoal-ink"
                    }`
                  }
                >
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </header>

      <main className="mx-auto w-full max-w-[1200px] flex-1 px-4 pb-20 pt-10 sm:px-6">
        <Outlet />
      </main>

      {/* Canary marquee footer band */}
      <footer className="overflow-hidden bg-canary-banner py-3">
        <p className="whitespace-nowrap text-center text-subheading font-medium uppercase tracking-[0.02em] text-charcoal-ink">
          LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS · LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS
        </p>
      </footer>
    </div>
  );
}
