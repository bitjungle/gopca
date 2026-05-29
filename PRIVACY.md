# Privacy Policy & Technical Verification

## Our Privacy Commitment

GoPCA Suite is designed with **absolute privacy** as a fundamental principle. We guarantee:

- ✅ **100% Local Processing** - All computations happen exclusively on your machine
- ✅ **Zero Telemetry** - No usage data, metrics, or analytics collection
- ✅ **No External Connections** - Complete offline operation capability
- ✅ **No Data Transmission** - Your data never leaves your computer
- ✅ **Source Transparency** - Every line of code is publicly viewable and auditable

## Framework Privacy Analysis

### Verified Privacy-Respecting Frameworks (January 2025)

| Framework/Library | Version | Telemetry Status | Verification Method |
|-------------------|---------|------------------|---------------------|
| **Wails** | v2.10.2 | ✅ None | Source code audit, no telemetry mechanisms found |
| **Vite** | v7.1.2 | ✅ None | No built-in telemetry or analytics |
| **React** | v18.2.0 | ✅ None | Library itself has no telemetry |
| **Plotly.js** | v2.35.3 | ✅ Offline-only | Version 4+ operates offline by default |
| **Gonum** | v0.16.0 | ✅ None | No telemetry mechanisms found |
| **TypeScript** | v5.9.2 | ✅ None | Development tool, no runtime telemetry |
| **Go Compiler** | v1.24+ | ✅ Opt-in only | Telemetry OFF by default, not in binaries |

### Development-Only Considerations

During development (not affecting end users):
- **npm registry**: Sends minimal headers (`Npm-Scope`, `Npm-In-CI`) during package installation
- **Go telemetry**: Opt-in only, disabled by default, never included in compiled binaries

## Technical Verification

### How to Verify Our Privacy Claims

#### 1. Network Traffic Analysis
```bash
# Monitor all network traffic while using GoPCA
sudo tcpdump -i any -w gopca_traffic.pcap host not 127.0.0.1

# Run GoPCA and use all features
# Then analyze the capture file
tcpdump -r gopca_traffic.pcap

# Expected result: No external connections
```

#### 2. DNS Query Monitoring
```bash
# Monitor DNS queries on macOS
sudo dscacheutil -statistics

# Monitor DNS queries on Linux
sudo tcpdump -i any port 53

# Run GoPCA
# Expected result: No DNS lookups to external services
```

#### 3. Firewall Verification
```bash
# Block all network access for GoPCA
# macOS: Use Little Snitch or LuLu
# Windows: Use Windows Firewall
# Linux: Use iptables or ufw

# GoPCA should work perfectly with network blocked
```

#### 4. Process Network Inspection
```bash
# macOS/Linux
lsof -i -P | grep -i gopca
# Expected: No output (no network connections)

# Windows
netstat -an | findstr gopca
# Expected: No output
```

### Automated Privacy Verification

Run our privacy verification script:
```bash
./scripts/verify-privacy.sh
```

This script automatically:
- Scans for telemetry-related dependencies
- Checks for external URLs in code
- Verifies no network calls in binaries
- Generates a privacy audit report

## Data Handling

### What We DON'T Collect
- ❌ Usage statistics or metrics
- ❌ Error reports or crash data
- ❌ User behavior or interactions
- ❌ System information
- ❌ File names or data content
- ❌ IP addresses or location data
- ❌ Any form of identifier

### What Stays on Your Machine
- ✅ All your data files
- ✅ Analysis results
- ✅ Configuration settings
- ✅ Application logs (if any)
- ✅ Temporary processing files

## Security Features

### Local-Only Architecture
- No cloud services or APIs
- No external authentication
- No remote configuration
- No update checks
- No license validation

### Code Signing & Integrity
- macOS apps are signed and notarized for security
- Windows binaries include checksums
- All releases include SHA-256 hashes

## Developer Guidelines

### Maintaining Privacy Standards

When contributing to GoPCA:

1. **Never add telemetry libraries**
   - No analytics (Google Analytics, Mixpanel, etc.)
   - No crash reporting (Sentry, Rollbar, etc.)
   - No usage tracking

2. **Avoid external dependencies**
   - Prefer standard library over external packages
   - Audit new dependencies for telemetry
   - Document why each dependency is necessary

3. **No network calls**
   - No HTTP/HTTPS requests
   - No WebSocket connections
   - No DNS lookups
   - Exception: User-initiated export/share features

4. **Privacy-first features**
   - All processing must be local
   - User data never leaves the machine
   - Clear user consent for any future network features

### Dependency Audit Process

Before adding new dependencies:
```bash
# For npm packages
npm view [package] | grep -i "telemetry\|analytics\|tracking"

# For Go modules
go get -u [module]
grep -r "http\.\|telemetry" [module-source]
```

## Compliance & Standards

### Privacy Regulations
GoPCA's architecture inherently complies with:
- **GDPR** - No personal data processing
- **CCPA** - No data collection or sale
- **HIPAA** - No data transmission risks
- **Corporate policies** - No cloud dependency

### Industry Standards
- Follows privacy-by-design principles
- Zero-knowledge architecture
- Local-first computing model
- Offline-first operation

## Verification Tools

### For Users
1. **Network Monitor** - Use system tools to verify no connections
2. **Firewall Test** - Block network and verify full functionality
3. **Privacy Script** - Run `./scripts/verify-privacy.sh`

### For Developers
1. **CI/CD Checks** - Automated privacy verification on every build
2. **Dependency Audit** - Regular scanning for telemetry
3. **Binary Analysis** - Verify no network code in releases

## Transparency Reports

### Audit History
- **January 2025**: Comprehensive framework audit completed
- All frameworks verified as privacy-respecting
- No telemetry or data collection found

### Continuous Monitoring
- Automated CI/CD privacy checks
- Regular dependency audits
- Community-reported verification

## Questions & Verification

### How to Verify Yourself
1. Download source code from GitHub
2. Run privacy verification scripts
3. Monitor network traffic during use
4. Review our automated test results

### Report Privacy Concerns
If you discover any privacy issue:
1. Create a GitHub issue (public or private)
2. Include verification steps
3. We commit to investigate within 48 hours

## License & Rights

This privacy policy is part of the GoPCA Suite documentation.
- Source code: GoPCA Suite Source-Available Freeware License — see `LICENSE` for full terms
- Your data: Remains 100% yours
- No rights claimed over user data

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-09-15 | Initial privacy policy and verification guide |

---

*Last updated: September 2025*
*This is a living document and will be updated as needed.*