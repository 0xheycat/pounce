import { useStore } from "../store";
import { DownloadCard } from "./DownloadCard";

export function DownloadList() {
  const ordered = useStore((s) => s.ordered());

  if (ordered.length === 0) {
    return (
      <div className="glass rounded-2xl p-10 text-center text-white/40">
        No downloads yet. Paste a URL above and let Pounce do the rest. 🐾
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {ordered.map((d) => (
        <DownloadCard key={d.id} d={d} />
      ))}
    </div>
  );
}
