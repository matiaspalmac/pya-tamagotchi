import { useAuth } from "../store/auth";

const BASE = (import.meta.env.VITE_API_URL as string | undefined) ?? "";

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = useAuth.getState().access;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((init.headers as Record<string, string>) || {}),
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(BASE + path, { ...init, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || "request failed");
  }
  return res.json();
}

export const api = {
  register: (email: string, username: string, password: string) =>
    req<{ user: any; tokens: { access_token: string; refresh_token: string } }>(
      "/auth/register",
      { method: "POST", body: JSON.stringify({ email, username, password }) },
    ),
  login: (email: string, password: string) =>
    req<{ user: any; tokens: { access_token: string; refresh_token: string } }>(
      "/auth/login",
      { method: "POST", body: JSON.stringify({ email, password }) },
    ),
  me: () => req<{ user: any }>("/auth/me"),
  petsMine: () => req<any[]>("/pets/mine"),
  createPet: (name: string, species = "blob") =>
    req<any>("/pets", { method: "POST", body: JSON.stringify({ name, species }) }),
  pet: (id: string) => req<any>(`/pets/${id}`),
  feed:  (id: string) => req<any>(`/pets/${id}/feed`,  { method: "POST" }),
  play:  (id: string) => req<any>(`/pets/${id}/play`,  { method: "POST" }),
  sleep: (id: string) => req<any>(`/pets/${id}/sleep`, { method: "POST" }),
  heal:  (id: string) => req<any>(`/pets/${id}/heal`,  { method: "POST" }),
};
