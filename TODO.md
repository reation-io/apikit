# APIKit - TODO List

**Version:** 1.0.0

---

## 🔴 Critical Issues (3)

- [ ] Snake case edge cases not covered
- [ ] HTTP status codes hardcoded
- [ ] Default values not validated

---

## 🔧 Validation (4)

- [ ] No return type validation
- [ ] No circular reference detection
- [ ] No struct tag validation
- [ ] No type checking for default values

---

## 🎯 Type System (4)

- [ ] Limited support for slices of structs
- [x] Pointers require manual handling
- [ ] Map types not supported
- [ ] Custom types registry limited

---

## 📦 Extractors (4)

- [ ] Form data extractor missing
- [ ] File upload not supported
- [ ] Response extractor is stub
- [ ] No content negotiation support

---

## 🎨 Handlers (4)

- [ ] ResponseWriter/Request handling inconsistent
- [ ] No streaming response support
- [ ] No middleware support
- [ ] Limited context enrichment

---

## ⚙️ Configuration (4)

- [ ] Body size limit hard-coded
- [ ] No global configuration
- [ ] Timeouts not configurable
- [ ] No template customization

---

## ❌ Error Handling (6)

- [x] Limited HTTP error status codes
- [ ] Generic error messages
- [ ] No error logging
- [ ] Error messages without context
- [x] Incomplete error wrapping
- [ ] No error context in template

---

## 📚 Code Quality (11)

### Simplification (3)
- [ ] Simplify case conversion
- [ ] Reduce duplication in slice parsing
- [ ] Reduce duplication in type registration

### Logging (2)
- [ ] Inconsistent logging
- [ ] Missing context propagation

### Constants (3)
- [ ] Magic string constants
- [ ] Magic numbers
- [ ] Hard-coded strings

---

## ⚡ Performance (8)

### Code Generation (6)
- [ ] Double AST traversal
- [ ] Inefficient extractor sorting
- [ ] Body reading in memory
- [ ] No caching mechanism
- [ ] Template re-parsing
- [ ] goimports is slow

### Runtime (2)
- [ ] JSON encoding without pooling
- [ ] String concatenation in loops

---

## 🏗️ Architecture (8)

### Design Patterns (4)
- [ ] Global registry pattern
- [ ] No plugin system
- [ ] Strong coupling to AST
- [ ] No interfaces for core types

### Separation of Concerns (4)
- [ ] Generator does too much
- [ ] CLI and generator coupled
- [ ] Mixed responsibilities
- [ ] No public API boundary

---

## 🔐 Security (28)

### Input Validation (7)
- [ ] No Content-Type validation
- [ ] Parameter names not sanitized
- [ ] Validation tags ignored
- [ ] No string length limits
- [ ] No path traversal protection
- [ ] No header sanitization
- [ ] JSON body without per-field size limit

### Code Injection (4)
- [ ] Default values not escaped (CRITICAL)
- [ ] Parameter names not validated
- [ ] Field names not sanitized in errors
- [ ] Template injection risk

### Resource Exhaustion (8)
- [ ] Handler processing without limit
- [ ] No nesting depth limit
- [ ] Slice allocation without limit
- [ ] No operation timeouts
- [ ] Large array allocation
- [ ] No rate limiting
- [ ] Circular reference risk
- [ ] No file size limit

### Information Disclosure (2)
- [ ] Detailed error messages
- [ ] Potential stack trace leak

### File System Security (4)
- [ ] World-readable files
- [ ] Output path not validated
- [ ] No atomic writes
- [ ] Overwrites without backup

### Type Safety (4)
- [ ] Numeric overflow risk
- [ ] No bounds checking
- [ ] Nil pointer risk
- [ ] Type safety loss

### Missing Security Features (3)
- [ ] No CSRF protection
- [ ] No authentication hooks
- [ ] No security headers

### Dependencies (2)
- [ ] Dependencies not audited
- [ ] x/tools version pinned

---

## 📚 Documentation (17)

### Missing Documentation (9)
- [ ] Main README.md
- [ ] CONTRIBUTING.md
- [ ] CHANGELOG.md
- [ ] API documentation
- [ ] Examples directory
- [ ] Package documentation
- [ ] No code examples
- [ ] Template syntax not documented
- [ ] No custom extractors guide

### Improve Existing Documentation (4)
- [ ] Template comments
- [ ] Parser comments
- [ ] Extractor documentation
- [ ] Minimal comments in generated code

### Learning Resources (4)
- [ ] No step-by-step tutorial
- [ ] No video demo
- [ ] No FAQ
- [ ] No troubleshooting guide

---

## 🔄 Technical Debt (23)

### Code Organization (6)
- [ ] Duplicated version constant
- [ ] Similar extractor structures
- [ ] Mixed responsibilities
- [ ] No public API boundary
- [ ] Duplication in extractors
- [ ] Long functions

### Naming (2)
- [ ] Inconsistent naming
- [ ] Non-obvious abbreviations

### Comments (2)
- [ ] Commented code
- [ ] TODO comments

### Testing Infrastructure (5)
- [ ] Global state makes testing difficult
- [ ] No mockable interfaces
- [ ] Template testing challenges
- [ ] No integration tests
- [ ] Template output not tested

### Maintenance (4)
- [ ] Hardcoded constants
- [ ] Fragile comment parsing
- [ ] No generated code versioning
- [ ] No deprecation strategy

### Dependencies (3)
- [ ] Heavy import dependency
- [ ] No dependency injection
- [ ] Strong coupling to AST

### Build & Release (1)
- [ ] No CI/CD configuration

---

## 🎨 Developer Experience (15)

### CLI (7)
- [ ] No `apikit init` command
- [ ] No `apikit validate` command
- [x] Limited verbose output
- [ ] No progress indicator
- [ ] No batch processing
- [ ] No quiet mode
- [ ] No structured output format

### Error Messages (6)
- [ ] Cryptic parser errors
- [ ] No line numbers in errors
- [ ] No fix suggestions
- [ ] Error messages without help
- [ ] Errors without context
- [ ] No internationalization

### IDE Integration (2)
- [ ] No language server
- [ ] No IDE snippets

---

## 🚀 New Features (12)

### OpenAPI/Swagger (2)
- [ ] Generate OpenAPI specs
- [ ] Swagger UI integration

### Client Generation (2)
- [ ] Generate Go clients
- [ ] Generate TypeScript types

### Monitoring (2)
- [ ] Metrics integration
- [ ] Tracing support

### Mock Generation (1)
- [ ] Generate mocks for testing

### Incomplete Features (5)
- [ ] Validate tag not used
- [ ] IsDTO flag not used
- [ ] PackagePath not used
- [x] Warnings not always reported
- [ ] No custom status codes support

---

## 📊 Metrics and Monitoring (5)

### Code Metrics (3)
- [ ] No code coverage report
- [ ] No complexity metrics
- [ ] No automatic linting

### Runtime Metrics (2)
- [ ] No performance benchmarks
- [ ] No memory profiling

---

## 🔧 DevOps and CI/CD (5)

### Missing (3)
- [ ] No GitHub Actions
- [ ] No Dockerfile
- [ ] No release process

### Tooling (2)
- [ ] No Makefile
- [ ] No pre-commit hooks

---

## 📦 Packaging and Distribution (3)

### Installation (2)
- [ ] No easy installation
- [x] No version check

### Module Management (1)
- [ ] Module path may change

---

## 🏆 Best Practices (6)

### Go Best Practices (3)
- [ ] Inconsistent error wrapping
- [ ] Context usage
- [ ] Defer usage

### HTTP Best Practices (3)
- [ ] No CORS headers
- [ ] No request ID tracking
- [ ] No rate limiting

---

**Total:** 188 pending tasks (6 completed)

