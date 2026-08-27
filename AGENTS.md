# Development Guidelines & Rules

## 1. Branching Model: GitHub Flow
This repository follows the **GitHub Flow** branching strategy:

1. **`main` branch is always stable and deployable**:
   - Do NOT commit directly to `main` for feature development or bug fixes.
   - All tests (`make test`) and builds (`make build`) must pass on `main`.

2. **Feature Branching**:
   - Always create a new branch from `main` before starting any work.
   - Branch naming conventions:
     - `feature/<feature-name>`: New features and enhancements
     - `fix/<bug-name>`: Bug fixes
     - `refactor/<refactor-name>`: Code refactoring
     - `docs/<doc-name>`: Documentation updates

3. **Development & Verification**:
   - Implement changes in the feature branch.
   - Add/update unit tests for all new functionality.
   - Always verify with `make test` and `make build`.

4. **ChangeLog Management**:
   - Update [`CHANGELOG.md`](./CHANGELOG.md) under the `[Unreleased]` section following the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) standard.

5. **Staging & User Confirmation**:
   - Stage changes (`git add`) and present the proposed diff/summary to the user before committing.
   - Ask for confirmation before merging branches into `main`.

6. **Merging**:
   - Merge back into `main` using **Normal Merge (`git merge --no-ff`)** to maintain clear release and feature commit history.

7. **Branch Cleanup**:
   - Always delete the working branch after it has been successfully merged into `main` (`git branch -d <branch-name>`).

