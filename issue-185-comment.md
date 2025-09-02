## Update: Manual Signing Solution Found and Implemented 🎉

I've discovered that **manual signing works perfectly** with SignPath's free tier! Here's what I learned and implemented:

### The Discovery

While the automated GitHub Actions integration fails with our free SignPath account (due to the "repository-based policy" limitation), I found that:
1. Manual upload through SignPath's web interface works flawlessly
2. The signed installer eliminates Windows security warnings
3. The process only adds ~5 minutes to our release workflow

### What I've Done

1. **Documented the manual process** in `docs/devel/release-guide.md`
   - Step-by-step instructions for signing after automated release
   - Clear explanation of why manual signing is needed
   - Instructions for updating the release with signed installer

2. **Created a helper script** (`scripts/upload-signed-windows-installer.sh`)
   - Automates uploading the signed installer to GitHub
   - Provides clear next steps for updating release notes
   - Makes the process consistent and repeatable

3. **Added clarifying comments** to the workflow
   - Explained why automated signing is currently disabled
   - Referenced the manual signing documentation

### The Manual Process (Summary)

1. After release workflow completes, download `GoPCA-Setup-vX.X.X.exe`
2. Upload to SignPath.io web interface for signing
3. Download signed file and rename to `GoPCA-Setup-vX.X.X-signed.exe`
4. Use our helper script: `./scripts/upload-signed-windows-installer.sh vX.X.X signed-file.exe`
5. Update release notes to point to signed version

### Example

I successfully tested this with v1.0.1:
- Signed installer: `GoPCA-Setup-v1.0.1-signed.exe`
- No more Windows security warnings! ✅

### Recommendation

This manual process is a **good long-term solution** for our needs:
- Works reliably with free SignPath account
- Only adds 5 minutes to release process
- Provides proper code signing without cost
- Can be upgraded to automated process if we ever need paid SignPath features

Unless we're doing very frequent releases, the manual process is perfectly acceptable and saves the cost of a paid SignPath subscription.

### Files Changed
- `docs/devel/release-guide.md` - Added Windows signing documentation
- `.github/workflows/release.yml` - Added explanatory comment
- `scripts/upload-signed-windows-installer.sh` - New helper script

The changes are in PR #[number] if you want to review them.

I consider this issue effectively resolved - we have a working solution that meets our needs! 🚀