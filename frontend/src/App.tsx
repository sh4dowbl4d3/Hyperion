import { Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./hooks/useAuth";
import { AppShell } from "./components/AppShell";
import { AutoAuth } from "./components/AutoAuth";
import Dashboard from "./pages/Dashboard";
import Labs from "./pages/Labs";
import LabDetailPage from "./pages/LabDetail";

export default function App() {
  return (
    <AuthProvider>
      <AutoAuth>
        <Routes>
            <Route element={<AppShell />}>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/labs" element={<Labs />} />
              <Route path="/labs/:slug" element={<LabDetailPage />} />
            </Route>

            {/* Legacy auth routes redirect directly to dashboard */}
            <Route path="/login" element={<Navigate to="/dashboard" replace />} />
            <Route path="/register" element={<Navigate to="/dashboard" replace />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </AutoAuth>
      </AuthProvider>
  );
}
