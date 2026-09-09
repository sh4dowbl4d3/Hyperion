import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { ThemeToggle } from "./ui/ThemeToggle";

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
    <div className="flex min-h-[100dvh] flex-col bg-black text-[#8a8a6f]">
      {/* Terminal top navigation bar */}
      <header className="border-b border-[#2a2a26] bg-black">
        <div className="mx-auto flex h-14 w-full max-w-[1200px] items-center justify-between px-4 sm:px-6">
          {/* Logo */}
          <NavLink to="/dashboard" className="flex items-center gap-2.5">
            <span
              aria-hidden
              className="flex size-7 items-center justify-center border border-[#ffa133] bg-[#ffa133] text-xs font-bold text-black"
            >
              ▶
            </span>
            <span className="text-body font-semibold tracking-[0.06em] text-[#eeeeee]">
              HYPERION
            </span>
            <span className="hidden sm:inline border-l border-[#333333] pl-2 text-[11px] text-[#777766]">
              SEC-LABS
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
                      `px-3 py-1.5 text-body-sm tracking-[0.04em] transition-colors ${
                        isActive
                          ? "bg-[#222222] font-medium text-[#eeeeee] border-b-2 border-[#ffa133]"
                          : "text-[#8a8a6f] hover:text-[#eeeeee] hover:bg-[#161614]"
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
              className="hidden max-w-[180px] truncate text-caption text-[#777766] md:inline"
              title={user?.email}
            >
              {user?.email}
            </span>
            <ThemeToggle />
            <button
              onClick={handleLogout}
              className="border border-[#8a8a6f] bg-transparent px-3 py-1.5 text-caption font-medium uppercase tracking-[0.04em] text-[#8a8a6f] transition-colors hover:bg-[#8a8a6f] hover:text-black active:bg-[#ffa133] active:border-[#ffa133]"
            >
              LOG OUT
            </button>
          </div>
        </div>
        {/* Mobile nav row */}
        <nav aria-label="Mobile navigation" className="border-t border-[#2a2a26] sm:hidden">
          <ul className="mx-auto flex w-full max-w-[1200px]">
            {navItems.map((item) => (
              <li key={item.to} className="flex-1">
                <NavLink
                  to={item.to}
                  className={({ isActive }) =>
                    `block px-4 py-2.5 text-center text-body-sm tracking-[0.04em] ${
                      isActive ? "bg-[#222222] font-medium text-[#eeeeee] border-b-2 border-[#ffa133]" : "text-[#8a8a6f]"
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
    </div>
  );
}
