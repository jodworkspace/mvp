# MVP

MVP Backend for my boring project, to keep myself learning Go and Clean Architecture.

- Enable user authentication through Google login.
- Markdown support for notes and documentation.
- Task management with a To-do list.
- Integrate the Google Drive for storage & YouTube API for music/video streaming functionality.

### Dependency Rules

| Package    | Rule                                                                                        |
|------------|---------------------------------------------------------------------------------------------|
| domain     | No dependencies. Other layers can depend on it                                              |
| repository | Depends on domain and external libraries                                                    |
| usecase    | Depends only on domain and repository interfaces                                            |
| middleware | Can depend on domain and sometimes usecase (if needed)                                      |
| handler    | Depends on usecase, domain, and middleware                                                  |
| pkg        | Shared utilities, reusable across all layers. Should not import any other internal packages |
| config     | Can import external libraries or standard Go packages                                       |

- If shared types or functions need to be used in the domain package but must remain independent of the rest of the application, they should be placed in a separate project or module.

## References

- [Clean Architecture](https://medium.com/@rayato159/how-to-implement-clean-architecture-in-golang-en-f50d66378ebf)
- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)

### Naming Conventions

- [Acronyms - Consistent Case](https://go.dev/wiki/CodeReviewComments#initialisms)