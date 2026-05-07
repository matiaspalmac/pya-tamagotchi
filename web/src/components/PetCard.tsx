import { api } from "../api/client";

type Stats = { hunger: number; happy: number; energy: number; health: number };
type Pet = Stats & {
  id: string;
  name: string;
  species: string;
  stage: string;
  xp: number;
  born_at: string;
  died_at?: string | null;
};

const stageEmoji: Record<string, string> = {
  egg: "🥚", baby: "🐣", teen: "🐤", adult: "🐔", elder: "🦅",
};

export function PetCard({ pet, onChange }: { pet: Pet; onChange: () => void }) {
  const dead = !!pet.died_at;

  async function act(fn: (id: string) => Promise<any>) {
    try {
      await fn(pet.id);
      onChange();
    } catch (e: any) {
      alert(e.message);
    }
  }

  return (
    <div className={`bg-slate-800 rounded-2xl p-5 ${dead ? "opacity-50" : ""}`}>
      <div className="flex justify-between items-start mb-3">
        <div>
          <div className="text-xl font-bold">{stageEmoji[pet.stage] || "🐣"} {pet.name}</div>
          <div className="text-xs opacity-60">{pet.stage} · xp {pet.xp}</div>
        </div>
        {dead && <span className="text-red-400 text-sm">muerta</span>}
      </div>
      <Bar label="Hambre" v={pet.hunger} color="bg-amber-500" />
      <Bar label="Feliz"  v={pet.happy}  color="bg-pink-500" />
      <Bar label="Energía" v={pet.energy} color="bg-sky-500" />
      <Bar label="Salud"  v={pet.health} color="bg-emerald-500" />
      {!dead && (
        <div className="grid grid-cols-2 gap-2 mt-4">
          <button className="btn-action" onClick={() => act(api.feed)}>Alimentar</button>
          <button className="btn-action" onClick={() => act(api.play)}>Jugar</button>
          <button className="btn-action" onClick={() => act(api.sleep)}>Dormir</button>
          <button className="btn-action" onClick={() => act(api.heal)}>Curar</button>
        </div>
      )}
    </div>
  );
}

function Bar({ label, v, color }: { label: string; v: number; color: string }) {
  return (
    <div className="mb-2">
      <div className="flex justify-between text-xs mb-1">
        <span>{label}</span>
        <span className="opacity-60">{v}</span>
      </div>
      <div className="h-2 bg-slate-700 rounded">
        <div className={`h-full rounded ${color}`} style={{ width: `${v}%` }} />
      </div>
    </div>
  );
}
