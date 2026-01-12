### Get version

A GitHub Action and CLI tool to retrieve and filter software versions from multiple sources:
- **DockerHub** - Image tags
- **Maven Central** - Artifact versions
- **GitHub** - Releases and tags

Features powerful semantic version filtering with pattern matching and positional selection.

## Example

### GitHub Action Usage

```yaml
jobs:
  get-versions:
    runs-on: ubuntu-latest
    steps:
      - name: Get Latest Ubuntu version
        id: get-latest
        uses: scylladb-actions/get-version@v0.4.1
        with:
          source: dockerhub-imagetag
          repo: ubuntu
          filters: LAST

      - name: Get stable version (second-to-last)
        id: get-stable
        uses: scylladb-actions/get-version@v0.4.1
        with:
          source: dockerhub-imagetag
          repo: ubuntu
          filters: "[0-9]+.[0-9]+.LAST and LAST-1"

      - name: Print versions
        run: |
          echo "Latest: ${{ fromJson(steps.get-latest.outputs.versions)[0] }}"
          echo "Stable: ${{ fromJson(steps.get-stable.outputs.versions)[0] }}"
```

### CLI Usage

**Arguments:**
* `--source` - Version source: `dockerhub-imagetag`, `maven-artifact`, `github-release`, `github-tag`
* `--repo` - Repository name (e.g., `ubuntu`, `alpine/git`, `golang/go`)
* `--filters` - Filter pattern (see Filter Syntax below)
* `--out-format` - Output format: `text`, `json`, `yaml` (default: `text`)
* `--out-no-prefix` - Remove version prefix from output
* `--out-reverse-order` - Reverse sort order
* `--prefix` - Version prefix to match
* `--version` - Print CLI version and exit
* `--mvn-group` - Maven artifact group
* `--mvn-artifact-id` - Maven artifact ID

**Examples:**

```bash
# Get all Ubuntu versions from DockerHub
get-version --source dockerhub-imagetag --repo ubuntu

# Get the latest version
get-version --source dockerhub-imagetag --repo alpine --filters "LAST"

# Get second-newest version
get-version --source dockerhub-imagetag --repo alpine --filters "LAST-1"

# Get latest patch for each major.minor, output as JSON
get-version --source dockerhub-imagetag --repo alpine \
  --filters "[0-9]+.[0-9]+.LAST" --out-format json

# Get latest Go release from GitHub
get-version --source github-release --repo golang/go --filters "LAST"

# Get all 3.x versions from latest major
get-version --source dockerhub-imagetag --repo alpine --filters "LAST.*.*"
```

## Filter Syntax

The tool supports two types of filters that can be combined using `and` / `or` operators:

### 1. Pattern Filters (Component-Level)

Pattern filters match semantic version components (Major.Minor.Patch):

**Syntax:** `<major>.<minor>.<patch>`

**Pattern Types:**
- `*` - Match any value
- `LAST` / `LAST-N` - Select Nth newest component value
- `FIRST` / `FIRST+N` - Select Nth oldest component value
- `<number>` - Exact match (e.g., `5`)
- `<regex>` - Regular expression (e.g., `[0-9]+`)

**Examples:**
```bash
# Get all versions with major version 5
--filters "5.*.*"

# Get latest patch for each major.minor combination
--filters "[0-9]+.[0-9]+.LAST"

# Get all versions from the newest major version
--filters "LAST.*.*"

# Get the absolute newest version
--filters "LAST.LAST.LAST"

# Get all versions ending in .0
--filters "*.*.0"
```

### 2. Global Position Filters (List-Level)

Global position filters select a single version from the **entire version list** (not per-component).

**Syntax:** `LAST`, `LAST-N`, `FIRST`, `FIRST+N` (no dots)

**Key Difference:** These operate on the full sorted version list, making them robust across major version changes.

**Examples:**
```bash
# Get the absolute newest version
--filters "LAST"

# Get the second-newest version
--filters "LAST-1"

# Get the oldest version
--filters "FIRST"

# Get the third-oldest version
--filters "FIRST+2"
```

### 3. Combining Filters

Use `and` / `or` operators to chain filters:

**Important:** The `and` operator applies filters **sequentially**:
1. First filter reduces the version list
2. Second filter operates on the results

**Examples:**

```bash
# Get all .1 patch versions, then select the newest
--filters "*.*.1 and LAST"

# Get latest patch for each major.minor, then select second-to-last
# This pattern SURVIVES MAJOR VERSION CHANGES!
--filters "[0-9]+.[0-9]+.LAST and LAST-1"

# Get versions from major 5 OR major 6
--filters "5.*.* or 6.*.*"

# Complex: Get latest major, filter to .0 patches, select newest
--filters "LAST.*.0 and LAST"
```

### Use Case: Surviving Major Version Changes

**Problem:** `LAST.LAST-1.LAST` returns nothing when a new major version is released (e.g., 7.9.1 → 8.0.0)

**Solution:** Use global position filters with chaining:

```bash
# Before 8.0.0 release: Returns 7.8.x (second-to-last from [7.7.x, 7.8.x, 7.9.1])
# After 8.0.0 release:  Returns 7.9.1 (second-to-last from [7.7.x, 7.8.x, 7.9.1, 8.0.0])
--filters "[0-9]+.[0-9]+.LAST and LAST-1"
```

This approach maintains a stable version selection regardless of major version progressions.
