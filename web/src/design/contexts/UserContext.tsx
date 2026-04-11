import { createContext, useContext, ReactNode } from "react";
import { UserContext } from "../../context/User";
import { API } from "../../helpers";

export type UserRole = "admin" | "user";

interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar?: string;
}

interface UserContextType {
  user: User | null;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
  isAdmin: () => boolean;
}

const DesignUserContext = createContext<UserContextType | undefined>(undefined);

export function DesignUserProvider({ children }: { children: ReactNode }) {
  const [userState, userDispatch] = useContext(UserContext);

  const rawUser = userState?.user;

  const user: User | null = rawUser
    ? {
        id: String(rawUser.id),
        name: rawUser.display_name || rawUser.username || "",
        email: rawUser.email || "",
        role: rawUser.role >= 10 ? "admin" : "user",
      }
    : null;

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      const res = await API.post("/api/user/login", { username, password });
      const { success, data } = res.data;
      if (success) {
        userDispatch({ type: "login", payload: data });
        localStorage.setItem("user", JSON.stringify(data));
        return true;
      }
      return false;
    } catch {
      return false;
    }
  };

  const logout = () => {
    userDispatch({ type: "logout" });
    localStorage.removeItem("user");
  };

  const isAdmin = () => {
    return user?.role === "admin";
  };

  return (
    <DesignUserContext.Provider value={{ user, login, logout, isAdmin }}>
      {children}
    </DesignUserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(DesignUserContext);
  if (context === undefined) {
    throw new Error("useUser must be used within a DesignUserProvider");
  }
  return context;
}
