# Test Cases for Graph Rendering

These test cases define expected graph output for various git topologies. Each includes the commit topology (parent relationships) and expected rendering. Use these as golden file test inputs during implementation.

**Key rule:** Every commit row has a `├` marker showing which column/branch the commit belongs to.

## Test Case 1: Linear History (No Branches)

**Topology:**
```
A ← B ← C ← D (HEAD)
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| D | C | HEAD -> main |
| C | B | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├ D (HEAD -> main) Latest commit - Alice (1 min ago)
  ├ C Third commit - Bob (2 min ago)
  ├ B Second commit - Alice (3 min ago)
  ├ A Initial commit - Alice (1 day ago)
```

## Test Case 2: Simple Feature Branch Merge

**Topology:**
```
    C ← D (feature, merged)
   /     \
A ← B ←───← E (HEAD -> main)
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| E | B, D | HEAD -> main |
| D | C | |
| C | A | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─╮ E (HEAD -> main) Merge feature - Alice (1 min ago)
  │ ├ D Feature complete - Bob (2 min ago)
  │ ├ C Feature work - Bob (3 min ago)
  ├ │ B Main work - Alice (4 min ago)
  ├─╯ A Initial commit - Alice (1 day ago)
```

## Test Case 3: Two Parallel Feature Branches (Octopus Merge)

**Topology:**
```
    C ← D (feature-1)
   /     \
A ←───────← G (HEAD -> main)
   \     /
    E ← F (feature-2)
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| G | A, D, F | HEAD -> main |
| F | E | |
| E | A | |
| D | C | |
| C | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─┬─╮ G (HEAD -> main) Merge features - Alice (1 min ago)
  │ ├ │ D Feature-1 done - Bob (2 min ago)
  │ ├ │ C Feature-1 work - Bob (3 min ago)
  │ │ ├ F Feature-2 done - Carol (2 min ago)
  │ │ ├ E Feature-2 work - Carol (3 min ago)
  ├─┴─╯ A Initial commit - Alice (1 day ago)
```

## Test Case 4: Sequential Merges

**Topology:**
```
    B ← C (feature-1)
   /     \
A ←───────← D ← G (HEAD -> main)
             \   /
              E─F (feature-2)
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| G | D, F | HEAD -> main |
| F | E | |
| E | D | |
| D | A, C | |
| C | B | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─╮ G (HEAD -> main) Merge feature-2 - Alice (1 min ago)
  │ ├ F Feature-2 done - Carol (2 min ago)
  │ ├ E Feature-2 work - Alice (3 min ago)
  ├─┤ D Merge feature-1 - Alice (4 min ago)
  │ ├ C Feature-1 done - Bob (5 min ago)
  │ ├ B Feature-1 work - Bob (6 min ago)
  ├─╯ A Initial commit - Alice (1 day ago)
```

## Test Case 5: Sequential Merges (with commits on main)

**Topology:**
```
A ← B ← E ← F ← H (HEAD -> main)
     \   /   \   /
      C─D     G
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| H | F, G | HEAD -> main |
| G | F | |
| F | E | |
| E | B, D | |
| D | C | |
| C | B | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─╮ H (HEAD -> main) Merge feature-2 - Alice (1 min ago)
  │ ├ G Feature-2 work - Carol (2 min ago)
  ├─╯ F Continue main - Alice (3 min ago)
  ├─╮ E Merge feature-1 - Alice (4 min ago)
  │ ├ D Feature-1 done - Bob (5 min ago)
  │ ├ C Feature-1 work - Bob (6 min ago)
  ├ │ B Main work - Alice (7 min ago)
  ├─╯ A Initial commit - Alice (1 day ago)
```

## Test Case 6: Root Commit (End of History)

**Topology:**
```
A (root, no parents)
```

**Expected Graph:**
```
> ├ A (HEAD -> main) Initial commit - Alice (1 day ago)
```

No continuation line below root commit.

## Test Case 7: Multiple Roots (Merged Repositories)

**Topology:**
```
A ← B ← D (HEAD -> main)
        ↑
C ──────┘ (was separate repo)
```

**Commits:**
| Hash | Parents | Refs |
|------|---------|------|
| D | B, C | HEAD -> main |
| C | (none) | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─╮ D (HEAD -> main) Merge repos - Alice (1 min ago)
  │ ├ C Other repo init - Bob (1 year ago)
  ├ B Main work - Alice (2 min ago)
  ├ A Initial commit - Alice (1 day ago)
```

Note: Column 1 has no continuation after C (it's a root), so after row C the graph narrows.

## Test Case 8: With Tags and Remote Refs

**Commits:**
| Hash | Parents | Refs |
|------|---------|------|
| C | B | HEAD -> main, origin/main |
| B | A | tag: v1.0 |
| A | (none) | |

**Expected Graph:**
```
> ├ C (HEAD -> main, origin/main) Latest - Alice (1 min ago)
  ├ B (tag: v1.0) Release - Alice (1 week ago)
  ├ A Initial commit - Alice (1 month ago)
```

## Test Case 9: Complex usecase

**Topology:**
```
                K ← L (feature-3, long-lived)
               /               \
      C    E  /  G ←─── I       \
     / \  / \/  / \      \       \
  A─B───D────F─────H──────J───────M (HEAD -> main)
```

**Commits (most recent first):**
| Hash | Parents | Refs |
|------|---------|------|
| M | J, L | HEAD -> main |
| L | K | |
| K | D | |
| J | H, I | |
| I | G | |
| H | F, G | |
| G | F | |
| F | D, E | |
| E | D | |
| D | B, C | |
| C | B | |
| B | A | |
| A | (none) | |

**Expected Graph:**
```
> ├─╮ M (HEAD -> main) Merge feature-3 - Alice (1 min ago)
  │ ├ L Feature-3 done- Carol (1 min ago)
  │ ├ K Feature-3 - Carol (2 min ago)
  ├─│─╮ J Merged Feature 2 - Alice (3 min ago)
  │ │ ├ I Feature 2 done - Carol (4 min ago)
  ├─│─│─╮ H Early Merge Feature 2 - Alice (5 min ago)
  │ │ ├─╯ G Feature-2 done - Bob (6 min ago)
  ├─│─┤ F Merged Hotfix Feature 1 - Alice (7 min ago)
  │ │ ├ E Hotfix Feature 1 - Alice (8 min ago)
  ├─┼─╯ D Merged Feature 1 - Diana (9 min ago)
  │ ├ C Feature 1 - Alice (9 min ago)
  ├─╯ B Main work - Alice (10 min ago)
  ├ A Initial commit- Alice (1 day ago)
```

