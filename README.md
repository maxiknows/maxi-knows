# Maxi Knows

Maxi Knows is a DevOps project developed as part of the Datamatiker education at KEA.

The project is based on a legacy version of the WhoKnows application and focuses on modernising, securing and deploying an existing application rather than building a new system from scratch.

The original application was written in Python and contains several intentionally outdated or problematic implementations. As part of the project, we are gradually rewriting the application in Go while improving the project structure, security, maintainability and deployment process.

## Project goals

The main focus of the project is to work with real-world DevOps and legacy modernisation challenges, including:

- Migrating legacy Python functionality to Go
- Refactoring the project structure
- Improving password security and authentication
- Working with SQLite and database persistence
- Code quality analysis with SonarQube Cloud and Qlty
- Git and GitHub workflows using branches and pull requests
- Deployment to a Linux VM
- Automated database backups
- Maintaining and improving an existing codebase as a team

## Technologies

- Go
- Python (legacy application)
- SQLite
- HTML / CSS
- Git & GitHub
- Linux
- SonarQube Cloud
- Qlty

## Code quality

**SonarQube Cloud**

[![Quality gate](https://sonarcloud.io/api/project_badges/quality_gate?project=maxiknows_maxi-knows)](https://sonarcloud.io/summary/new_code?id=maxiknows_maxi-knows)

[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=maxiknows_maxi-knows&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=maxiknows_maxi-knows)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=maxiknows_maxi-knows&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=maxiknows_maxi-knows)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=maxiknows_maxi-knows&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=maxiknows_maxi-knows)

**Qlty**

[![Maintainability](https://qlty.sh/gh/maxiknows/projects/maxi-knows/maintainability.svg)](https://qlty.sh/gh/maxiknows/projects/maxi-knows)

## Project structure

The repository contains both the original legacy implementation and the newer Go implementation.

```text
go/
├── cmd/
│   ├── maxi-knows/
│   └── init-db/
├── internal/
│   └── storage/
├── templates/
└── static/

legacy-python/
