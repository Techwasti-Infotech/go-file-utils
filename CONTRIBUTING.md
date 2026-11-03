# Contributing to go-file-utils

Thank you for your interest in contributing to `go-file-utils`! We welcome contributions, bug fixes, feature requests, and documentation improvements.

---

## Code of Conduct
We are committed to providing a welcoming, inclusive, and harassment-free environment for all contributors.

---

## Development Setup

### Prerequisites
- Go 1.20 or newer installed
- Git installed and configured

### Clone the Repository
```bash
git clone https://github.com/Techwasti-Infotech/go-file-utils.git
cd go-file-utils
```

### Running Tests
Execute the comprehensive test suite with coverage:
```bash
go test -v -race -cover ./...
```

Verify formatting:
```bash
go fmt ./...
go vet ./...
```

---

## Contribution Guidelines

1. **Zero External Dependencies**:
   - `go-file-utils` does not use third-party runtime libraries. All additions must rely on standard Go libraries.
2. **Comprehensive Unit Tests**:
   - Any new helper or modification must include comprehensive unit tests verifying both happy paths and edge/error conditions.
3. **Cross-Platform Compatibility**:
   - Make sure your code operates seamlessly on Linux, macOS, and Windows. Use `filepath` instead of hardcoding slashes `/` or backslashes `\`.
4. **Security by Design**:
   - Always sanitize paths before operating on archives or user-controlled input.
5. **Incremental Commits**:
   - Use descriptive commit messages following the Conventional Commits convention:
     - `feat(...)`: New features
     - `fix(...)`: Bug fixes
     - `docs(...)`: Documentation updates
     - `test(...)`: Adding or updating test cases
     - `refactor(...)`: Code cleanup with no functional change

---

## Pull Request Process

1. Fork the repository and create a branch for your feature or bug fix:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Implement your changes along with tests and documentation.
3. Verify all tests pass: `go test -v ./...`
4. Commit your changes and push to your fork.
5. Open a Pull Request on GitHub against the `main` branch.
6. Address any feedback or review suggestions.

---

## Contact
For security disclosures or questions, please contact the maintainers at **techwastiinfotech@gmail.com**.
