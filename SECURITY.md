# Security Policy

## Supported Versions

NEXUS is currently in **alpha stage**. Security updates are provided for the latest release only.

| Version | Status | Security Support |
|---------|--------|------------------|
| 0.2.x-alpha | ✅ Active | ✅ Full support |
| 0.1.x-alpha | ❌ Deprecated | ❌ No support |
| < 0.1.0 | ❌ Unsupported | ❌ No support |

**Note**: Alpha versions may contain security vulnerabilities. **Do not use in production environments.**

---

## Reporting a Vulnerability

**We take security seriously.** If you discover a security vulnerability in NEXUS, please report it responsibly.

### Reporting Process

1. **DO NOT** open a public GitHub issue for security vulnerabilities
2. **DO NOT** disclose the vulnerability publicly until it's been fixed
3. **DO** email security reports to: **[security@nexus-ai.dev](mailto:61723019+Omkar0612@users.noreply.github.com)** (temporary)

### What to Include in Your Report

Please provide as much detail as possible:

- **Vulnerability description**: Clear explanation of the issue
- **Impact assessment**: What can an attacker do?
- **Affected versions**: Which releases are vulnerable?
- **Reproduction steps**: Detailed steps to reproduce
- **Proof of concept**: Code, screenshots, or videos (if applicable)
- **Suggested fix**: Optional, but helpful
- **Your contact info**: For follow-up questions

### Example Report Template

```markdown
**Vulnerability Type**: [e.g., SQL Injection, XSS, Path Traversal]
**Severity**: [Critical / High / Medium / Low]
**Affected Component**: [e.g., internal/vault, internal/browser]
**Affected Versions**: [e.g., v0.2.0-alpha and earlier]

**Description**:
[Detailed explanation of the vulnerability]

**Steps to Reproduce**:
1. [First step]
2. [Second step]
3. [Observe the vulnerability]

**Impact**:
[What can an attacker achieve? Data breach? Code execution? DoS?]

**Suggested Mitigation**:
[Optional: Your recommendation for fixing it]

**PoC Code**:
```go
// Your proof-of-concept code here
```
```

---

## Response Timeline

| Phase | Timeline | Action |
|-------|----------|--------|
| **Initial Response** | Within 48 hours | Acknowledge receipt |
| **Triage** | Within 1 week | Assess severity and impact |
| **Fix Development** | 1-4 weeks | Develop and test patch |
| **Disclosure** | After fix | Public disclosure with credit |

**Severity-based timelines**:
- **Critical**: Patch within 7 days
- **High**: Patch within 14 days
- **Medium**: Patch within 30 days
- **Low**: Patch in next minor release

---

## Vulnerability Severity Classification

We use [CVSS v3.1](https://www.first.org/cvss/calculator/3.1) scoring:

| Severity | CVSS Score | Description | Example |
|----------|------------|-------------|----------|
| **Critical** | 9.0 - 10.0 | Remote code execution | Unauthenticated RCE in WASM sandbox |
| **High** | 7.0 - 8.9 | Privilege escalation | Credential theft from vault |
| **Medium** | 4.0 - 6.9 | Information disclosure | API key leakage in logs |
| **Low** | 0.1 - 3.9 | Minor issues | Denial of service requiring local access |

---

## Security Measures

### Current Security Features

✅ **Implemented**:
- AES-256 encryption for credentials (`internal/vault`)
- WASM sandboxing for untrusted code (`internal/plugin`)
- Input validation using Go's `html/template` escaping
- Dependency pinning in `go.sum`
- CGO-based SQLite for memory safety

🚧 **In Development**:
- Agentic fuzzing for auto-generated code (`internal/fuzzer`)
- Circuit breakers for API abuse prevention (`internal/circuit`)
- Kill-switch for emergency shutdowns (`internal/killswitch`)
- Audit logging for all agent actions (`internal/audit`)

📋 **Planned**:
- mTLS for P2P mesh communication
- Rate limiting per API key
- Anomaly detection for hallucination loops
- Automated security scanning in CI/CD
- External security audit (pre-v1.0)

---

## Known Security Limitations

### Alpha Stage Warnings

⚠️ **NEXUS is in alpha. Expect vulnerabilities:**

1. **No external security audit** has been conducted
2. **Test coverage is 0%** - bugs are likely
3. **WASM sandbox may have escapes** - Wazero is experimental
4. **Credential encryption** - Key management is basic
5. **LLM prompt injection** - Not fully mitigated
6. **Agent actions are irreversible** - Rollback is conceptual

### Specific Risk Areas

| Component | Risk Level | Details |
|-----------|------------|----------|
| `internal/vault` | 🟡 Medium | AES-256 implemented but key rotation missing |
| `internal/plugin` | 🔴 High | WASM sandbox untested against exploits |
| `internal/browser` | 🔴 High | Headless browser may execute malicious JS |
| `internal/fuzzer` | 🟡 Medium | Fuzzer itself is untested |
| `internal/memory` | 🟢 Low | SQLite injection prevented by prepared statements |
| `internal/routing` | 🟢 Low | No PII handled, only routing logic |

---

## Security Best Practices for Users

### Do's ✅

- **Use separate API keys** for NEXUS (never your production keys)
- **Run in isolated environments** (Docker, VMs, sandboxes)
- **Enable audit logging** to track agent actions
- **Review agent-generated code** before execution
- **Keep dependencies updated** (`go get -u ./...`)
- **Use environment variables** for secrets (never hardcode)
- **Enable kill-switches** with cost/action thresholds

### Don'ts ❌

- **Don't run NEXUS as root** (use least privilege)
- **Don't expose port 7070** to the public internet
- **Don't disable WASM sandbox** for performance
- **Don't trust LLM outputs** without validation
- **Don't use in production** (it's alpha!)
- **Don't store sensitive data** in agent memory
- **Don't share `.env` files** or commit them to Git

---

## Security Roadmap

### Pre-v1.0 Requirements

Before declaring NEXUS production-ready, we commit to:

- [ ] **External security audit** by a reputable firm
- [ ] **Penetration testing** of all agent capabilities
- [ ] **Fuzzing suite** for WASM sandbox escapes
- [ ] **OWASP Top 10 compliance** audit
- [ ] **Dependency vulnerability scanning** in CI/CD
- [ ] **Secret scanning** (prevent accidental key commits)
- [ ] **SBOM generation** (Software Bill of Materials)
- [ ] **CVE monitoring** for all dependencies

### Security Milestones

| Milestone | Target | Status |
|-----------|--------|--------|
| CodeQL integration | v0.3.0-alpha | 📋 Planned |
| Trivy container scanning | v0.3.0-alpha | 📋 Planned |
| Dependabot alerts | v0.3.0-alpha | 📋 Planned |
| Basic fuzzing suite | v0.4.0-beta | 📋 Planned |
| First security audit | v0.9.0-rc | 📋 Planned |
| Bug bounty program | v1.0.0 | 📋 Planned |

---

## Bug Bounty Program

**Status**: 📋 Planned for v1.0.0 launch

We plan to launch a bug bounty program with rewards for:
- **Critical**: $500 - $2,000 USD
- **High**: $200 - $500 USD
- **Medium**: $50 - $200 USD
- **Low**: Public acknowledgment

*Bounties subject to change. Final terms TBD.*

---

## Security Hall of Fame

*This section will recognize security researchers who responsibly disclose vulnerabilities.*

No vulnerabilities reported yet. Be the first!

---

## Security Contact

**Primary Contact**: [61723019+Omkar0612@users.noreply.github.com](mailto:61723019+Omkar0612@users.noreply.github.com)

**PGP Key**: Coming soon

**Backup Contact**: GitHub Issues (for non-sensitive bugs only)

---

## Compliance & Standards

NEXUS aims to comply with:

- **OWASP Top 10** - Web application security
- **CWE Top 25** - Common weakness enumeration
- **NIST Cybersecurity Framework** - Security best practices
- **GDPR** - Data privacy (when applicable)
- **SOC 2 Type II** - Future goal for enterprise adoption

---

## Third-Party Security

### Dependency Security

We monitor dependencies using:
- GitHub Security Advisories
- Go vulnerability database (`govulncheck`)
- Dependabot (planned)

### Infrastructure Security

NEXUS is designed to run in:
- **Docker containers** (minimal attack surface)
- **Kubernetes** (with network policies)
- **Air-gapped environments** (no internet required for Ollama)

---

## Incident Response Plan

In case of a confirmed critical vulnerability:

1. **Immediate**: Disable affected features via kill-switch
2. **24 hours**: Release hotfix patch
3. **48 hours**: Publish security advisory (CVE if applicable)
4. **1 week**: Post-mortem and root cause analysis
5. **2 weeks**: Implement automated regression tests

---

## Legal

**Responsible Disclosure**: We follow responsible disclosure practices and will not take legal action against security researchers who:
- Report vulnerabilities privately
- Give us reasonable time to fix issues
- Do not exploit vulnerabilities maliciously

**Safe Harbor**: Security research conducted in good faith will not be considered a violation of applicable laws (CFAA, DMCA, etc.)

---

## Updates to This Policy

This security policy may be updated as NEXUS matures. Check back regularly.

**Last Updated**: March 2, 2026

---

*For implementation details, see [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)*
*For general questions, see [README.md](README.md) or [CONTRIBUTING.md](CONTRIBUTING.md)*
