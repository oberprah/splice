# Truncation Strategy Examples

This document shows the progressive truncation strategy applied to realistic examples.

## Example Components

**Starting with:**
- Selector: `"> "`
- Graph: `"├ "`
- Hash: `"abc123d"`
- Refs: `"(HEAD -> feature/implement-advanced-user-authentication-system, origin/feature/implement-advanced-user-authentication-system, tag: v2.1.0-beta.1) "`
- Message: `"Implement comprehensive user authentication system with OAuth2 support, JWT tokens, refresh token rotation, and secure password hashing"`
- Author: `"Maximilian Alexander von Mustermann"`
- Time: `"(11 months ago)"`

## Progressive Truncation Steps

### Step 0: Very wide terminal (300 chars) - Everything fits
```
> ├ abc123d (HEAD -> feature/implement-advanced-user-authentication-system, origin/feature/implement-advanced-user-authentication-system, tag: v2.1.0-beta.1) Implement comprehensive user authentication system with OAuth2 support, JWT tokens, refresh token rotation, and secure password hashing - Maximilian Alexander von Mustermann (11 months ago)
```
**Length:** ~298 chars ✓

---

### Step 1: Terminal 180 chars - Cap message at 72 chars
```
> ├ abc123d (HEAD -> feature/implement-advanced-user-authentication-system, origin/feature/implement-advanced-user-authentication-system, tag: v2.1.0-beta.1) Implement comprehensive user authentication system with OAuth2 su... - Maximilian Alexander von Mustermann (11 months ago)
```
**Length:** ~237 chars (still too long, message now capped at 72)
**Applied:** Rule 1 - cap message at 72 chars

---

### Step 2: Terminal 180 chars - Truncate author to 25 chars
```
> ├ abc123d (HEAD -> feature/implement-advanced-user-authentication-system, origin/feature/implement-advanced-user-authentication-system, tag: v2.1.0-beta.1) Implement comprehensive user authentication system with OAuth2 su... - Maximilian Alexander v... (11 months ago)
```
**Length:** ~229 chars (still too long)
**Applied:** Rule 2 - truncate author to 25 chars

---

### Step 3: Terminal 170 chars - Shorten refs (Level 1: shorten individual names)
```
> ├ abc123d (HEAD -> feature/implement-advanced…, origin/feature/implement-advanced…, tag: v2.1.0-beta.1) Implement comprehensive user authentication system with OAuth2 su... - Maximilian Alexander v... (11 months ago)
```
**Length:** ~200 chars (still too long)
**Applied:** Rule 3.1 - shorten individual ref names

---

### Step 4: Terminal 140 chars - Shorten refs (Level 2: show first + count)
```
> ├ abc123d (HEAD -> feature/implement-advanced… +2 more) Implement comprehensive user authentication system with OAuth2 su... - Maximilian Alexander v... (11 months ago)
```
**Length:** ~160 chars (still too long)
**Applied:** Rule 3.2 - show first ref + count

---

### Step 5: Terminal 120 chars - Shorten refs (Level 3: count only)
```
> ├ abc123d (3 refs) Implement comprehensive user authentication system with OAuth2 su... - Maximilian Alexander v... (11 months ago)
```
**Length:** ~128 chars (still too long)
**Applied:** Rule 3.3 - show count only

---

### Step 6: Terminal 100 chars - Truncate author to 5 chars
```
> ├ abc123d (3 refs) Implement comprehensive user authentication system with OAuth2 su... - Maxim... (11 months ago)
```
**Length:** ~107 chars (still too long)
**Applied:** Rule 4 - truncate author to 5 chars

---

### Step 7: Terminal 90 chars - Drop time
```
> ├ abc123d (3 refs) Implement comprehensive user authentication system with OAuth2 su... - Maxim...
```
**Length:** ~90 chars ✓
**Applied:** Rule 5 - drop time

---

### Step 8: Terminal 80 chars - Shorten message to 40 chars
```
> ├ abc123d (3 refs) Implement comprehensive user auth... - Maxim...
```
**Length:** ~59 chars ✓
**Applied:** Rule 6 - shorten message to 40 chars

---

### Step 9: Terminal 60 chars - Drop author
```
> ├ abc123d (3 refs) Implement comprehensive user auth...
```
**Length:** ~50 chars ✓
**Applied:** Rule 7 - drop author

---

### Step 10: Terminal 50 chars - Truncate message from right
```
> ├ abc123d (3 refs) Implement comprehe...
```
**Length:** ~43 chars ✓
**Applied:** Rule 8 - continue truncating message

---

### Step 11: Terminal 40 chars - Continue truncating
```
> ├ abc123d (3 refs) Implem...
```
**Length:** ~33 chars ✓
**Applied:** Rule 8 - continue truncating message

---

### Step 12: Terminal 30 chars - Drop refs, minimal message
```
> ├ abc123d Implement...
```
**Length:** ~25 chars ✓
**Applied:** Rule 8 - dropped refs, keep minimal message

---

## Additional Scenarios

### Scenario A: Large git graph with long branch names (80 chars)
```
> ├─┬─╮ abc123d (3 refs) Fix authentication bug... - Alice (2 days ago)
```
Graph takes more space (8 chars vs 2), but we accept it as mandatory.

---

### Scenario B: Short components, plenty of space (80 chars)
```
> ├ abc123d (main) Fix bug - Alice Johnson (2 days ago)
```
Everything fits without any truncation.

---

### Scenario C: No refs (80 chars)
```
> ├ abc123d Implement comprehensive user authentication system with OAuth2... - Alice (2 days ago)
```
More space available for message when no refs present.

---

## Key Observations

1. **Message preservation**: We keep as much of the commit message as possible since it's the most important info
2. **Refs handling**: Three-level degradation provides good balance between info and space
3. **Author flexibility**: Can be heavily truncated or dropped since it's less critical in most workflows
4. **Time optional**: Time ago is nice-to-have but can be dropped
5. **Graph priority**: We never truncate graph symbols, accepting they can be large with many branches
