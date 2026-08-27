# Contributing to discord_notify

Thank you for contributing to `discord_notify`! This project follows the **GitHub Flow** development model.

---

## 🌿 GitHub Flow Workflow

1. **Create a branch from `main`**:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes**:
   - Write clean, formatted Go code (`make fmt`).
   - Run linter checks (`make lint`).
   - Add unit tests for any new features or bug fixes.

3. **Verify tests and build**:
   ```bash
   make test
   make build
   ```

4. **Update the Changelog**:
   - Document any notable additions, changes, or fixes in [`CHANGELOG.md`](./CHANGELOG.md) under the `[Unreleased]` section.

5. **Commit and Merge**:
   - Commit your changes with descriptive commit messages.
   - Open a Pull Request or merge into `main` using **normal merge (`--no-ff`)**:
     ```bash
     git checkout main
     git merge --no-ff feature/my-new-feature
     ```

6. **Delete Merged Branch**:
   - Clean up the local branch after merging:
     ```bash
     git branch -d feature/my-new-feature
     ```

---

## 🏷️ Branch Naming Conventions

- `feature/<name>` : New features or capabilities
- `fix/<name>`     : Bug fixes
- `refactor/<name>`: Code restructuring without feature changes
- `docs/<name>`    : Documentation updates or improvements
