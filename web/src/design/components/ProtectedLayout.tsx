import { useEffect } from "react";
import { useNavigate, Outlet } from "react-router-dom";
import { useUser } from "../contexts/UserContext";

export function ProtectedLayout() {
  const { user } = useUser();
  const navigate = useNavigate();

  useEffect(() => {
    if (!user) {
      navigate("/login");
    }
  }, [user, navigate]);

  if (!user) {
    return null;
  }

  return <Outlet />;
}
