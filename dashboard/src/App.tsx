import { lazy, Suspense, useEffect } from "react";
import { useStore } from "./store";

const OrbitScene = lazy(() =>
  import("./scene/OrbitScene").then((module) => ({ default: module.OrbitScene })),
);
import { AddBar } from "./components/AddBar";
import { Stats } from "./components/Stats";
import { DownloadList } from "./components/DownloadList";
import { ThemeSwitcher } from "./components/ThemeSwitcher";
import { SettingsPanel } from "./components/SettingsPanel";
import { RemotePanel } from "./components/RemotePanel";

export default function App() {
  const init = useStore((s) => s.init);

  useEffect(() => {
    void init();
  }, [init]);

  return (
    <div className="relative min-h-screen overflow-x-hidden">
      {/* 3D scene as an immersive background */}
      <div className="pointer-events-none fixed inset-0 z-0 bg-[radial-gradient(circle_at_center,rgba(124,92,255,0.08),transparent_55%)]">
        <Suspense fallback={null}>
          <OrbitScene />
        </Suspense>
      </div>

      <div className="relative z-10 mx-auto max-w-5xl px-4 py-8">
        <header className="mb-8 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-3xl">🐾</span>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">Pounce</h1>
              <p className="text-sm text-white/50">Downloads, pounced.</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <ThemeSwitcher />
            <RemotePanel />
            <SettingsPanel />
            <Stats />
          </div>
        </header>

        <AddBar />
        <DownloadList />

        <footer className="mt-12 text-center text-xs text-white/30">
          Pounce · open-source download manager · engine + dashboard
        </footer>
      </div>
    </div>
  );
}
