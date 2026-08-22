import { Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./hooks/useAuth";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AppShell } from "./components/AppShell";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        <Route element={<ProtectedRoute />}>
          <Route element={<AppShell />}>
            <Route path="/dashboard" element={<Dashboard />} />
          </Route>
        </Route>

        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </AuthProvider>
  );
}
