```markdown
# pounce Development Patterns

> Auto-generated skill from repository analysis

## Overview
This skill teaches the core development patterns and conventions used in the `pounce` Go codebase. You'll learn about file naming, import/export styles, commit message conventions, and how to write and organize tests. While no specific automation workflows were detected, this guide provides recommended commands and best practices for efficient development.

## Coding Conventions

### File Naming
- Use **camelCase** for all file names.
  - Example: `myModule.go`, `userService.go`

### Import Style
- Use **relative imports** within the codebase.
  - Example:
    ```go
    import "./utils"
    ```

### Export Style
- Use **default exports** (in Go, this means exporting identifiers that start with an uppercase letter).
  - Example:
    ```go
    // Exported function
    func MyFunction() {}

    // Exported struct
    type User struct {
        Name string
    }
    ```

### Commit Messages
- Follow the **conventional commit** format.
- Use the `fix` prefix for bug fixes.
- Keep commit messages concise (average ~47 characters).
  - Example:
    ```
    fix: handle nil pointer in userService
    ```

## Workflows

### Committing Changes
**Trigger:** When you have made code changes and are ready to commit.
**Command:** `/commit-fix`

1. Stage your changes:
    ```
    git add .
    ```
2. Write a conventional commit message using the `fix` prefix:
    ```
    git commit -m "fix: brief description of the fix"
    ```
3. Push your changes:
    ```
    git push
    ```

### Importing Modules
**Trigger:** When you need to use code from another file or package in the repository.
**Command:** `/import-module`

1. Use a relative import in your Go file:
    ```go
    import "./moduleName"
    ```

### Exporting Functions or Types
**Trigger:** When you want to make a function, type, or variable accessible from other packages.
**Command:** `/export-symbol`

1. Name the exported item with an uppercase first letter:
    ```go
    func ExportedFunction() {}
    type ExportedType struct {}
    ```

## Testing Patterns

- Test files follow the `*.test.*` pattern (e.g., `userService.test.go`).
- The specific testing framework is unknown, but standard Go testing practices likely apply.
- Example test file structure:
    ```go
    package mypackage

    import "testing"

    func TestMyFunction(t *testing.T) {
        result := MyFunction()
        if result != expected {
            t.Errorf("expected %v, got %v", expected, result)
        }
    }
    ```

## Commands
| Command         | Purpose                                    |
|-----------------|--------------------------------------------|
| /commit-fix     | Commit code changes with a fix prefix      |
| /import-module  | Import a module using relative import      |
| /export-symbol  | Export a function or type (uppercase name) |
```
