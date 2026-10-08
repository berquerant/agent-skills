# Platform Tools and Integrations Guide

The `multi-repo-change` skill is code-hosting-agnostic, supporting standard platforms like GitHub and GitLab. Use the best available tool for the detected hosting service.

---

## 1. Tool Selection Priority

For remote platform operations (fetching project info, opening Pull/Merge Requests, updating descriptions, adding labels):

1. **Dedicated MCP Server** (e.g. `github` MCP server, `gitlab` MCP server)
   - Preferred when registered and available in the agent environment.
   - Inspect available MCP tool definitions before invoking; never guess nonexistent tool names.
2. **Platform Official CLI** (`gh` for GitHub, `glab` for GitLab)
   - Use standard CLI commands when installed and authenticated on the machine.
3. **Web URL Fallback**
   - If neither MCP nor CLI is configured or authenticated, do **NOT** attempt undocumented raw HTTP requests or insecure token extraction.
   - Push the branch and provide the web compare / PR creation URL for the user to open manually.

---

## 2. Platform Adapters

### 2.1 GitHub (`github.com` or GitHub Enterprise)
- **Detection**:
  - Remote URL points to `github.com` or an enterprise GitHub host.
- **MCP Tools**:
  - Look for tools under `github_*` (e.g., `github_create_pull_request`, `github_update_pull_request`).
- **CLI Commands (`gh`)**:
  - Check auth: `gh auth status`
  - Create Draft PR:
    ```bash
    gh pr create \
      --repo <owner/repo> \
      --head <branch> \
      --base <default-branch> \
      --title "<title>" \
      --body "<description>" \
      --draft
    ```
  - Update PR body (for cross-linking):
    ```bash
    gh pr edit <pr-number-or-url> --body "<updated-description>"
    ```
- **Templates**:
  - Check `.github/PULL_REQUEST_TEMPLATE.md` or `.github/PULL_REQUEST_TEMPLATE/`.

---

### 2.2 GitLab (`gitlab.com` or Self-hosted GitLab)
- **Detection**:
  - Remote URL points to `gitlab.com` or a self-hosted GitLab domain.
- **MCP Tools**:
  - Look for tools under `gitlab_*` (e.g., `gitlab_create_merge_request`, `gitlab_update_merge_request`).
- **CLI Commands (`glab`)**:
  - Check auth: `glab auth status`
  - Create Draft MR:
    ```bash
    glab mr create \
      --repo <group/project> \
      --source-branch <branch> \
      --target-branch <default-branch> \
      --title "Draft: <title>" \
      --description "<description>"
    ```
  - Update MR description (for cross-linking):
    ```bash
    glab mr update <mr-iid> --description "<updated-description>"
    ```
- **Templates**:
  - Check `.gitlab/merge_request_templates/`.

---

## 3. Cross-Linking Change Requests

Once all Change Requests are opened, update their descriptions to cross-reference each other. This gives reviewers complete context across dependent or coordinated changes.

### Cross-Linking Markdown Format

Append a section to the end of each PR/MR description (or update the designated section in existing templates):

```markdown
## Related Change Requests

- `<repo-1>`: <PR/MR-URL-1>
- `<repo-2>`: <PR/MR-URL-2>
- `<repo-3>`: <PR/MR-URL-3>
```

Include all related PRs/MRs in each description to maintain bidirectional traceability.
