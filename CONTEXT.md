# Context

## Terms

- Commit workflow: the module that stages changes, reads repository context, generates a commit message, runs the edit or approve loop, commits, pushes, and optionally creates a pull request.
- Branch workflow: the module that stages changes, reads the staged diff, generates a branch name, sanitizes it, and creates the branch.
- Generation module: the module that builds prompts for commit workflow, branch workflow, and pull request generation, resolves provider fallback, and chooses the completion path.
- Repository module: the module that loads the staged snapshot, publishes the current branch, prepares pull request context, and creates branches or pull requests from repository facts.
- Change set: the workflow-owned view of staged changes, including diff text and prompt context, used at the workflow seam instead of repository module types.
- Pull request plan: the workflow-owned view of pull request readiness, including skip reason and combined diff, used at the workflow seam instead of repository module types.
- Resolved settings: the config seam output that combines defaults, file loading, legacy migration, environment overrides, provider selection, model selection, and config-file provenance into one immutable module.
- Repository context: branch name, changed file count, change summary, and staged file details used during commit message and pull request generation.
- YOLO mode: automatic staging, commit, and push with no confirmation.
