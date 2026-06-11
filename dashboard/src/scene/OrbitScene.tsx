import { Canvas } from "@react-three/fiber";
import { Stars, Float } from "@react-three/drei";
import { useStore } from "../store";
import { getTheme } from "../themes";
import { DownloadOrb } from "./DownloadOrb";
import { progress } from "../lib/format";

// Full-scene 3D backdrop: a starfield, a central pulsing core, and one orbiting
// orb per download. Colors are pulled from the active theme so the switcher
// restyles the WebGL scene too (CSS variables don't reach three.js materials).
const sceneCamera = {
  position: [0, 1.5, 8] as [number, number, number],
  fov: 55,
};

export function OrbitScene() {
  const downloads = useStore((s) => s.ordered());
  const themeKey = useStore((s) => s.theme);
  const t = getTheme(themeKey);
  const count = Math.max(downloads.length, 1);

  return (
    <Canvas camera={sceneCamera} dpr={[1, 2]}>
      <color attach="background" args={[t.bg]} />
      <ambientLight intensity={0.4} />
      <pointLight position={[0, 0, 0]} intensity={2} color={t.accent} />
      <pointLight position={[6, 4, 2]} intensity={0.6} color={t.glow} />

      <Stars radius={60} depth={40} count={2500} factor={4} fade speed={1} />

      {/* central core */}
      <Float speed={2} rotationIntensity={0.6} floatIntensity={0.6}>
        <mesh>
          <icosahedronGeometry args={[1, 1]} />
          <meshStandardMaterial
            color={t.card}
            emissive={t.accent}
            emissiveIntensity={0.5}
            wireframe
          />
        </mesh>
      </Float>

      {downloads.map((d, i) => (
        <DownloadOrb
          key={d.id}
          index={i}
          count={count}
          progress={progress(d) / 100}
          active={d.status === "running"}
          speed={d.speed}
          accent={t.accent}
          glow={t.glow}
          core={t.core}
        />
      ))}
    </Canvas>
  );
}
