import { create } from "zustand";
import { persist } from "zustand/middleware";

type State = {
  access: string | null;
  refresh: string | null;
  user: { id: string; username: string; email: string } | null;
  setSession: (a: string, r: string, u: State["user"]) => void;
  clear: () => void;
};

export const useAuth = create<State>()(
  persist(
    (set) => ({
      access: null,
      refresh: null,
      user: null,
      setSession: (a, r, u) => set({ access: a, refresh: r, user: u }),
      clear: () => set({ access: null, refresh: null, user: null }),
    }),
    { name: "tama-auth" },
  ),
);
