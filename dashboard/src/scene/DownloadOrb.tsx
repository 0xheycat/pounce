import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import type { Mesh, Group } from "three";

interface Props {
  index: number;
  count: number;
  progress: number; // 0..1
  active: boolean;
  speed: number; // bytes/sec, drives glow
  accent: string;
  glow: string;
  core: string;
}

// A single glowing orb that orbits the center. Its emissive intensity tracks
// throughput, and a torus ring fills up with progress. Colors come from theme.
export function DownloadOrb({
  index,
  count,
  progress,
  active,
  speed,
  accent,
  glow,
  core,
}: Props) {
  const group = useRef<Group>(null);
  const orb = useRef<Mesh>(null);

  const radius = 2.4 + (index % 3) * 0.8;
  const baseAngle = (index / Math.max(count, 1)) * Math.PI * 2;
  const tilt = (index % 2 === 0 ? 1 : -1) * 0.4;
  const intensity = active
    ? 0.6 + Math.min(speed / (2 * 1024 * 1024), 2)
    : 0.15;

  useFrame((state) => {
    const time = state.clock.elapsedTime;
    const speedFactor = active ? 0.4 : 0.08;
    const angle = baseAngle + time * speedFactor;
    if (group.current) {
      group.current.position.set(
        Math.cos(angle) * radius,
        Math.sin(angle) * radius * tilt,
        Math.sin(angle) * radius,
      );
    }
    if (orb.current) {
      orb.current.scale.setScalar(0.5 + progress * 0.5);
    }
  });

  const ringArc = Math.PI * 2 * Math.max(progress, 0.001);

  return (
    <group ref={group}>
      <mesh ref={orb}>
        <sphereGeometry args={[0.35, 32, 32]} />
        <meshStandardMaterial
          color={active ? accent : "#3a3a55"}
          emissive={active ? glow : core}
          emissiveIntensity={intensity}
          roughness={0.3}
          metalness={0.6}
        />
      </mesh>
      {/* progress ring */}
      <mesh rotation={[Math.PI / 2, 0, 0]}>
        <torusGeometry args={[0.55, 0.03, 8, 64, ringArc]} />
        <meshStandardMaterial
          color={glow}
          emissive={glow}
          emissiveIntensity={0.8}
        />
      </mesh>
    </group>
  );
}
