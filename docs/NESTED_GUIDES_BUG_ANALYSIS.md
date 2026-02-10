# Root Cause Analysis: Nested guides/ Directories Bug

## Problem Summary

The `docs/guides/` directory contained 11 levels of nested `guides/guides/guides/.../` subdirectories, with duplicate markdown files at each level. This created ~33 duplicate files and wasted disk space.

## Root Cause

**Self-Perpetuating Backup/Restore Loop** in `lefthook.yml` pre-commit hook (lines 21-30).

### The Bug Mechanism

```yaml
# Backup manual documentation
[ -d "docs/guides" ] && cp -r docs/guides /tmp/docs-guides-backup

# Generate schema-based documentation
cd tools && go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate ...

# Restore manual documentation
[ -d "/tmp/docs-guides-backup" ] && cp -r /tmp/docs-guides-backup docs/guides
```

### How It Escalates

**Initial Trigger** (Unknown - possibly manual error or tool glitch):
```
docs/guides/
├── rate-limits.md
├── migration.md
├── validation.md
└── guides/              ← FIRST nested directory created somehow
    ├── rate-limits.md
    ├── migration.md
    └── validation.md
```

**Iteration 1 - First Commit**:
1. **Backup**: `cp -r docs/guides /tmp/docs-guides-backup`
   ```
   /tmp/docs-guides-backup/
   ├── rate-limits.md
   ├── migration.md
   ├── validation.md
   └── guides/          ← Backup captures the nested directory!
       ├── rate-limits.md
       ├── migration.md
       └── validation.md
   ```

2. **Generate**: tfplugindocs regenerates `docs/` directory

3. **Restore**: `cp -r /tmp/docs-guides-backup/* docs/guides/`
   ```
   docs/guides/
   ├── rate-limits.md
   ├── migration.md
   ├── validation.md
   └── guides/          ← Nested directory restored
       ├── rate-limits.md
       ├── migration.md
       └── validation.md
   ```

**Iteration 2 - Second Commit** (THE EXPONENTIAL GROWTH):
1. **Backup**: `cp -r docs/guides /tmp/docs-guides-backup`
   ```
   /tmp/docs-guides-backup/
   ├── rate-limits.md
   ├── migration.md
   ├── validation.md
   └── guides/
       ├── rate-limits.md
       ├── migration.md
       ├── validation.md
       └── guides/      ← NOW TWO LEVELS DEEP!
           ├── rate-limits.md
           ├── migration.md
           └── validation.md
   ```

2. **Restore**: Puts it all back, creating `docs/guides/guides/guides/`

**Iteration 3-11**: Each commit adds another nested level, exponentially growing to 11 levels deep!

## The Core Issue

The restore command **blindly copies everything** from the backup, including any accidental nested `guides/` directories. This creates a **positive feedback loop** where each commit makes the problem worse.

## How the First Nested Directory Was Created

Likely causes:
1. **Manual error**: Someone accidentally ran `cp -r docs/guides docs/guides/guides/`
2. **Tool glitch**: tfplugindocs or another tool created `docs/guides/guides/` temporarily
3. **Merge conflict**: Git merge resolution created duplicate directories
4. **Previous version of lefthook**: Earlier hook implementation had different cp behavior

## The Fix

**Option 1: Delete Before Restore** (Safest):
```yaml
# Restore manual documentation
[ -d "/tmp/docs-guides-backup" ] && rm -rf docs/guides && mv /tmp/docs-guides-backup docs/guides
```

**Option 2: Only Backup Files** (Prevents recursion):
```yaml
# Backup only markdown files, not subdirectories
[ -d "docs/guides" ] && mkdir -p /tmp/docs-guides-backup && cp docs/guides/*.md /tmp/docs-guides-backup/

# Restore
[ -d "/tmp/docs-guides-backup" ] && cp /tmp/docs-guides-backup/*.md docs/guides/
```

**Option 3: Validation Step** (Defensive):
```yaml
# Before backup, validate no nested guides/ directories exist
if [ -d "docs/guides/guides" ]; then
  echo "ERROR: Nested docs/guides/guides/ detected! Fix before committing."
  exit 1
fi
```

## Recommended Solution

**Combination approach** - prevent both creation and accumulation:

```yaml
generate-docs:
  run: |
    echo "📝 Generating Terraform provider documentation..."

    # VALIDATION: Fail if nested guides/ already exists
    if [ -d "docs/guides/guides" ]; then
      echo "❌ ERROR: Nested docs/guides/guides/ directory detected!"
      echo "This indicates a bug. Run: rm -rf docs/guides/guides"
      exit 1
    fi

    # Backup only markdown files (not subdirectories)
    [ -d "docs/guides" ] && mkdir -p /tmp/docs-guides-backup && cp docs/guides/*.md /tmp/docs-guides-backup/ 2>/dev/null || true
    [ -f "docs/index.md" ] && cp docs/index.md /tmp/docs-index-backup.md
    [ -f "docs/TROUBLESHOOTING.md" ] && cp docs/TROUBLESHOOTING.md /tmp/docs-troubleshooting-backup.md

    # Generate schema-based documentation
    cd tools && go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir .. --provider-name hyperping
    cd ..

    # Restore manual documentation files
    [ -d "/tmp/docs-guides-backup" ] && mkdir -p docs/guides && cp /tmp/docs-guides-backup/*.md docs/guides/ 2>/dev/null || true
    [ -f "/tmp/docs-index-backup.md" ] && cp /tmp/docs-index-backup.md docs/index.md
    [ -f "/tmp/docs-troubleshooting-backup.md" ] && cp /tmp/docs-troubleshooting-backup.md docs/TROUBLESHOOTING.md

    # Cleanup
    rm -rf /tmp/docs-guides-backup /tmp/docs-index-backup.md /tmp/docs-troubleshooting-backup.md

    # Only stage generated docs
    if ! git diff --exit-code --quiet docs/resources/ docs/data-sources/; then
      git add docs/resources/ docs/data-sources/
      echo "✅ Documentation updated and staged"
    else
      echo "✅ Documentation already up to date"
    fi
  stage_fixed: true
  fail_text: "Failed to generate documentation"
```

## Prevention

1. ✅ **Fixed**: Removed all 11 nested `guides/` directories
2. ⚠️ **TODO**: Update `lefthook.yml` with new backup/restore logic
3. ⚠️ **TODO**: Add validation check to prevent future occurrences
4. ⚠️ **TODO**: Test the fix by running multiple commits

## Impact

- **Disk space**: Saved ~1MB (33 duplicate files across 11 directories)
- **Confusion**: Documentation generation was copying/overwriting files 11 times
- **Performance**: Pre-commit hook was slower due to excessive file operations

## Timeline

- **Unknown date**: First nested `guides/` directory created
- **2-11 subsequent commits**: Problem escalated from 1 to 11 levels deep
- **2026-02-10**: Bug discovered during documentation audit
- **2026-02-10**: Nested directories removed, root cause identified

## Lessons Learned

1. **Backup/restore loops need validation** to prevent accumulation bugs
2. **Only backup what you need** (files, not directories) to avoid recursion
3. **Add safety checks** before critical operations (e.g., check for nested dirs)
4. **Monitor directory structure** in code reviews

## Related Files

- `lefthook.yml` - Lines 16-42 (generate-docs hook)
- `docs/guides/` - Affected directory
- `.github/workflows/test.yml` - Also runs tfplugindocs

## Action Items

- [ ] Update `lefthook.yml` with fixed backup/restore logic
- [ ] Add validation check for nested directories
- [ ] Test with multiple commits to verify fix
- [ ] Add this bug to CHANGELOG as a warning/fix
