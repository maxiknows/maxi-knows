# Playwright End-to-End Testing

## Purpose

We use Playwright to automate end-to-end (E2E) testing of the Maxi Knows application.

E2E testing checks complete user flows from the user's perspective, rather than testing individual functions in isolation.

Playwright interacts with the application through a real browser. This allows us to test whether the frontend, Go backend and SQLite database work together as expected.

The tests help us detect regressions when changes are introduced and provide additional confidence before merging pull requests.

## Test Structure

Playwright is installed separately from the Go application.

```text
go/tests/e2e/
├── package.json
├── package-lock.json
├── playwright.config.js
└── tests/
    ├── homepage.spec.js
    ├── search.spec.js
    └── auth.spec.js
```

The tests are written in JavaScript using `@playwright/test`.

We currently use Chromium as our test browser.

## Implemented Tests

### Homepage – `homepage.spec.js`

Checks that:

- The homepage loads successfully.
- The page title is correct.
- The search input is visible.
- The search button is visible.

### Search – `search.spec.js`

Simulates a user searching for a query.

Checks that:

- A search query can be entered.
- The search button can be clicked.
- The URL updates with the query.
- The search input retains the entered value.

This test verifies the search interaction but does not currently assert that matching database results are returned.

### Authentication – `auth.spec.js`

Tests a complete authentication flow:

1. Register a new user with a unique username.
2. Verify redirection to the login page.
3. Log in with the new credentials.
4. Verify that the username is displayed in the navigation.
5. Log out.
6. Verify that the login link is visible again.

This test involves the browser, Go authentication handlers, session handling and SQLite database.

## Test Environment

Playwright is configured to test the Go application at:

`http://localhost:8080`

Authentication tests create users in the database. To avoid modifying production data, we use a separate SQLite test database.

The database schema is initialized using the existing `scripts/smoke-schema.sql` file.

To start the Go application with the local test database, run from the `go/` directory:

```powershell
$env:DB_PATH = "tests/e2e/test.db"
go run ./cmd/maxi-knows
```

The test database is excluded from Git.

## Running Tests Locally

From `go/tests/e2e/`, install dependencies:

```bash
npm ci
npx playwright install chromium
```

Run all tests:

```bash
npx playwright test --project=chromium
```

To see the browser during execution:

```bash
npx playwright test --headed --project=chromium
```

To use Playwright's interactive test interface:

```bash
npx playwright test --ui
```

The Go application must be running before executing the tests.

## GitHub Actions

A separate workflow has been added:

`.github/workflows/playwright.yml`

It is configured to run on pull requests targeting `main` and pushes to `main`.

The workflow:

1. Checks out the repository.
2. Installs Go and Node.js.
3. Installs Playwright dependencies and Chromium.
4. Creates a temporary SQLite database.
5. Starts the Go application using the temporary database.
6. Runs the Playwright E2E tests.
7. Uploads the Playwright HTML report as an artifact.

The workflow uses an isolated database on the GitHub Actions runner. It does not modify the production database.

## Security Considerations

SonarQube identified security concerns related to npm lifecycle scripts and the use of `npx` in the workflow.

We updated the workflow to use:

- `npm ci --ignore-scripts` to prevent npm dependency lifecycle scripts from running during installation.
- The locally installed Playwright executable instead of `npx` in the GitHub Actions commands.

The project also uses `package-lock.json` to lock dependency versions.

## Current Results

All three Playwright tests have passed locally using Chromium.

The GitHub Actions workflow has been added and the security-related installation commands have been updated. The final workflow result should be verified in the Pull Request checks.

## Limitations and Future Improvements

- Only Chromium is currently tested.
- Tests cover selected user flows rather than every possible scenario.
- The search test does not currently verify the returned search results.
- Authentication tests create users in the test database, which is recreated for each GitHub Actions run.
- Additional tests could cover invalid input, duplicate registrations, incorrect passwords and session expiration.
- Tests could later be extended to other browsers.
- Automated E2E testing complements unit tests, integration tests and the existing smoke tests rather than replacing them.

## References

- [Playwright documentation](https://playwright.dev/docs/intro)
- [Anders' E2E testing example](https://github.com/who-knows-inc/whoknows_variations/tree/end-to-end_testing)
