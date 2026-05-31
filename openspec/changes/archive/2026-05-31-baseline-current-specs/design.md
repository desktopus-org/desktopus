## Context

Desktopus already ships a working CLI (build → run → connect, plus module and volume management). OpenSpec was just initialized on this branch and `openspec/specs/` is empty. This change reverse-documents the shipped behavior into capability specs so they can serve as the source of truth going forward.

## Goals / Non-Goals

**Goals:**
- Capture the currently-shipped contract as source-of-truth specs.
- Carve the system into coherent capabilities aligned to the `internal/` package layout, so specs live conceptually close to the code.
- Provide a paper trail (this change) for where the initial specs came from.

**Non-Goals:**
- No behavior changes — documentation only.
- Not documenting unimplemented or reserved surface (the future server API, the state store DB) as behavior; these are noted as reserved configuration only.
- Not exhaustively enumerating every flag default — specs stay at the contract level.

## Decisions

- **Seven capabilities**, mapped to the internal package split:
  `desktop-authoring`, `module-system`, `image-build`, `container-runtime`, `volume-management`, `desktop-viewer`, `app-configuration`.
- **`desktopus.runtime.yaml` folds into `container-runtime`** rather than being its own capability, because it is the input contract to `run` and has no independent behavior.
- **Global flags, `version`, and `~/.desktopus/config.yaml` fold into `app-configuration`** — they are cross-cutting CLI/runtime configuration rather than a feature.
- **Specs describe behavior at the contract level** (WHEN/THEN), citing concrete values only where the value *is* the contract: the OS×desktop compatibility matrix, container ports 3000/3001, the `org.desktopus.*` label keys, and the default user (`desktopus`) / home (`/home/<user>` or `/config` for `abc`).
- **Present-only facts are stated as the current set, not the requirement.** Example: "ships built-in modules" is the durable requirement; "currently only `chrome`" is noted as today's set so the spec stays honest without freezing the catalog.

## Risks / Trade-offs

- **Spec drift**: reverse-documented specs can fall out of sync with code. Mitigated by tasks that verify each spec against its source package during apply.
- **Reality vs intent**: chose to document what exists (e.g. only `chrome` ships) while noting the extensible contract, rather than speccing aspirational features.
- **Altitude**: keeping specs behavioral risks omitting a detail that later turns out contractual; concrete values were pulled in only where they are externally observable guarantees.
