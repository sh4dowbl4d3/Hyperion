import { Sun, Moon } from "@phosphor-icons/react";
import { useTheme } from "../../hooks/useTheme";

export function ThemeToggle({ className = "" }: { className?: string }) {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  return (
    <button
      type="button"
      onClick={toggleTheme}
      aria-label={`Switch to ${isDark ? "light" : "dark"} mode`}
      title={`Switch to ${isDark ? "light" : "dark"} mode`}
      className={`inline-flex items-center gap-1.5 border border-[#333333] px-2.5 py-1 text-caption font-medium uppercase tracking-[0.04em] transition-colors cursor-pointer text-[#8a8a6f] hover:bg-[#8a8a6f] hover:text-black dark:hover:text-black ${className}`}
      style={{
        borderColor: "var(--theme-border-subtle)",
        color: "var(--theme-text-primary)",
        backgroundColor: "transparent",
      }}
    >
      {isDark ? (
        <>
          <Sun size={13} weight="bold" className="text-[#ffa133]" aria-hidden />
          <span>LIGHT</span>
        </>
      ) : (
        <>
          <Moon size={13} weight="bold" className="text-[#d46300]" aria-hidden />
          <span>DARK</span>
        </>
      )}
    </button>
  );
}
