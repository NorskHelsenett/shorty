// check if a user is admin

import axios from "axios";
import { ReactNode, createContext, useContext } from "react";
import { AuthContext, IAuthContext } from "react-oauth2-code-pkce";
import useSWR from "swr";
import { API_URL } from "../Service/config";

// Create context
const AdminContext = createContext<boolean>(false);
const DEFAULT_REFRESH_INTERVAL = 1000 * 60 * 5; // 5 min

// Provider
export const AdminProvider = ({ children }: { children: ReactNode }) => {
  const { token }: IAuthContext = useContext(AuthContext);
  const { data: isAdmin } = useSWR(["/api/v1", token], fetcher, {
    refreshInterval: DEFAULT_REFRESH_INTERVAL,
  }); //

  return (
    <AdminContext.Provider value={isAdmin ?? false}>
      {children}
    </AdminContext.Provider>
  );
};

// hook
export const useAdminContext = () => {
  const context = useContext(AdminContext);

  return context;
};

const fetcher = async ([_path, token]: [string, string]) => {
  try {
    const response = await axios.get(`${API_URL}/v1/`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const adminHeader = response.headers["x-is-admin"];

    if (adminHeader) {
      const userIsAdmin = adminHeader === "true";
      return userIsAdmin;
    }
    return false;
  } catch (error) {
    console.error("Error fetching admin status:", error);
    return false;
  }
};
