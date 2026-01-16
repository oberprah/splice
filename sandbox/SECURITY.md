# Security Model

This document describes the security architecture for running Claude Code in an isolated Kubernetes environment.

## Isolation Layers

```
+-----------------------------------------------------+
|  Host Machine (macOS)                               |
|  +-----------------------------------------------+  |
|  |  Kind Node (Docker container)                 |  |
|  |  +----------------------------------------+   |  |
|  |  |  agent-env namespace                   |   |  |
|  |  |  +-----------+      +---------------+  |   |  |
|  |  |  | claude-   | ---- |  api-proxy    |  |   |  |
|  |  |  | agent     |:8080 |  (envoy)      |------------> api.anthropic.com
|  |  |  |           |:8081 |               |------------> api.openai.com
|  |  |  +-----------+      +---------------+  |   |  |
|  |  +----------------------------------------+   |  |
|  +-----------------------------------------------+  |
+-----------------------------------------------------+
```

### Layer 1: Network Isolation (NetworkPolicy)

The agent pod can only communicate with:
- **api-proxy** on port 8080 (for Anthropic API calls)
- **api-proxy** on port 8081 (for OpenAI API calls)
- **kube-dns** on port 53 (for DNS resolution)

All other egress is blocked. The agent cannot:
- Reach the internet directly
- Push to git remotes (github.com, gitlab.com, etc.)
- Exfiltrate data via network (except through the API proxy)

### Layer 2: API Proxy (Envoy)

All API traffic flows through the Envoy proxy which:
- Injects authentication credentials (agent never sees real tokens)
- Routes to api.anthropic.com (port 8080) and api.openai.com (port 8081)
- Logs all requests for audit

### Layer 3: Container Hardening

The agent container runs with restricted privileges:

```yaml
securityContext:
  runAsNonRoot: true        # Cannot run as root
  runAsUser: 1000           # Runs as unprivileged user
  allowPrivilegeEscalation: false  # Cannot gain privileges
  capabilities:
    drop: ["ALL"]           # No Linux capabilities
  seccompProfile:
    type: RuntimeDefault    # Restricted syscalls
```

### Layer 4: Kind Isolation

The Kubernetes cluster runs inside a Docker container (kind). Even if an attacker escapes the pod, they land in the kind node container, not directly on the host. A second container escape would be required.

## Threat Model

### Mitigated Threats

| Threat | Mitigation |
|--------|------------|
| Data exfiltration via network | NetworkPolicy blocks all egress except proxy |
| Credential theft | Agent never sees real API tokens |
| Privilege escalation in container | securityContext prevents escalation |
| Container escape via capabilities | All capabilities dropped |
| Direct internet access | Only proxy is reachable |
| Git push (code exfiltration) | Git hosts blocked by NetworkPolicy |

### Residual Risks

| Risk | Likelihood | Impact | Notes |
|------|------------|--------|-------|
| Kernel exploit (container escape) | Low | High | Requires 0-day, mitigated by seccomp |
| Data exfiltration via API responses | Medium | Medium | Agent could encode data in API calls |
| Malicious code in mounted volume | Medium | Medium | Agent can write to mounted directories |
| Kind container escape | Very Low | High | Requires two escapes (pod -> kind -> host) |

### Considerations for Volume Mounts

When mounting a repository into the agent pod:

1. **Mount scope**: Only mount the specific repo directory, never broader paths
2. **Review before running**: Agent can modify files; review changes before executing on host
3. **No sensitive data**: Don't mount directories containing credentials or secrets
4. **Git push blocked**: Agent can commit but cannot push (NetworkPolicy)

## Audit

### Proxy Logs

View all API requests made by the agent:
```bash
kubectl logs -n agent-env api-proxy
```

### Agent Activity

Exec into the agent to inspect:
```bash
kubectl exec -it -n agent-env claude-agent -- bash
```

## Future Hardening (Optional)

Additional hardening that could be added:

- `readOnlyRootFilesystem: true` with explicit writable mounts
- Pod Security Standards (restricted)
- Runtime security monitoring (Falco)
- Read-only volume mounts where write access isn't needed
