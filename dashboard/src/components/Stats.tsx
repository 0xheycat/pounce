import { useStore } from "../store";
import { formatSpeed } from "../lib/format";

export function Stats() {
  const downloads = useStore((s) => Object.values(s.downloads));
  const active = downloads.filter((d) => d.status === "running");
  const totalSpeed = active.reduce((sum, d) => sum + (d.speed || 0), 0);
  const completed = downloads.filter((d) => d.status === "completed").length;

  return (
    <div className="flex gap-6 text-right text-sm">
      <div>
        <div className="font-semibold text-pounce-glow">
          {formatSpeed(totalSpeed)}
        </div>
        <div className="text-white/40">total speed</div>
      </div>
      <div>
        <div className="font-semibold">{active.length}</div>
        <div className="text-white/40">active</div>
      </div>
      <div>
        <div className="font-semibold">{completed}</div>
        <div className="text-white/40">done</div>
      </div>
    </div>
  );
}
