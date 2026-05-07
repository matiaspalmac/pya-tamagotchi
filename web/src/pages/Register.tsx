import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";

export function Register() {
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState("");
  const setSession = useAuth((s) => s.setSession);
  const nav = useNavigate();

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    try {
      const r = await api.register(email, username, password);
      setSession(r.tokens.access_token, r.tokens.refresh_token, r.user);
      nav("/");
    } catch (e: any) {
      setErr(e.message);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center">
      <form onSubmit={submit} className="bg-slate-800 p-8 rounded-2xl w-96 space-y-4">
        <h1 className="text-2xl font-bold">Registro</h1>
        <input className="input" placeholder="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input className="input" placeholder="username" value={username} onChange={(e) => setUsername(e.target.value)} />
        <input className="input" type="password" placeholder="contraseña (8+)" value={password} onChange={(e) => setPassword(e.target.value)} />
        {err && <div className="text-red-400 text-sm">{err}</div>}
        <button className="btn-primary w-full">Crear cuenta</button>
        <p className="text-sm">¿Ya tienes? <Link to="/login" className="underline">Login</Link></p>
      </form>
    </div>
  );
}
