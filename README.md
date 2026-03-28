# zone2

Small Go CLI to control **Zone 2 power** on Arcam/AudioControl/JBL Synthesis receivers that expose the setup WebSocket API on port `50001`.

The tool supports:

- `on`
- `off`
- `toggle`
- `status` (prints `on` or `off`)

## Build

Use the provided build script:

```bash
./build.sh
```

It generates:

- `zone2` -> Linux ARM64 binary (for Home Assistant OS on Raspberry Pi 5)
- `zone2-macos-arm64` -> local Apple Silicon test binary
- `zone2-macos-amd64` -> local Intel macOS test binary

## Local usage

Examples:

```bash
./zone2-macos-arm64 -host YOUR_AVR_IP -mode status
./zone2-macos-arm64 -host YOUR_AVR_IP -mode on
./zone2-macos-arm64 -host YOUR_AVR_IP -mode off
./zone2-macos-arm64 -host YOUR_AVR_IP -mode toggle
```

Flags:

- `-host` (default: `YOUR_AVR_IP`)
- `-mode` (`on|off|toggle|status`)
- `-timeout` (default: `4s`)
- `-verify` (default: `20`, only used for writes)
- `-verbose` (prints raw TX/RX frames)

## Home Assistant OS (native on Raspberry Pi 5)

### 1) Build the HA binary

On your dev machine:

```bash
./build.sh
```

This creates `zone2` (Linux ARM64).

### 2) Copy binary to HA config storage

On Home Assistant, create a bin folder:

```bash
mkdir -p /config/bin
```

Copy `zone2` to `/config/bin/zone2` (for example via Samba share or SSH/SCP).

Make it executable:

```bash
chmod +x /config/bin/zone2
```

### 3) Quick manual test from HA terminal

```bash
/config/bin/zone2 -host YOUR_AVR_IP -mode status -timeout 4s
/config/bin/zone2 -host YOUR_AVR_IP -mode on -timeout 4s -verify 20
/config/bin/zone2 -host YOUR_AVR_IP -mode off -timeout 4s -verify 20
```

### 4) Add a switch entity in `configuration.yaml`

```yaml
command_line:
  - switch:
      name: AVR Zone 2
      unique_id: avr_zone2
      command_on: "/config/bin/zone2 -host YOUR_AVR_IP -mode on -timeout 4s -verify 20"
      command_off: "/config/bin/zone2 -host YOUR_AVR_IP -mode off -timeout 4s -verify 20"
      command_state: "/config/bin/zone2 -host YOUR_AVR_IP -mode status -timeout 4s"
      value_template: "{{ value | trim | lower == 'on' }}"
      scan_interval: 10
```

### 5) Validate and restart Home Assistant Core

```bash
ha core check
ha core restart
```

After restart, add `switch.avr_zone_2` to your dashboard.

## Notes

- Receiver must be reachable from HA host on TCP `50001`.
- If commands are intermittent, keep `-timeout 4s` and `-verify 20`.
- `status` output is intentionally plain (`on` or `off`) for easy HA parsing.
