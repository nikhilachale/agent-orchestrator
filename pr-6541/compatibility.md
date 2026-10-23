# Mastra compatibility evidence
PR adb222647b06e3f4b0ffa1adcb777106fb0ccf6f; skill04c8c95b23d717ccd2bc314e4577a0596b92e840.
Native package mastracode0.45.0 installed privately. Default semver installation resolved SDK1.15.0; explicit1.11.0 pin attempted before native/AO live probing.
- Binary: mastracode; npm package engines Node>=22.19.0. --version exits1 with required-prompt error; --help exits0. This is retained as a CLI version FAIL, not replaced with a fake successful version command.
- Official install: npm. AO only offers manual documentation; no executable installer.
- Login: native --help exposes no login/status command. No invented status command in the contract. Existing login remains unverified; no credential file inspected or copied.
- Models: --help exposes --model provider/model freeform; no model-list command. Contract omits model list.
- Headless permission modes: auto/deny. TUI policy has separate AO mapping, default and explicit bypass only; unsupported modes fail closed.
- TUI, private instructions/hooks, exact restore: product source claims native TUI with --tui-initial-prompt and resume by UUID. Live qualification blocked until verified AO authentication; not asserted by source.
- Chat/reviewer: unregistered intentional scope, unsupported; no Chat contract surface added.
- Native auth and model list: BLOCKED/unsupported contract commands, not PASS.
