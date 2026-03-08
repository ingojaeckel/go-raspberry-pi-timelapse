---
name: deep-code-review 
description: Performs a detailed code review to catch regressions of recent changes, long standing bugs, and identify a prioritised list of things to improve
---

Perform a deep system design and code review of this project. Pay special attention to:

1. robustness and reliality; anything that could impact long term operation such as resource leaks and race conditions
2. areas where code differs from code such as mismatches with code comments, markdown files, or other project documentation
3. validate automated tests are passing; assess coverage - which issues are not validated due to the lack of tests?
4. are expectations described sufficiently? is there ambiguity in how the system should behave in certain situations? does this ambiguity expose the system to potential bugs, missing code, missing tests?
5. are DRY principles applied?
6. are coding standards applied uniformly across the codebase? is code written consistently across different subsystems?
7. does claude.md or other files need to be updated with context which has changed recently?

