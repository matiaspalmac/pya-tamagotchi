import { useEffect, useState } from "react";
import { api } from "../api/client";
import { connectWS } from "../api/ws";
import { useAuth } from "../store/auth";
import { PetCard } from "../components/PetCard";

export function Dashboard() {
  const { access, user, clear } = useAuth();
  const [pets, setPets] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [err, setErr] = useState("");

  async function load() {
    try {
      setPets(await api.petsMine());
    } catch (e: any) {
      setErr(e.message);
    }
  }

  useEffect(() => { load(); }, []);

  useEffect(() => {
    if (!access) return;
    const ws = connectWS(access, (msg) => {
      if (msg.type === "pet:tick" || msg.type === "pet:event" || msg.type === "pet:evolved") {
        load();
      }
    });
    return () => ws.close();
  }, [access]);

  async function create(e: React.FormEvent) {
    e.preventDefault();
    if (!name) return;
    try {
      await api.createPet(name);
      setName("");
      load();
    } catch (e: any) {
      setErr(e.message);
    }
  }

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      <header className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Mis mascotas</h1>
        <div className="flex items-center gap-4">
          <span className="opacity-70">{user?.username}</span>
          <button onClick={clear} className="btn-secondary">Salir</button>
        </div>
      </header>

      <form onSubmit={create} className="flex gap-2 mb-6">
        <input className="input flex-1" placeholder="nombre nueva mascota" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="btn-primary">Crear</button>
      </form>

      {err && <div className="text-red-400 mb-4">{err}</div>}

      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {pets.map((p) => <PetCard key={p.id} pet={p} onChange={load} />)}
      </div>
      {pets.length === 0 && <p className="opacity-60">Aún no tienes mascotas.</p>}
    </div>
  );
}
