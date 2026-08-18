import { execSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";

execSync("npx tsc", { stdio: "inherit" });

mkdirSync("dist", { recursive: true });
writeFileSync(
  "dist/index.mjs",
  `import mod from "./index.js";
export const { API_VERSION, parseDigest, canonicalJSON, digestOf, digestFromBytes, Reference, ContractError } = mod;
export const { EvidenceBundle, ItemKind } = mod;
export const { HarnessVersion, LocalRunner, COMPONENT_KINDS } = mod;
export const { StaticUICapture, UiRepairFailureEvidence } = mod;
export const { TraceEnvelope, TracePerspective } = mod;
export default mod;
`
);
